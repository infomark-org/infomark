// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package authenticate

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/infomark-org/infomark-backend/auth"
	"github.com/infomark-org/infomark-backend/symbol"
	"github.com/spf13/viper"
	"github.com/ulule/limiter/v3"

	// "github.com/ulule/limiter/v3/drivers/store/memory"
	redis "github.com/go-redis/redis"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

// RequiredValidAccessClaimsMiddleware tries to get information about the identity which
// issues a request by looking into the authorization header and then into
// the cookie.
func RequiredValidAccessClaims(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accessClaims := &AccessClaims{}

		if !viper.GetBool("debug_user_enabled") {
			// first we test the JWT autorization
			if HasHeaderToken(r) {

				// parse token from from header
				tokenStr := jwtauth.TokenFromHeader(r)

				// ok, there is a access token in the header
				err := accessClaims.ParseAccessClaimsFromToken(tokenStr)
				if err != nil {
					// fmt.Println(err)
					render.Render(w, r, auth.ErrUnauthorized)
					return
				}

			} else {
				// fmt.Println("no token, try session")
				if HasSessionToken(r) {
					// fmt.Println("found session")

					// session data is stored in cookie
					err := accessClaims.ParseRefreshClaimsFromSession(r)
					if err != nil {
						// fmt.Println(err)
						render.Render(w, r, auth.ErrUnauthorized)
						return
					}

					// session is valid --> we will extend the session
					w = accessClaims.UpdateSession(w, r)
				} else {
					// fmt.Println("NO session found")

					render.Render(w, r, auth.ErrUnauthenticated)
					return

				}

			}
		} else {
			accessClaims.LoginID = viper.GetInt64("debug_user_id")
			accessClaims.Root = viper.GetBool("debug_user_is_root")
		}

		// nothing given
		// serve next
		ctx := context.WithValue(r.Context(), symbol.CtxKeyAccessClaims, accessClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type LoginLimiterKey interface {
	Key() string
}

type LoginLimiter struct {
	Store  *limiter.Store
	Rate   *limiter.Rate
	Prefix string
	Redis  *redis.Client
}

type LoginLimiterKeyFromIP struct {
	R *http.Request
}

func NewLoginLimiterKeyFromIP(r *http.Request) *LoginLimiterKeyFromIP {
	return &LoginLimiterKeyFromIP{R: r}
}

func (obj *LoginLimiterKeyFromIP) Key() string {

	options := limiter.Options{
		IPv4Mask:           limiter.DefaultIPv4Mask,
		IPv6Mask:           limiter.DefaultIPv6Mask,
		TrustForwardHeader: true,
	}

	return limiter.GetIP(obj.R, options).String()
}

func NewLoginLimiter(prefix string, limit string, redisURL string) (*LoginLimiter, error) {
	// Define a limit rate to 4 requests per hour.
	rate, err := limiter.NewRateFromFormatted(limit)
	if err != nil {
		return nil, err
	}

	// Create a redis client.
	option, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(option)

	// Create a store with the redis client.
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   prefix,
		MaxRetry: 3,
	})

	if err != nil {
		return nil, err
	}

	return &LoginLimiter{Store: &store, Rate: &rate, Prefix: prefix, Redis: client}, nil
}

func (ll *LoginLimiter) Get(r *http.Request, KeyFunc LoginLimiterKey) (limiter.Context, error) {

	return limiter.Store.Get(
		*ll.Store,
		r.Context(),
		fmt.Sprintf("%s-%s", KeyFunc.Key(), ll.Prefix),
		*ll.Rate,
	)

}

func (ll *LoginLimiter) WriteHeaders(w http.ResponseWriter, context limiter.Context) {
	w.Header().Add("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
	w.Header().Add("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
	w.Header().Add("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))
}

// func RateLimitMiddleware(prefix string, limit string, redisURL string) func(h http.Handler) http.Handler {
func RateLimitMiddleware(ll *LoginLimiter) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			keyFunc := NewLoginLimiterKeyFromIP(r)

			context, err := ll.Get(r, keyFunc)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ll.WriteHeaders(w, context)
			if context.Reached {
				http.Error(w, "Limit exceeded", http.StatusTooManyRequests)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
