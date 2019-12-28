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

package app

import (
	"net/http"

	txdb "github.com/DATA-DOG/go-txdb"

	"github.com/infomark-org/infomark-backend/auth/authenticate"
	"github.com/infomark-org/infomark-backend/configuration"
	otape "github.com/infomark-org/infomark-backend/tape"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // need for Postgres
)

type JWTRequest struct {
	Claims    authenticate.AccessClaims
	TokenAuth *authenticate.TokenAuth
}

func (t JWTRequest) Modify(r *http.Request) {
	accessToken, err := t.TokenAuth.CreateAccessJWT(t.Claims)
	if err != nil {
		panic(err)
	}
	// we use JWT here
	r.Header.Add("Authorization", "Bearer "+accessToken)
}

func (t *Tape) NewJWTRequest(loginID int64, root bool) JWTRequest {
	return JWTRequest{
		Claims:    authenticate.NewAccessClaims(loginID, root),
		TokenAuth: t.TokenAuth,
	}
}

type Tape struct {
	otape.Tape
	DB        *sqlx.DB
	TokenAuth *authenticate.TokenAuth
}

var registered_txdb = false

func NewTape() *Tape {

	configuration.MustFindAndReadConfiguration()
	// Ensure transaction between tests
	if !registered_txdb {
		txdb.Register("psql_txdb", "postgres", configuration.Configuration.Server.PostgresURL())
		registered_txdb = true
	}

	return &Tape{
		TokenAuth: authenticate.NewTokenAuth(&configuration.Configuration.Server.Authentication),
	}
}

func (t *Tape) AfterEach() {
	t.DB.Close()
}

// TransactionDB creates a sql-driver which seemlessly supports transactions.
// This is used for running the unit tests.
func transactionDB() (*sqlx.DB, error) {
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

func (t *Tape) BeforeEach() {
	var err error

	t.DB, err = transactionDB()
	if err != nil {
		panic(err)
	}

	t.Router, _ = New(t.DB, false)

}
