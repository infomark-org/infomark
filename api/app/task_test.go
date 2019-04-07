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
  "github.com/franela/goblin"
  "github.com/spf13/viper"
)

func TestTask(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

  g.Describe("Task", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      stores = NewStores(tape.DB)
    })

    g.It("Query should require access claims", func() {

      w := tape.Get("/api/v1/courses/1/sheets/1/tasks")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses/1/sheets/1/tasks", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Should list all tasks from a sheet", func() {
      w := tape.GetWithClaims("/api/v1/courses/1/sheets/1/tasks", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      tasks_expected, err := stores.Task.TasksOfSheet(1)
      g.Assert(err).Equal(nil)

      tasks_actual := []TaskResponse{}
      err = json.NewDecoder(w.Body).Decode(&tasks_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(tasks_actual)).Equal(3)

      for k := range tasks_actual {
        g.Assert(tasks_expected[k].ID).Equal(tasks_actual[k].ID)
        g.Assert(tasks_expected[k].MaxPoints).Equal(tasks_actual[k].MaxPoints)
        g.Assert(tasks_expected[k].Name).Equal(tasks_actual[k].Name)
        g.Assert(tasks_expected[k].PublicDockerImage).Equal(tasks_actual[k].PublicDockerImage)
        g.Assert(tasks_expected[k].PrivateDockerImage).Equal(tasks_actual[k].PrivateDockerImage)
      }
    })

    g.It("Should get a specific task", func() {

      task_expected, err := stores.Task.Get(1)
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/courses/1/tasks/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_actual := &TaskResponse{}
      err = json.NewDecoder(w.Body).Decode(task_actual)
      g.Assert(err).Equal(nil)

      g.Assert(task_actual.ID).Equal(task_expected.ID)
      g.Assert(task_actual.MaxPoints).Equal(task_expected.MaxPoints)
      g.Assert(task_actual.Name).Equal(task_expected.Name)
      g.Assert(task_actual.PublicDockerImage).Equal(task_expected.PublicDockerImage)
      g.Assert(task_actual.PrivateDockerImage).Equal(task_expected.PrivateDockerImage)

    })

    g.It("Creating should require claims", func() {
      w := tape.Post("/api/v1/courses/1/sheets/1/tasks", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.Xit("Creating should require body", func() {
      // TODO empty request with claims
    })

    g.It("Should create valid task", func() {
      tasks_before, err := stores.Task.TasksOfSheet(1)
      g.Assert(err).Equal(nil)

      task_sent := TaskRequest{
        Name:               "new Task",
        MaxPoints:          88,
        PublicDockerImage:  "TestImage-Public",
        PrivateDockerImage: "TestImage-Private",
      }

      err = task_sent.Validate()
      g.Assert(err).Equal(nil)

      w := tape.PostWithClaims("/api/v1/courses/1/sheets/1/tasks", helper.ToH(task_sent), 1, true)
      g.Assert(w.Code).Equal(http.StatusCreated)

      task_return := &TaskResponse{}
      err = json.NewDecoder(w.Body).Decode(&task_return)
      g.Assert(task_return.Name).Equal("new Task")
      g.Assert(task_return.MaxPoints).Equal(88)
      g.Assert(task_return.PrivateDockerImage).Equal(task_sent.PrivateDockerImage)
      g.Assert(task_return.PublicDockerImage).Equal(task_sent.PublicDockerImage)

      tasks_after, err := stores.Task.TasksOfSheet(1)
      g.Assert(err).Equal(nil)
      g.Assert(len(tasks_after)).Equal(len(tasks_before) + 1)
    })

    g.It("Should skip non-existent test files", func() {
      w := tape.GetWithClaims("/api/v1/courses/1/tasks/1/public_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusNotFound)

      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/private_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusNotFound)
    })

    g.It("Should upload public test file", func() {
      defer helper.NewPublicTestFileHandle(1).Delete()
      defer helper.NewPrivateTestFileHandle(1).Delete()

      // no files so far
      g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(false)
      g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(false)

      w := tape.GetWithClaims("/api/v1/courses/1/tasks/1/public_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusNotFound)

      // public test
      filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
      w, err := tape.UploadWithClaims("/api/v1/courses/1/tasks/1/public_file", filename, "application/zip", 1, true)
      g.Assert(err).Equal(nil)
      g.Assert(w.Code).Equal(http.StatusOK)

      // only public file
      g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(true)
      g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(false)

      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/public_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/public_file", 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/public_file", 122, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)
    })

    g.It("Should upload private test file", func() {
      defer helper.NewPublicTestFileHandle(1).Delete()
      defer helper.NewPrivateTestFileHandle(1).Delete()

      // no files so far
      g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(false)
      g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(false)

      w := tape.GetWithClaims("/api/v1/courses/1/tasks/1/private_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusNotFound)

      // public test
      filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
      w, err := tape.UploadWithClaims("/api/v1/courses/1/tasks/1/private_file", filename, "application/zip", 1, true)
      g.Assert(err).Equal(nil)
      g.Assert(w.Code).Equal(http.StatusOK)

      // only public file
      g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(false)
      g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(true)

      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/private_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/private_file", 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/private_file", 122, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)
    })

    g.It("Changes should require claims", func() {
      w := tape.Put("/api/v1/courses/1/sheets/1/tasks", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.It("Should perform updates", func() {
      data := H{
        "max_points":           555,
        "name":                 "new blub",
        "public_docker_image":  "new_public",
        "private_docker_image": "new_private",
      }

      w := tape.PutWithClaims("/api/v1/courses/1/tasks/1", data, 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_after, err := stores.Task.Get(1)
      g.Assert(err).Equal(nil)
      g.Assert(task_after.MaxPoints).Equal(555)
      g.Assert(task_after.Name).Equal("new blub")
      g.Assert(task_after.PublicDockerImage).Equal("new_public")
      g.Assert(task_after.PrivateDockerImage).Equal("new_private")

      w = tape.PutWithClaims("/api/v1/courses/1/tasks/1", data, 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      w = tape.PutWithClaims("/api/v1/courses/1/tasks/1", data, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)
    })

    g.It("Should delete when valid access claims", func() {

      entries_before, err := stores.Task.GetAll()
      g.Assert(err).Equal(nil)

      w := tape.Delete("/api/v1/courses/1/tasks/1")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.DeleteWithClaims("/api/v1/courses/1/tasks/1", 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      w = tape.DeleteWithClaims("/api/v1/courses/1/tasks/1", 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // verify nothing has changes
      entries_after, err := stores.Task.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before))

      w = tape.DeleteWithClaims("/api/v1/courses/1/tasks/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      // verify a sheet less exists
      entries_after, err = stores.Task.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before) - 1)

    })

    g.It("students should see public results", func() {
      w := tape.Get("/api/v1/courses/1/tasks/1/result")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/result", 1, false)
      g.Assert(w.Code).Equal(http.StatusBadRequest)

      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/result", 2, false)
      g.Assert(w.Code).Equal(http.StatusBadRequest)

      w = tape.GetWithClaims("/api/v1/courses/1/tasks/1/result", 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      actual := &GradeResponse{}
      err := json.NewDecoder(w.Body).Decode(actual)
      g.Assert(err).Equal(nil)
      g.Assert(actual.PrivateTestLog).Equal("")
      g.Assert(actual.PrivateTestStatus).Equal(-1)

    })

    g.It("Permission test", func() {
      // sheet (id=1) belongs to group(id=1)
      url := "/api/v1/courses/1/sheets/1/tasks"

      // global root can do whatever they want
      w := tape.GetWithClaims(url, 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      // enrolled tutors can access
      w = tape.GetWithClaims(url, 2, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // enrolled students can access
      w = tape.GetWithClaims(url, 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // disenroll student
      w = tape.DeleteWithClaims("/api/v1/courses/1/enrollments", 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // cannot access anymore
      w = tape.GetWithClaims(url, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)
    })

    g.It("Should get all missing tasks", func() {
      // in the mock script each student has a submission to all tasks
      // therefore we need to delete it temporarily
      _, err := tape.DB.Exec("DELETE FROM submissions WHERE task_id = 2")
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/courses/1/tasks/missing", 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      result := []MissingTaskResponse{}
      err = json.NewDecoder(w.Body).Decode(&result)
      g.Assert(err).Equal(nil)
      for _, el := range result {
        g.Assert(el.Task.ID).Equal(int64(2))
      }

    })

    g.AfterEach(func() {
      tape.AfterEach()
    })

  })

}
