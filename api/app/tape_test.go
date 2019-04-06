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
  "encoding/json"
  olog "log"
  "net/http"
  "net/http/httptest"
  "os"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  otape "github.com/cgtuebingen/infomark-backend/tape"
  "github.com/jmoiron/sqlx"
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
  var err error
  home := os.Getenv("INFOMARK_CONFIG_DIR")

  if home == "" {
    // Find home directory.
    home, err = os.Getwd()
    if err != nil {
      olog.Fatal(err)
    }
  }

  viper.AddConfigPath(home)
  viper.SetConfigName(".infomark")
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

  DB *sqlx.DB
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

func (t *Tape) GetWithClaims(url string, loginID int64, root bool) *httptest.ResponseRecorder {
  h := make(map[string]interface{})
  r := otape.BuildDataRequest("GET", url, h)
  addJWTClaims(r, loginID, root)
  return t.PlayRequest(r)
}

func (t *Tape) PostWithClaims(url string, data map[string]interface{}, loginID int64, root bool) *httptest.ResponseRecorder {
  r := otape.BuildDataRequest("POST", url, data)
  addJWTClaims(r, loginID, root)
  return t.PlayRequest(r)
}

func (t *Tape) PutWithClaims(url string, data map[string]interface{}, loginID int64, root bool) *httptest.ResponseRecorder {
  r := otape.BuildDataRequest("PUT", url, data)
  addJWTClaims(r, loginID, root)
  return t.PlayRequest(r)
}

func (t *Tape) PatchWithClaims(url string, data map[string]interface{}, loginID int64, root bool) *httptest.ResponseRecorder {
  r := otape.BuildDataRequest("PATCH", url, data)
  addJWTClaims(r, loginID, root)
  return t.PlayRequest(r)
}

func (t *Tape) DeleteWithClaims(url string, loginID int64, root bool) *httptest.ResponseRecorder {
  h := make(map[string]interface{})
  r := otape.BuildDataRequest("DELETE", url, h)
  addJWTClaims(r, loginID, root)
  return t.PlayRequest(r)
}

func (t *Tape) UploadWithClaims(url string, filename string, contentType string, loginID int64, root bool) (*httptest.ResponseRecorder, error) {

  body, ct, err := otape.CreateFileRequestBody(filename, contentType)
  if err != nil {
    return nil, err
  }

  r, err := http.NewRequest("POST", url, body)
  if err != nil {
    return nil, err
  }
  r.Header.Set("Content-Type", ct)

  addJWTClaims(r, loginID, root)
  return t.PlayRequest(r), nil
}

func (t *Tape) ToH(z interface{}) map[string]interface{} {
  data, _ := json.Marshal(z)
  var msgMapTemplate interface{}
  _ = json.Unmarshal(data, &msgMapTemplate)
  return msgMapTemplate.(map[string]interface{})
}
