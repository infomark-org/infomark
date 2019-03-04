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
  "fmt"
  "net/http"
  "testing"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/franela/goblin"
  "github.com/spf13/viper"
)

func TestSubmission(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

  g.Describe("Submission", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      stores = NewStores(tape.DB)
      _ = stores
    })

    g.It("Query should require access claims", func() {

      w := tape.Get("/api/v1/tasks/1/submission")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/tasks/1/submission", 112, false)
      g.Assert(w.Code).Equal(http.StatusNotFound)
    })

    g.It("Students can upload solution", func() {

      defer helper.NewSubmissionFileHandle(3001).Delete()

      // no files so far
      g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)

      // upload
      filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
      w, err := tape.UploadWithClaims("/api/v1/tasks/1/submission", filename, "application/zip", 112, false)
      g.Assert(err).Equal(nil)
      g.Assert(w.Code).Equal(http.StatusOK)
      g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(true)

      // files exists
      w = tape.GetWithClaims("/api/v1/tasks/1/submission", 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

    })

    g.Xit("Students can only access their own submissions", func() {

      defer helper.NewSubmissionFileHandle(3001).Delete()

      // no files so far
      g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)

      // upload
      filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
      w, err := tape.UploadWithClaims("/api/v1/tasks/1/submission", filename, "application/zip", 112, false)
      g.Assert(err).Equal(nil)
      g.Assert(w.Code).Equal(http.StatusOK)
      g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(true)

      // access own submission
      w = tape.GetWithClaims("/api/v1/submissions/3001", 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // access others submission
      w = tape.GetWithClaims("/api/v1/submissions/3001", 113, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

    })

    g.AfterEach(func() {
      tape.AfterEach()
    })
  })

}
