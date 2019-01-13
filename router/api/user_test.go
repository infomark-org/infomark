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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/cgtuebingen/infomark-backend/router/helper"
	"github.com/cgtuebingen/infomark-backend/store"
	"github.com/franela/goblin"
	. "github.com/franela/goblin"
)

func SimulateRequest(payload interface{}, api func(w http.ResponseWriter, r *http.Request)) *httptest.ResponseRecorder {

	payload_json, _ := json.Marshal(payload)
	r, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(payload_json))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api(w, r)
	return w
}

func AssertEmptyUserTable(g *goblin.G) {
	count, err := store.DS().CountUsers()
	g.Assert(err).Equal(nil)
	g.Assert(count).Equal(0)
}

func TestUserCreate(t *testing.T) {

	g := Goblin(t)
	g.Describe("UsersRouter", func() {

		store.ORM().Exec("TRUNCATE users;")
		store.ORM().Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1;")

		g.It("Should have empty user table", func() {
			AssertEmptyUserTable(g)
		})

		g.It("Should reject to create invalid users", func() {
			// missing LastName
			w := SimulateRequest(helper.H{
				"id":        33,
				"last_name": "Zwegat",
				"email":     "peter.zwegat@uni-tuebingen.de",
			}, UserCreate,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
			AssertEmptyUserTable(g)

			// missing FirstName
			w = SimulateRequest(helper.H{
				"id":         33,
				"first_name": "Peter",
				"email":      "peter.zwegat@uni-tuebingen.de",
			}, UserCreate,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
			AssertEmptyUserTable(g)

		})

		g.It("Should create valid users", func() {
			w := SimulateRequest(helper.H{
				"id":         33,
				"first_name": "Peter",
				"last_name":  "Zwegat",
				"email":      "peter.zwegat@uni-tuebingen.de",
				"password":   "demo123",
			}, UserCreate,
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
			g.Assert(helper.CheckPasswordHash("demo123", user_db.PasswordHash)).Equal(true)

		})

	})
}

func TestUserGet(t *testing.T) {

	g := Goblin(t)
	g.Describe("UsersRouter", func() {
	})
}
