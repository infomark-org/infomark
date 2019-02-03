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
  "errors"

  jwt "github.com/dgrijalva/jwt-go"
)

// AccessClaims represent the claims parsed from JWT access token.
type AccessClaims struct {
  LoginID int  // the id to get user information
  Root    bool // a global flag to bypass all permission checks
}

// ParseClaims parses JWT claims into AccessClaims.
func (c *AccessClaims) ParseClaims(claims jwt.MapClaims) error {
  // loginID represents the userID of the identity who is doing the request.
  loginID, ok := claims["login_id"]
  if !ok {
    return errors.New("could not parse claim login_id")
  }
  c.LoginID = int(loginID.(float64))

  return nil
}

// RefreshClaims represent the claims parsed from JWT refresh token.
type RefreshClaims struct {
  Token string
}

// ParseClaims parses the JWT claims into RefreshClaims.
func (c *RefreshClaims) ParseClaims(claims jwt.MapClaims) error {
  token, ok := claims["token"]
  if !ok {
    return errors.New("could not parse claim token")
  }
  c.Token = token.(string)
  return nil
}
