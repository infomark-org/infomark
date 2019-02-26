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

package app

import (
  olog "log"
  "net/http"
  "net/http/httptest"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  otape "github.com/cgtuebingen/infomark-backend/tape"
  homedir "github.com/mitchellh/go-homedir"
  "github.com/spf13/viper"
)

// just wrap tape

func addJWTClaims(r *http.Request, loginID int64, root bool) {
  accessToken, err := tokenManager.CreateAccessJWT(authenticate.NewAccessClaims(loginID, root))
  if err != nil {
    panic(err)
  }
  // we use JWT here
  r.Header.Add("Authorization", "Bearer "+accessToken)
}

var tokenManager *authenticate.TokenAuth

func SetConfigFile() {

  // Find home directory.
  home, err := homedir.Dir()
  if err != nil {
    olog.Fatal(err)
  }

  viper.AddConfigPath(home)
  viper.SetConfigName(".infomark-backend")
}

func InitConfig() {
  SetConfigFile()
  viper.AutomaticEnv()

  // If a config file is found, read it in.
  if err := viper.ReadInConfig(); err != nil {
    panic(err)
  }
}

func init() {
  // prepate token management
  tokenManager, _ = authenticate.NewTokenAuth()
  InitConfig()

}

type Tape struct {
  otape.Tape
}

func NewTape() *Tape {
  t := &Tape{}
  return t
}

func (t *Tape) AfterEach() {
  t.DB.Close()
}

func (t *Tape) BeforeEach() {
  var err error

  t.DB, err = helper.TransactionDB()
  if err != nil {
    panic(err)
  }

  t.Router, _ = New(t.DB, false)

}

// PlayWithClaims will send a request without any request body (like GET) but with JWT bearer.
func (t *Tape) PlayWithClaims(method, url string, loginID int64, root bool) *httptest.ResponseRecorder {

  h := make(map[string]interface{})
  r := otape.BuildDataRequest(method, url, h)
  addJWTClaims(r, loginID, root)
  return t.PlayRequest(r)
}

// PlayDataWithClaims will send a request with given data in body and add JWT bearer.
func (t *Tape) PlayDataWithClaims(method, url string, data map[string]interface{}, loginID int64, root bool) *httptest.ResponseRecorder {
  r := otape.BuildDataRequest(method, url, data)
  addJWTClaims(r, loginID, root)
  return t.PlayRequest(r)
}

func (t *Tape) PlayRequestWithClaims(r *http.Request, loginID int64, root bool) *httptest.ResponseRecorder {
  w := httptest.NewRecorder()
  addJWTClaims(r, loginID, root)
  t.Router.ServeHTTP(w, r)
  return w
}
