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
