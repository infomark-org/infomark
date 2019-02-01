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
  "time"

  jwt "github.com/dgrijalva/jwt-go"
  "github.com/go-chi/jwtauth"
  "github.com/spf13/viper"
)

// TokenAuth implements JWT authentication flow.
type TokenAuth struct {
  JwtAuth          *jwtauth.JWTAuth
  JwtAccessExpiry  time.Duration
  JwtRefreshExpiry time.Duration
}

// NewTokenAuth configures and returns a JWT authentication instance.
func NewTokenAuth() (*TokenAuth, error) {
  secret := viper.GetString("auth_jwt_secret")

  a := &TokenAuth{
    JwtAuth:          jwtauth.New("HS256", []byte(secret), nil),
    JwtAccessExpiry:  viper.GetDuration("auth_jwt_access_expiry"),
    JwtRefreshExpiry: viper.GetDuration("auth_jwt_refresh_expiry"),
  }

  return a, nil
}

// Verifier http middleware will verify a jwt string from a http request.
func (a *TokenAuth) Verifier() func(http.Handler) http.Handler {
  return jwtauth.Verifier(a.JwtAuth)
}

// GenTokenPair returns both an access token and a refresh token.
func (a *TokenAuth) GenTokenPair(ca jwt.MapClaims, cr jwt.MapClaims) (string, string, error) {
  access, err := a.CreateAccessJWT(ca)
  if err != nil {
    return "", "", err
  }
  refresh, err := a.CreateRefreshJWT(cr)
  if err != nil {
    return "", "", err
  }
  return access, refresh, nil
}

// CreateAccessJWT returns an access token for provided account claims.
func (a *TokenAuth) CreateAccessJWT(c jwt.MapClaims) (string, error) {
  jwtauth.SetIssuedNow(c)
  jwtauth.SetExpiryIn(c, a.JwtAccessExpiry)
  _, tokenString, err := a.JwtAuth.Encode(c)
  return tokenString, err
}

// CreateRefreshJWT returns a refresh token for provided token Claims.
func (a *TokenAuth) CreateRefreshJWT(c jwt.MapClaims) (string, error) {
  jwtauth.SetIssuedNow(c)
  jwtauth.SetExpiryIn(c, a.JwtRefreshExpiry)
  _, tokenString, err := a.JwtAuth.Encode(c)
  return tokenString, err
}
