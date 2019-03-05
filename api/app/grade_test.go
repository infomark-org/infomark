// InfoMark - a platform for managing courses with
//            distributing exercise materials and testing exercise submissions
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

  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/franela/goblin"
)

func TestGrade(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

  g.Describe("Grade", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      stores = NewStores(tape.DB)
      _ = stores
    })

    g.It("Query should require access claims", func() {
      url := "/api/v1/courses/1/grades?group_id=1"
      w := tape.Get(url)
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims(url, 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Should list all grades of a group", func() {
      url := "/api/v1/courses/1/grades?group_id=1"

      w := tape.GetWithClaims(url, 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      grades_actual := []model.Grade{}
      err := json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(grades_actual)).Equal(186)
    })

    g.It("Should list all grades of a group with some filters", func() {

      w := tape.GetWithClaims("/api/v1/courses/1/grades?group_id=1&public_test_status=0", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
      grades_actual := []model.Grade{}
      err := json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)
      for _, el := range grades_actual {
        g.Assert(el.PublicTestStatus).Equal(0)
      }

      w = tape.GetWithClaims("/api/v1/courses/1/grades?group_id=1&private_test_status=0", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
      err = json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)
      for _, el := range grades_actual {
        g.Assert(el.PrivateTestStatus).Equal(0)
      }

      w = tape.GetWithClaims("/api/v1/courses/1/grades?group_id=1&tutor_id=3", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
      err = json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)
      for _, el := range grades_actual {
        g.Assert(el.TutorID).Equal(int64(3))
      }
    })

    g.AfterEach(func() {
      tape.AfterEach()
    })
  })

}
