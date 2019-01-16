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
	"github.com/cgtuebingen/infomark-backend/router/helper"
	"github.com/cgtuebingen/infomark-backend/store"
	"github.com/franela/goblin"
	. "github.com/franela/goblin"
)

func AssertEmptyUserTable(g *goblin.G) {
	count, err := store.DS().CountUsers()
	g.Assert(err).Equal(nil)
	g.Assert(count).Equal(0)
}

func TestUserCreate(t *testing.T) {

	store.ORM().Exec("TRUNCATE users;")
	store.ORM().Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1;")

	g := Goblin(t)
	g.Describe("UsersRouter", func() {

		g.It("Should have empty user table", func() {
			AssertEmptyUserTable(g)
		})

		g.It("Should reject to create invalid users", func() {
			// missing LastName
			w := helper.SimulateRequest(helper.H{
				// the API will suppress the `id` field
				"id":        33,
				"last_name": "Zwegat",
				"email":     "peter.zwegat@uni-tuebingen.de",
			}, UsersCreate,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
			AssertEmptyUserTable(g)

			// missing FirstName
			w = helper.SimulateRequest(helper.H{
				// the API will suppress the `id` field
				"id":         33,
				"first_name": "Peter",
				"email":      "peter.zwegat@uni-tuebingen.de",
			}, UsersCreate,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
			AssertEmptyUserTable(g)

		})

		g.It("Should create valid users", func() {
			w := helper.SimulateRequest(helper.H{
				// the API will suppress the `id` field
				"id":         33,
				"first_name": "Peter",
				"last_name":  "Zwegat",
				"email":      "peter.zwegat@uni-tuebingen.de",
				"password":   "demo123",
			}, UsersCreate,
			)
			g.Assert(w.Code).Equal(http.StatusCreated)

			// unmarshal
			user_returned := &model.User{}
			err := json.Unmarshal(w.Body.Bytes(), user_returned)
			g.Assert(err).Equal(nil)

			// returned information should be correct
			g.Assert(user_returned.ID).Equal(uint(1))
			g.Assert(user_returned.FirstName).Equal("Peter")
			g.Assert(user_returned.LastName).Equal("Zwegat")
			g.Assert(user_returned.Email).Equal("peter.zwegat@uni-tuebingen.de")
			// we skip the hash in User.render
			g.Assert(user_returned.PasswordHash).Equal("")

			// database should be correct
			user_db := &model.User{}
			store.ORM().First(&user_db, 1)

			g.Assert(user_db.FirstName).Equal("Peter")
			g.Assert(user_db.LastName).Equal("Zwegat")
			g.Assert(user_db.Email).Equal("peter.zwegat@uni-tuebingen.de")

			// correct password should be saved
			g.Assert(helper.CheckPasswordHash("demo123", user_db.PasswordHash)).IsTrue()

		})

	})

	store.ORM().Exec("TRUNCATE users;")
	store.ORM().Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1;")
}
