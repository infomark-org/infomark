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
  "net/http"

  "github.com/go-chi/jwtauth"
)

// type ctxKey int

// const (
//   ctxAccessClaims ctxKey = iota
//   ctxRefreshToken
// )

// // ClaimsFromCtx retrieves the parsed AccessClaims from request context.
// func ClaimsFromCtx(ctx context.Context) AccessClaims {
//   return ctx.Value(ctxAccessClaims).(AccessClaims)
// }

// // RefreshTokenFromCtx retrieves the parsed refresh token from context.
// func RefreshTokenFromCtx(ctx context.Context) string {
//   return ctx.Value(ctxRefreshToken).(string)
// }

// // AuthenticateAccessJWT is a default authentication middleware to enforce access from the
// // Verifier middleware request context values. The AuthenticateAccessJWT sends a 401 Unauthorized
// // response for any unverified tokens and passes the good ones through.
// func AuthenticateAccessJWT(next http.Handler) http.Handler {
//   return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//     token, claims, err := jwtauth.FromContext(r.Context())

//     if err != nil {
//       logging.GetLogEntry(r).Warn(err)
//       render.Render(w, r, auth.ErrUnauthorized(auth.ErrTokenUnauthorized))
//       return
//     }

//     if !token.Valid {
//       render.Render(w, r, auth.ErrUnauthorized(auth.ErrTokenExpired))
//       return
//     }

//     // AccessToken is authenticated, parse claims
//     var c AccessClaims
//     err = c.ParseClaims(claims)
//     if err != nil {
//       logging.GetLogEntry(r).Error(err)
//       render.Render(w, r, auth.ErrUnauthorized(auth.ErrInvalidAccessToken))
//       return
//     }

//     // Set AccessClaims on context
//     ctx := context.WithValue(r.Context(), ctxAccessClaims, c)
//     next.ServeHTTP(w, r.WithContext(ctx))
//   })
// }

func HasHeaderToken(r *http.Request) bool {
  jwt := jwtauth.TokenFromHeader(r)
  return jwt != ""
}

// // AuthenticateRefreshJWT checks validity of refresh tokens and is
// // only used for access token refresh and logout requests.
// // It responds with 401 Unauthorized for invalid or expired refresh tokens.
// func AuthenticateRefreshJWT(next http.Handler) http.Handler {
//   return http.HandlerFunc(
//     func(w http.ResponseWriter, r *http.Request) {
//       token, claims, err := jwtauth.FromContext(r.Context())
//       if err != nil {
//         logging.GetLogEntry(r).Warn(err)
//         render.Render(w, r, auth.ErrUnauthorized(auth.ErrTokenUnauthorized))
//         return
//       }
//       if !token.Valid {
//         render.Render(w, r, auth.ErrUnauthorized(auth.ErrTokenExpired))
//         return
//       }

//       // Token is authenticated, parse refresh token string
//       var c RefreshClaims
//       err = c.ParseClaims(claims)
//       if err != nil {
//         logging.GetLogEntry(r).Error(err)
//         render.Render(w, r, auth.ErrUnauthorized(auth.ErrInvalidRefreshToken))
//         return
//       }

//       // Set refresh token string on context
//       ctx := context.WithValue(r.Context(), ctxRefreshToken, "c.Token")
//       next.ServeHTTP(w, r.WithContext(ctx))
//     })
// }
