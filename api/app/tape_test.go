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
	"os"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	otape "github.com/cgtuebingen/infomark-backend/tape"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

type JWTRequest struct {
	Claims authenticate.AccessClaims
}

func (t JWTRequest) Modify(r *http.Request) {
	accessToken, err := tokenManager.CreateAccessJWT(t.Claims)
	if err != nil {
		panic(err)
	}
	// we use JWT here
	r.Header.Add("Authorization", "Bearer "+accessToken)
}

func NewJWTRequest(loginID int64, root bool) JWTRequest {
	return JWTRequest{
		Claims: authenticate.NewAccessClaims(loginID, root),
	}
}

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

func PrepareTests() {
	helper.FakeDatabase()
	PrepareTokenManager()
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
