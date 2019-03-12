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

      grades_actual := []GradeResponse{}
      err := json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(grades_actual)).Equal(229)
    })

    g.It("Should list all grades of a group with some filters", func() {

      w := tape.GetWithClaims("/api/v1/courses/1/grades?group_id=1&public_test_status=0", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
      grades_actual := []GradeResponse{}
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

    g.It("Should perform updates", func() {

      data := H{
        "acquired_points": 3,
        "feedback":        "Lorem Ipsum_update",
      }

      w := tape.Put("/api/v1/courses/1/grades/1", data)
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // students
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // tutors
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 3, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_after, err := stores.Grade.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(entry_after.Feedback).Equal("Lorem Ipsum_update")
      g.Assert(entry_after.AcquiredPoints).Equal(3)
      g.Assert(entry_after.TutorID).Equal(int64(3))
    })

    g.It("Should list missing grades", func() {
      result := []MissingGradeResponse{}
      // students have no missing data
      // but we do not know if a user is student in a course
      w := tape.GetWithClaims("/api/v1/courses/1/grades/missing", 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)
      err := json.NewDecoder(w.Body).Decode(&result)
      g.Assert(err).Equal(nil)
      g.Assert(len(result)).Equal(0)

      // admin (mock creates feed back for all submissions)
      w = tape.GetWithClaims("/api/v1/courses/1/grades/missing", 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)
      err = json.NewDecoder(w.Body).Decode(&result)
      g.Assert(err).Equal(nil)
      g.Assert(len(result)).Equal(0)

      // tutors (mock creates feed back for all submissions)
      w = tape.GetWithClaims("/api/v1/courses/1/grades/missing", 3, false)
      g.Assert(w.Code).Equal(http.StatusOK)
      err = json.NewDecoder(w.Body).Decode(&result)
      g.Assert(err).Equal(nil)
      g.Assert(len(result)).Equal(0)

      _, err = tape.DB.Exec("UPDATE grades SET feedback='' WHERE tutor_id = 3 ")
      g.Assert(err).Equal(nil)

      // tutors (mock creates feed back for all submissions)
      w = tape.GetWithClaims("/api/v1/courses/1/grades/missing", 3, false)
      g.Assert(w.Code).Equal(http.StatusOK)
      err = json.NewDecoder(w.Body).Decode(&result)
      g.Assert(err).Equal(nil)
      // see mock.py
      g.Assert(len(result)).Equal(945)

      for _, el := range result {
        g.Assert(el.Grade.TutorID).Equal(int64(3))
        g.Assert(el.Grade.Feedback).Equal("")
      }
    })

    g.It("Should handle feedback from public tests", func() {

      url := "/api/v1/courses/1/grades/1/public_result"

      data := H{
        "log":    "some new logs",
        "status": 2,
      }

      w := tape.Post(url, data)
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // students
      w = tape.PostWithClaims(url, data, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.PostWithClaims(url, data, 3, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PostWithClaims(url, data, 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_after, err := stores.Grade.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(entry_after.PublicTestLog).Equal("some new logs")
      g.Assert(entry_after.PublicTestStatus).Equal(2)

    })

    g.It("Should handle feedback from private tests", func() {

      url := "/api/v1/courses/1/grades/1/private_result"

      data := H{
        "log":    "some new logs",
        "status": 2,
      }

      w := tape.Post(url, data)
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // students
      w = tape.PostWithClaims(url, data, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.PostWithClaims(url, data, 3, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PostWithClaims(url, data, 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_after, err := stores.Grade.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(entry_after.PrivateTestLog).Equal("some new logs")
      g.Assert(entry_after.PrivateTestStatus).Equal(2)

    })

    g.AfterEach(func() {
      tape.AfterEach()
    })
  })

}
