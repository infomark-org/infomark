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
	"net/http"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/infomark-org/infomark-backend/configuration"
)

// TokenAuth implements JWT authentication flow.
type TokenAuth struct {
	JwtAuth          *jwtauth.JWTAuth
	JwtAccessExpiry  time.Duration
	JwtRefreshExpiry time.Duration
}

// NewTokenAuth configures and returns a JWT authentication instance.
func NewTokenAuth(config *configuration.AuthenticationConfiguration) *TokenAuth {
	return &TokenAuth{
		JwtAuth:          jwtauth.New("HS256", []byte(config.JWT.Secret), nil),
		JwtAccessExpiry:  config.JWT.AccessExpiry,
		JwtRefreshExpiry: config.JWT.RefreshExpiry,
	}

}

// Verifier http middleware will verify a jwt string from a http request.
func (a *TokenAuth) Verifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(a.JwtAuth)
}

// CreateAccessJWT returns an access token for provided account claims.
func (a *TokenAuth) CreateAccessJWT(claims AccessClaims) (string, error) {
	claims.StandardClaims.IssuedAt = time.Now().UTC().Unix()
	claims.StandardClaims.ExpiresAt = time.Now().UTC().Unix() + int64(a.JwtAccessExpiry)

	_, tokenString, err := a.JwtAuth.Encode(claims)
	return tokenString, err
}

// CreateRefreshJWT returns a refresh token for provided token Claims.
func (a *TokenAuth) CreateRefreshJWT(claims RefreshClaims) (string, error) {

	claims.StandardClaims.IssuedAt = time.Now().UTC().Unix()
	claims.StandardClaims.ExpiresAt = time.Now().UTC().Unix() + int64(a.JwtRefreshExpiry)

	_, tokenString, err := a.JwtAuth.Encode(claims)
	return tokenString, err
}
