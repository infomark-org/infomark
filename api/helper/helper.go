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

package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	txdb "github.com/DATA-DOG/go-txdb"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// similar to gin.H as a neat wrapper
type H map[string]interface{}

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
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func init() {
	InitConfig()
	tokenManager, _ = authenticate.NewTokenAuth()
}

type Payload struct {
	Data         H
	Method       string
	AccessClaims authenticate.AccessClaims
}

// https://github.com/go-chi/chi/blob/cca4135d8dddff765463feaf1118047a9e506b4a/chain.go#L34-L49
// type Handler interface {
//         ServeHTTP(ResponseWriter, *Request)
// }
// type HandlerFunc func(ResponseWriter, *Request)
//
// chain builds a http.Handler composed of an inline middleware stack and endpoint
// handler in the order they are passed.
func chain(endpoint http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Return ahead of time if there aren't any middlewares for the chain
	if len(middlewares) == 0 {
		return endpoint
	}

	// Wrap the end handler with the middleware chain
	h := middlewares[len(middlewares)-1](endpoint)
	for i := len(middlewares) - 2; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}

func formatRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}

func SimulateRequestWithURL(
	// payload interface{},
	request Payload,
	url string,
	apiHandler http.HandlerFunc,
	middlewares ...func(http.Handler) http.Handler) *httptest.ResponseRecorder {

	// create request
	payload_json, _ := json.Marshal(request.Data)
	r, _ := http.NewRequest(request.Method, url, bytes.NewBuffer(payload_json))
	r.Header.Set("Content-Type", "application/json")

	// If there are some access claims, we add them to the header.
	// We currently support JWT only for testing.
	if request.AccessClaims.LoginID != 0 {
		// generate some valid claims
		accessToken, err := tokenManager.CreateAccessJWT(authenticate.NewAccessClaims(1, true))
		if err != nil {
			panic(err)
		}
		r.Header.Add("Authorization", "Bearer "+accessToken)
	}

	// fmt.Println(formatRequest(r))

	w := httptest.NewRecorder()

	// apply middlewares
	handler := chain(apiHandler, middlewares...)
	handler.ServeHTTP(w, r)

	return w
}

func SimulateRequest(
	// payload interface{},
	request Payload,
	apiHandler http.HandlerFunc,
	middlewares ...func(http.Handler) http.Handler) *httptest.ResponseRecorder {

	return SimulateRequestWithURL(request, "/", apiHandler, middlewares...)

}

func init() {
	// we register an sql driver named "txdb"
	// This allows to run all tests as transaction in isolated environemnts to make sure
	// we do not accidentially alter the database in a persistent way. Hence,  all tests can run
	// in an arbitrary order.
	txdb.Register("psql_txdb", "postgres", "postgres://postgres:postgres@localhost/infomark?sslmode=disable")
}

// TransactionDB creates a sql-driver which seemlessly supports transactions.
func TransactionDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect("psql_txdb", "identifier")

	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, err
}
