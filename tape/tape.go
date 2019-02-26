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

package tape

import (
  "bytes"
  "encoding/json"
  "log"
  "net/http"
  "net/http/httptest"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/go-chi/chi"
  "github.com/jmoiron/sqlx"
  homedir "github.com/mitchellh/go-homedir"
  "github.com/spf13/viper"
)

// TOKEN --------------------------------------------------
var tokenManager *authenticate.TokenAuth

func SetConfigFile() {

  // Find home directory.
  home, err := homedir.Dir()
  if err != nil {
    log.Fatal(err)
  }

  // Search config in home directory with name ".go-base" (without extension).
  viper.AddConfigPath(home)
  viper.SetConfigName(".infomark-backend")
}
func InitConfig() {

  SetConfigFile()
  viper.AutomaticEnv()

  // If a config file is found, read it in.
  if err := viper.ReadInConfig(); err == nil {
    // fmt.Println("Using config file:", viper.ConfigFileUsed())
  }
}

func init() {
  InitConfig()
  tokenManager, _ = authenticate.NewTokenAuth()
}

// TOKEN --------------------------------------------------

// type Testerer interface {
//   BeforeEach(*chi.Mux) (*sqlx.DB)
//   AfterEach()

//   Play(method, url string, body io.Reader) *httptest.ResponseRecorder
//   PlayData(method, url string, data map[string]interface{}) *httptest.ResponseRecorder

//   PlayRequest(r *http.Request) *httptest.ResponseRecorder
//   PlayRequestWithClaims(r *http.Request, loginID int64, root bool) *httptest.ResponseRecorder
// }

type Tape struct {
  DB     *sqlx.DB
  Router *chi.Mux
  // Stores *app.Stores
}

func NewTape() *Tape {
  return &Tape{}
}

// (*sqlx.DB, *app.Stores, *chi.Mux)
func (t *Tape) BeforeEach() {
  var err error

  t.DB, err = helper.TransactionDB()
  if err != nil {
    panic(err)
  }

  // t.Stores = app.NewStores(t.DB)
  // t.Router = router
  // t.Router, err = api.New(t.DB, false)
  // if err != nil {
  //   panic(err)
  // }

}

func (t *Tape) AfterEach() {
  t.DB.Close()
}

func (t *Tape) Play(method, url string) *httptest.ResponseRecorder {

  r, err := http.NewRequest(method, url, nil)
  if err != nil {
    panic(err)
  }

  r.Header.Set("Content-Type", "application/json")

  return t.PlayRequest(r)
}

func (t *Tape) PlayWithClaims(method, url string, loginID int64, root bool) *httptest.ResponseRecorder {

  r, err := http.NewRequest(method, url, nil)
  if err != nil {
    panic(err)
  }

  r.Header.Set("Content-Type", "application/json")
  addClaims(r, loginID, root)
  return t.PlayRequest(r)
}

func (t *Tape) PlayData(method, url string, data map[string]interface{}) *httptest.ResponseRecorder {

  var payload_json *bytes.Buffer

  if data != nil {
    dat, err := json.Marshal(data)
    if err != nil {
      panic(err)
    }
    payload_json = bytes.NewBuffer(dat)
  } else {
    payload_json = nil
  }

  r, err := http.NewRequest(method, url, payload_json)
  if err != nil {
    panic(err)
  }

  r.Header.Set("Content-Type", "application/json")

  return t.PlayRequest(r)
}

func (t *Tape) PlayDataWithClaims(method, url string, data map[string]interface{}, loginID int64, root bool) *httptest.ResponseRecorder {

  var payload_json *bytes.Buffer

  if data != nil {
    dat, err := json.Marshal(data)
    if err != nil {
      panic(err)
    }
    payload_json = bytes.NewBuffer(dat)
  } else {
    payload_json = nil
  }

  r, err := http.NewRequest(method, url, payload_json)
  if err != nil {
    panic(err)
  }

  r.Header.Set("Content-Type", "application/json")
  addClaims(r, loginID, root)
  return t.PlayRequest(r)
}

// payload_json, _ := json.Marshal(request.Data)
//   r, _ := http.NewRequest(request.Method, "/", bytes.NewBuffer(payload_json))

func (t *Tape) PlayRequest(r *http.Request) *httptest.ResponseRecorder {
  w := httptest.NewRecorder()
  t.Router.ServeHTTP(w, r)
  return w
}

func addClaims(r *http.Request, loginID int64, root bool) {
  accessToken, err := tokenManager.CreateAccessJWT(authenticate.NewAccessClaims(loginID, root))
  if err != nil {
    panic(err)
  }

  r.Header.Add("Authorization", "Bearer "+accessToken)

}

func (t *Tape) PlayRequestWithClaims(r *http.Request, loginID int64, root bool) *httptest.ResponseRecorder {
  w := httptest.NewRecorder()
  addClaims(r, loginID, root)
  t.Router.ServeHTTP(w, r)
  return w
}

// func addAccessClaimsIfNeeded(r *http.Request, request Payload) *http.Request {
//   // If there are some access claims, we add them to the header.
//   // We currently support JWT only for testing.
//   if request.AccessClaims.LoginID != 0 {
//     // generate some valid claims
//     accessToken, err := tokenManager.CreateAccessJWT(authenticate.NewAccessClaims(request.AccessClaims.LoginID, request.AccessClaims.Root))
//     if err != nil {
//       panic(err)
//     }
//     r.Header.Add("Authorization", "Bearer "+accessToken)
//   }
//   return r
// }
