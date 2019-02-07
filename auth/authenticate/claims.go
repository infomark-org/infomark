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
  "fmt"
  "net/http"

  jwt "github.com/dgrijalva/jwt-go"
  "github.com/spf13/viper"
)

// AccessClaims represent the claims parsed from JWT access token.
type AccessClaims struct {
  jwt.StandardClaims
  AccessNotRefresh bool  `json:"anr"`      // to distinguish between access and refresh code
  LoginID          int64 `json:"login_id"` // the id to get user information
  Root             bool  `json:"root"`     // a global flag to bypass all permission checks
}

func NewAccessClaims(loginId int64, root bool) AccessClaims {
  return AccessClaims{
    LoginID:          loginId,
    AccessNotRefresh: true,
    Root:             root,
  }
}

// RefreshClaims represent the claims parsed from JWT refresh token.
type RefreshClaims struct {
  jwt.StandardClaims
  AccessNotRefresh bool  `json:"anr"`
  LoginID          int64 `json:"login_id"`
}

func NewRefreshClaims(loginId int64) RefreshClaims {
  return RefreshClaims{
    LoginID:          loginId,
    AccessNotRefresh: false,
  }
}

// Parse refresh claims from a token string
func (ret *RefreshClaims) ParseRefreshClaimsFromToken(tokenStr string) error {

  secret := viper.GetString("auth_jwt_secret")

  // verify the token
  token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
    return []byte(secret), nil
  })

  if err != nil {
    return err
  }

  if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {

    if !claims.AccessNotRefresh {
      ret.LoginID = claims.LoginID
      ret.AccessNotRefresh = claims.AccessNotRefresh
      return nil
    } else {
      return errors.New("token is an access token, but refresh token was required")
    }

  } else {
    return errors.New("token is invalid")
  }

}

// Parse access claims from a JWT token string
func (ret *AccessClaims) ParseAccessClaimsFromToken(tokenStr string) error {

  secret := viper.GetString("auth_jwt_secret")

  // verify the token
  token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
    return []byte(secret), nil
  })

  if err != nil {
    return err
  }

  if claims, ok := token.Claims.(*AccessClaims); ok && token.Valid {

    if !claims.AccessNotRefresh {
      ret.LoginID = claims.LoginID
      ret.AccessNotRefresh = claims.AccessNotRefresh
      ret.Root = claims.Root
      return nil
    } else {
      return errors.New("token is an refresh token, but access token was required")
    }

  } else {
    return errors.New("token is invalid")
  }

}

// Parse access claims from a cookie
func (ret *AccessClaims) ParseRefreshClaimsFromSession(r *http.Request) error {

  session := SessionManager.Load(r)

  loginId, err := session.GetInt64("login_id")
  if err != nil {
    return err
  }
  root, err := session.GetBool("root")
  if err != nil {
    return err
  }

  ret.LoginID = loginId
  // cookie based authentification is access-token only
  ret.AccessNotRefresh = true
  ret.Root = root
  return nil
}

func (ret *AccessClaims) WriteToSession(w http.ResponseWriter, r *http.Request) http.ResponseWriter {
  session := SessionManager.Load(r)

  err := session.PutInt64(w, "login_id", ret.LoginID)
  if err != nil {
    panic("hh")
  }
  fmt.Println("Wrote ret.LoginID", ret.LoginID)
  err = session.PutBool(w, "root", ret.Root)
  if err != nil {
    panic("hh")
  }
  fmt.Println("Wrote ret.Root", ret.Root)

  return w
}
