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
  "fmt"
  "net/http"
  "testing"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/franela/goblin"
)

func TestUser(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

  g.Describe("User", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      stores = NewStores(tape.DB)
      _ = stores
    })

    g.It("Query should require access claims", func() {
      w := tape.Get("/api/v1/users")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/users", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Query should list all users", func() {
      users_expected, err := stores.User.GetAll()
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/users", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      users_actual := []model.User{}
      err = json.NewDecoder(w.Body).Decode(&users_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(users_actual)).Equal(len(users_expected))
    })

    g.It("Should get a specific user", func() {

      user_expected, err := stores.User.Get(1)
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/users/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      user_actual := &userResponse{}
      err = json.NewDecoder(w.Body).Decode(user_actual)
      g.Assert(err).Equal(nil)

      g.Assert(user_actual.ID).Equal(user_expected.ID)

      g.Assert(user_actual.FirstName).Equal(user_expected.FirstName)
      g.Assert(user_actual.LastName).Equal(user_expected.LastName)
      g.Assert(user_actual.Email).Equal(user_expected.Email)
      g.Assert(user_actual.StudentNumber).Equal(user_expected.StudentNumber)
      g.Assert(user_actual.Semester).Equal(user_expected.Semester)
      g.Assert(user_actual.Subject).Equal(user_expected.Subject)
      g.Assert(user_actual.Language).Equal(user_expected.Language)

    })

    g.Xit("Should send email", func() {})

    g.It("Changes should require access claims", func() {
      w := tape.Put("/api/v1/users/1", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.It("Should perform updates (incl email)", func() {
      // this is NOT the /me enpoint, we can update the user here

      user_db, err := stores.User.Get(1)
      g.Assert(err).Equal(nil)

      user_sent := &userRequest{
        FirstName: "Info2_update",
        LastName:  "Lorem Ipsum_update",
        Email:     "new@mail.com",
        Semester:  1,

        StudentNumber: user_db.StudentNumber,
        Subject:       user_db.Subject,
        Language:      user_db.Language,
      }

      err = user_sent.Validate()
      g.Assert(err).Equal(nil)

      w := tape.PutWithClaims("/api/v1/users/1", helper.ToH(user_sent), 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      user_after, err := stores.User.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(user_after.FirstName).Equal(user_sent.FirstName)
      g.Assert(user_after.LastName).Equal(user_sent.LastName)
      g.Assert(user_after.Email).Equal(user_sent.Email)

    })

    g.It("Should delete whith claims", func() {
      users_before, err := stores.User.GetAll()
      g.Assert(err).Equal(nil)

      w := tape.Delete("/api/v1/users/1")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.DeleteWithClaims("/api/v1/users/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      users_after, err := stores.User.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(users_after)).Equal(len(users_before) - 1)
    })

    g.It("Self-query require claims", func() {
      w := tape.Get("/api/v1/me")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/me", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Should list myself", func() {
      user_expected, err := stores.User.Get(1)
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/me", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      user_actual := &userResponse{}
      err = json.NewDecoder(w.Body).Decode(&user_actual)
      g.Assert(err).Equal(nil)

      g.Assert(user_actual.FirstName).Equal(user_expected.FirstName)
      g.Assert(user_actual.LastName).Equal(user_expected.LastName)
      g.Assert(user_actual.Email).Equal(user_expected.Email)
      g.Assert(user_actual.StudentNumber).Equal(user_expected.StudentNumber)
      g.Assert(user_actual.Semester).Equal(user_expected.Semester)
      g.Assert(user_actual.Subject).Equal(user_expected.Subject)
      g.Assert(user_actual.Language).Equal(user_expected.Language)
    })

    g.Xit("Should not perform self-updates when some data is missing", func() {
      // This endpoint acts like a PATCH, since we need to start anyway from the
      // database entry to avoid overriding "email".
      // Theoretically (by definition of PUT) this endpoint must fail.
      // But in practise, it is ever to act like PATCH here and pass also this request.
      data_sent := H{
        "first_name": "blub",
      }

      w := tape.PutWithClaims("/api/v1/me", data_sent, 1, true)
      g.Assert(w.Code).Equal(http.StatusBadRequest)
    })

    g.It("Should perform self-updates (excl email)", func() {
      user_before, err := stores.User.Get(1)
      g.Assert(err).Equal(nil)

      user_db, err := stores.User.Get(1)
      g.Assert(err).Equal(nil)

      // this is NOT the /me enpoint, we can update the user here

      user_sent := &userRequest{
        FirstName: "Info2_update",
        LastName:  "Lorem Ipsum_update",
        Email:     "new@mail.com",
        Semester:  1,

        StudentNumber: user_db.StudentNumber,
        Subject:       user_db.Subject,
        Language:      user_db.Language,
      }

      err = user_sent.Validate()
      g.Assert(err).Equal(nil)

      err = user_sent.Validate()
      g.Assert(err).Equal(nil)

      w := tape.PutWithClaims("/api/v1/me", helper.ToH(user_sent), 1, true)
      fmt.Println(w.Body)
      g.Assert(w.Code).Equal(http.StatusOK)

      user_after, err := stores.User.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(user_after.FirstName).Equal(user_sent.FirstName)
      g.Assert(user_after.LastName).Equal(user_sent.LastName)
      g.Assert(user_after.Email).Equal(user_before.Email) // should be really before
      g.Assert(user_after.StudentNumber).Equal(user_sent.StudentNumber)
      g.Assert(user_after.Semester).Equal(user_sent.Semester)
      g.Assert(user_after.Subject).Equal(user_sent.Subject)
      g.Assert(user_after.Language).Equal(user_sent.Language)

    })

    g.AfterEach(func() {
      tape.AfterEach()
    })
  })

}
