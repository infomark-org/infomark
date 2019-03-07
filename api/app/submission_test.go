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

    g.It("Students can upload solution (create)", func() {

      defer helper.NewSubmissionFileHandle(3001).Delete()

      // no files so far
      g.Assert(helper.NewSubmissionFileHandle(3001).Exists()).Equal(false)

      // remove all submission from student
      _, err := tape.DB.Exec("DELETE FROM submissions WHERE user_id = 112;")
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/tasks/1/submission", 112, false)
      g.Assert(w.Code).Equal(http.StatusNotFound)

      // upload
      filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
      w, err = tape.UploadWithClaims("/api/v1/tasks/1/submission", filename, "application/zip", 112, false)
      g.Assert(err).Equal(nil)
      g.Assert(w.Code).Equal(http.StatusOK)

      createdSubmission, err := stores.Submission.GetByUserAndTask(112, 1)
      g.Assert(err).Equal(nil)

      g.Assert(helper.NewSubmissionFileHandle(createdSubmission.ID).Exists()).Equal(true)
      defer helper.NewSubmissionFileHandle(createdSubmission.ID).Delete()

      // files exists
      w = tape.GetWithClaims("/api/v1/tasks/1/submission", 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

    })

    g.It("Students can upload solution (update)", func() {

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

    g.It("Students can only access their own submissions", func() {

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

    g.It("tutors/admins can filter submissions", func() {

      w := tape.Get("/api/v1/courses/1/submissions")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses/1/submissions", 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      w = tape.GetWithClaims("/api/v1/courses/1/submissions", 2, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      submissions_all_actual := []SubmissionResponse{}
      err := json.NewDecoder(w.Body).Decode(&submissions_all_actual)
      g.Assert(err).Equal(nil)

      w = tape.GetWithClaims("/api/v1/courses/1/submissions?group_id=4", 2, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      submissions_g4_actual := []SubmissionResponse{}
      err = json.NewDecoder(w.Body).Decode(&submissions_g4_actual)
      g.Assert(err).Equal(nil)

      w = tape.GetWithClaims("/api/v1/courses/1/submissions?task_id=2", 2, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      submissions_t4_actual := []SubmissionResponse{}
      err = json.NewDecoder(w.Body).Decode(&submissions_t4_actual)
      g.Assert(err).Equal(nil)

      for _, el := range submissions_t4_actual {
        g.Assert(el.TaskID).Equal(int64(2))
      }

      w = tape.GetWithClaims("/api/v1/courses/1/submissions?user_id=112", 2, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      submissions_u112_actual := []SubmissionResponse{}
      err = json.NewDecoder(w.Body).Decode(&submissions_u112_actual)
      g.Assert(err).Equal(nil)

      for _, el := range submissions_u112_actual {
        g.Assert(el.UserID).Equal(int64(112))
      }

    })

    g.AfterEach(func() {
      tape.AfterEach()
    })
  })

}
