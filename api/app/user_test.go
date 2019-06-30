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
	"net/http"
	"testing"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/franela/goblin"
)

func TestUser(t *testing.T) {
	PrepareTests()
	g := goblin.Goblin(t)
	email.DefaultMail = email.VoidMail

	tape := &Tape{}

	var stores *Stores

	adminJWT := NewJWTRequest(1, true)

	g.Describe("User", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
			stores = NewStores(tape.DB)
			_ = stores
		})

		g.It("Query should require access claims", func() {
			w := tape.Get("/api/v1/users")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/users", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Query should list all users", func() {
			usersExpected, err := stores.User.GetAll()
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/users", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			usersActual := []model.User{}
			err = json.NewDecoder(w.Body).Decode(&usersActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(usersActual)).Equal(len(usersExpected))
		})

		g.It("Query should find a user", func() {
			usersExpected, err := stores.User.Find("%%meinhard%%")
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/users/find?query=meinhard", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			usersActual := []model.User{}
			err = json.NewDecoder(w.Body).Decode(&usersActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(usersActual)).Equal(len(usersExpected))
		})

		g.It("Should get a specific user", func() {

			userExpected, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/users/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			userActual := &UserResponse{}
			err = json.NewDecoder(w.Body).Decode(userActual)
			g.Assert(err).Equal(nil)

			g.Assert(userActual.ID).Equal(userExpected.ID)

			g.Assert(userActual.FirstName).Equal(userExpected.FirstName)
			g.Assert(userActual.LastName).Equal(userExpected.LastName)
			g.Assert(userActual.Email).Equal(userExpected.Email)
			g.Assert(userActual.StudentNumber).Equal(userExpected.StudentNumber)
			g.Assert(userActual.Semester).Equal(userExpected.Semester)
			g.Assert(userActual.Subject).Equal(userExpected.Subject)
			g.Assert(userActual.Language).Equal(userExpected.Language)

		})

		g.Xit("Should send email", func() {})

		g.It("Changes should require access claims", func() {
			w := tape.Put("/api/v1/users/1", H{})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should perform updates (incl email)", func() {
			// this is NOT the /me enpoint, we can update the user here

			userDb, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)

			userSent := &userRequest{
				FirstName: "Info2_update",
				LastName:  "Lorem Ipsum_update",
				Email:     "new@mail.com",
				Semester:  1,

				StudentNumber: userDb.StudentNumber,
				Subject:       userDb.Subject,
				Language:      userDb.Language,
			}

			err = userSent.Validate()
			g.Assert(err).Equal(nil)

			w := tape.Put("/api/v1/users/1", helper.ToH(userSent), adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			userAfter, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(userAfter.FirstName).Equal(userSent.FirstName)
			g.Assert(userAfter.LastName).Equal(userSent.LastName)
			g.Assert(userAfter.Email).Equal(userSent.Email)

		})

		g.It("Should delete whith claims", func() {
			usersBefore, err := stores.User.GetAll()
			g.Assert(err).Equal(nil)

			w := tape.Delete("/api/v1/users/1")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Delete("/api/v1/users/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			usersAfter, err := stores.User.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(usersAfter)).Equal(len(usersBefore) - 1)
		})

		g.It("Self-query require claims", func() {
			w := tape.Get("/api/v1/me")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/me", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Should list myself", func() {
			userExpected, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/me", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			userActual := &UserResponse{}
			err = json.NewDecoder(w.Body).Decode(&userActual)
			g.Assert(err).Equal(nil)

			g.Assert(userActual.FirstName).Equal(userExpected.FirstName)
			g.Assert(userActual.LastName).Equal(userExpected.LastName)
			g.Assert(userActual.Email).Equal(userExpected.Email)
			g.Assert(userActual.StudentNumber).Equal(userExpected.StudentNumber)
			g.Assert(userActual.Semester).Equal(userExpected.Semester)
			g.Assert(userActual.Subject).Equal(userExpected.Subject)
			g.Assert(userActual.Language).Equal(userExpected.Language)
		})

		g.Xit("Should not perform self-updates when some data is missing", func() {
			// This endpoint acts like a PATCH, since we need to start anyway from the
			// database entry to avoid overriding "email".
			// Theoretically (by definition of PUT) this endpoint must fail.
			// But in practise, it is ever to act like PATCH here and pass also this request.
			dataSent := H{
				"first_name": "blub",
			}

			w := tape.Put("/api/v1/me", dataSent, adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should perform self-updates (excl email)", func() {
			userBefore, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)

			userDb, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)

			// this is NOT the /me enpoint, we can update the user here

			userSent := &userRequest{
				FirstName: "Info2_update",
				LastName:  "Lorem Ipsum_update",
				Email:     "new@mail.com",
				Semester:  1,

				StudentNumber: userDb.StudentNumber,
				Subject:       userDb.Subject,
				Language:      userDb.Language,
			}

			err = userSent.Validate()
			g.Assert(err).Equal(nil)

			err = userSent.Validate()
			g.Assert(err).Equal(nil)

			w := tape.Put("/api/v1/me", helper.ToH(userSent), adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			userAfter, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(userAfter.FirstName).Equal(userSent.FirstName)
			g.Assert(userAfter.LastName).Equal(userSent.LastName)
			g.Assert(userAfter.Email).Equal(userBefore.Email) // should be really before
			g.Assert(userAfter.StudentNumber).Equal(userSent.StudentNumber)
			g.Assert(userAfter.Semester).Equal(userSent.Semester)
			g.Assert(userAfter.Subject).Equal(userSent.Subject)
			g.Assert(userAfter.Language).Equal(userSent.Language)

		})

		g.AfterEach(func() {
			tape.AfterEach()
		})
	})

}
