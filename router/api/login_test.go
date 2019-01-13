// Copyright 2019 ComputerGraphics Tuebingen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ==============================================================================
// Authors: Patrick Wieschollek

package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/cgtuebingen/infomark-backend/router/auth"
	"github.com/cgtuebingen/infomark-backend/router/helper"
	"github.com/cgtuebingen/infomark-backend/store"
	jwt "github.com/dgrijalva/jwt-go"
	. "github.com/franela/goblin"
)

func TestLogin(t *testing.T) {

	store.ORM().Exec("TRUNCATE users;")
	store.ORM().Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1;")

	auth.InitializeJWT("demo")

	g := Goblin(t)
	g.Describe("LoginRouter", func() {

		hash, _ := helper.HashPassword("demo123")

		user := &model.User{
			LastName:     "Zwegat",
			FirstName:    "Peter",
			Email:        "peter.zwegat@uni-tuebingen.de",
			PasswordHash: hash,
		}

		err := store.ORM().Create(&user).Error
		if err != nil {
			panic(err)
		}

		g.It("Should not get JWT token (wrong email)", func() {
			// wrong login
			w := helper.SimulateRequest(helper.H{
				"email":    "test@example.com",
				"password": "demo",
			}, Login,
			)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

		})

		g.It("Should not get JWT token (wrong password)", func() {
			// wrong login
			w := helper.SimulateRequest(helper.H{
				"email":    user.Email,
				"password": "321omed",
			}, Login,
			)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

		})

		g.It("Should get JWT token (correct credentials)", func() {
			// wrong login
			w := helper.SimulateRequest(helper.H{
				"email":    user.Email,
				"password": "demo123",
			}, Login,
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			// parse result
			resp := &LoginResponse{}
			err := json.Unmarshal(w.Body.Bytes(), resp)
			g.Assert(err).Equal(nil)
			g.Assert(resp.Success).IsTrue()
			g.Assert(resp.Token == "").IsFalse()

			// verify token
			token, err := auth.GetTokenAuth().Decode(resp.Token)
			g.Assert(err).Equal(nil)
			g.Assert(token.Valid).IsTrue()

			// verify claim
			tokenClaims, ok := token.Claims.(jwt.MapClaims)
			g.Assert(ok).IsTrue()
			login_id, ok := tokenClaims["login_id"].(float64)
			g.Assert(ok).IsTrue()
			g.Assert(int(login_id)).Equal(int(user.ID))

		})

	})

	store.ORM().Exec("TRUNCATE users;")
	store.ORM().Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1;")
}
