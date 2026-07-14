// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019 ComputerGraphics Tuebingen
//               2020-present InfoMark.org
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
	"testing"
	"time"

	"github.com/franela/goblin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/infomark-org/infomark/configuration"
)

// testJWTSecret is the shared secret used to sign and verify tokens in these
// tests. It is arbitrary because the tests only exercise the round-trip.
const testJWTSecret = "unit-test-secret-do-not-use-in-production"

// testLoginID is a representative non-zero account identifier carried in claims.
const testLoginID int64 = 42

// oneYearInSeconds is an upper bound used both to craft clearly-expired tokens
// and to assert that a tiny configured expiry stays far below a year, proving
// the expiry is computed as a real duration rather than a nanosecond count.
const oneYearInSeconds int64 = 365 * 24 * 60 * 60

// newTestTokenAuth builds a TokenAuth from an in-memory configuration so the
// tests do not depend on any config file. The access expiry is deliberately
// set to one second to expose how CreateAccessJWT treats the duration.
func newTestTokenAuth() *TokenAuth {
	config := &configuration.AuthenticationConfiguration{}
	config.JWT.Secret = testJWTSecret
	config.JWT.AccessExpiry = time.Second
	config.JWT.RefreshExpiry = time.Second
	return NewTokenAuth(config)
}

func TestTokenAuth(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Access token round-trip", func() {

		// The core invariant: claims encoded into an access token must be
		// recoverable unchanged after signing and verifying.
		g.It("Should round-trip access claims through create and parse", func() {
			tokenAuth := newTestTokenAuth()
			claims := NewAccessClaims(testLoginID, true)

			tokenString, err := tokenAuth.CreateAccessJWT(claims)
			g.Assert(err).Equal(nil)
			g.Assert(tokenString == "").IsFalse()

			parsed := &AccessClaims{}
			err = parsed.ParseAccessClaimsFromToken(testJWTSecret, tokenString)
			g.Assert(err).Equal(nil)
			g.Assert(parsed.LoginID).Equal(testLoginID)
			g.Assert(parsed.Root).IsTrue()
			g.Assert(parsed.AccessNotRefresh).IsTrue()
		})

		// A token signed with one secret must not verify under a different secret.
		g.It("Should reject an access token verified with the wrong secret", func() {
			tokenAuth := newTestTokenAuth()
			tokenString, err := tokenAuth.CreateAccessJWT(NewAccessClaims(testLoginID, false))
			g.Assert(err).Equal(nil)

			parsed := &AccessClaims{}
			err = parsed.ParseAccessClaimsFromToken("a-completely-different-secret", tokenString)
			g.Assert(err == nil).IsFalse()
		})

		// Arbitrary non-token bytes must be rejected rather than accepted or
		// causing a panic.
		g.It("Should reject a garbage token string", func() {
			parsed := &AccessClaims{}
			err := parsed.ParseAccessClaimsFromToken(testJWTSecret, "this.is.not-a-jwt")
			g.Assert(err == nil).IsFalse()
		})

		// A refresh token must not be accepted where an access token is required,
		// even though its signature is valid.
		g.It("Should reject a refresh token when parsing access claims", func() {
			tokenAuth := newTestTokenAuth()
			refreshToken, err := tokenAuth.CreateRefreshJWT(NewRefreshClaims(testLoginID))
			g.Assert(err).Equal(nil)

			parsed := &AccessClaims{}
			err = parsed.ParseAccessClaimsFromToken(testJWTSecret, refreshToken)
			g.Assert(err == nil).IsFalse()
		})

		// An expired access token must be rejected. We craft the claims directly
		// with an exp in the past and sign them through the same JwtAuth encoder
		// the production code uses, so this exercises the real verification path.
		g.It("Should reject an expired access token", func() {
			tokenAuth := newTestTokenAuth()
			claims := NewAccessClaims(testLoginID, false)
			// jwt/v5 models the standard iat/exp claims as NumericDate values
			// instead of raw int64 seconds. Wrapping whole-second unix
			// timestamps preserves the exact numeric claims this test crafts.
			claims.RegisteredClaims.IssuedAt = jwt.NewNumericDate(time.Unix(time.Now().UTC().Unix()-2*oneYearInSeconds, 0))
			claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Unix(time.Now().UTC().Unix()-oneYearInSeconds, 0))

			_, tokenString, err := tokenAuth.JwtAuth.Encode(claims.ToMap())
			g.Assert(err).Equal(nil)

			parsed := &AccessClaims{}
			err = parsed.ParseAccessClaimsFromToken(testJWTSecret, tokenString)
			g.Assert(err == nil).IsFalse()
		})
	})

	g.Describe("Refresh token round-trip", func() {

		// Refresh claims must round-trip and the parser must confirm the token is
		// a refresh token (AccessNotRefresh == false).
		g.It("Should round-trip refresh claims through create and parse", func() {
			tokenAuth := newTestTokenAuth()
			refreshToken, err := tokenAuth.CreateRefreshJWT(NewRefreshClaims(testLoginID))
			g.Assert(err).Equal(nil)

			parsed := &RefreshClaims{}
			err = parsed.ParseRefreshClaimsFromToken(testJWTSecret, refreshToken)
			g.Assert(err).Equal(nil)
			g.Assert(parsed.LoginID).Equal(testLoginID)
			g.Assert(parsed.AccessNotRefresh).IsFalse()
		})

		// An access token must not be accepted where a refresh token is required.
		g.It("Should reject an access token when parsing refresh claims", func() {
			tokenAuth := newTestTokenAuth()
			accessToken, err := tokenAuth.CreateAccessJWT(NewAccessClaims(testLoginID, false))
			g.Assert(err).Equal(nil)

			parsed := &RefreshClaims{}
			err = parsed.ParseRefreshClaimsFromToken(testJWTSecret, accessToken)
			g.Assert(err == nil).IsFalse()
		})
	})

	g.Describe("Access token expiry unit", func() {

		// CreateAccessJWT must treat the configured expiry as a real duration:
		// a one-second configured expiry produces a token whose exp claim is
		// about one second after issuance, not decades away. This pins the fix
		// for the nanoseconds-added-to-seconds bug documented in SESSION.md.
		g.It("Should expire one second after issuance for a 1s config", func() {
			tokenAuth := newTestTokenAuth()
			before := time.Now().UTC().Unix()
			tokenString, err := tokenAuth.CreateAccessJWT(NewAccessClaims(testLoginID, false))
			g.Assert(err).Equal(nil)
			after := time.Now().UTC().Unix()

			// Decode the token independently to inspect the raw exp claim.
			decoded := &AccessClaims{}
			_, parseErr := jwt.ParseWithClaims(tokenString, decoded, func(token *jwt.Token) (interface{}, error) {
				return []byte(testJWTSecret), nil
			})
			g.Assert(parseErr).Equal(nil)

			// The 1s configured expiry must land within one second of issuance.
			// It must be at least the issuance time and at most one second past
			// the latest possible issuance, and nowhere near a year away.
			expiresAt := decoded.RegisteredClaims.ExpiresAt.Unix()
			g.Assert(expiresAt >= before).IsTrue()
			g.Assert(expiresAt <= after+1).IsTrue()
			g.Assert(expiresAt < before+oneYearInSeconds).IsTrue()
		})
	})
}
