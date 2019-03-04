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
  "github.com/cgtuebingen/infomark-backend/database"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/cgtuebingen/infomark-backend/model"
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

      w := tape.Get("/api/v1/sheets/1/tasks")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/sheets/1/tasks", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Should list all tasks from a sheet", func() {
      w := tape.GetWithClaims("/api/v1/sheets/1/tasks", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      tasks_actual := []model.Task{}
      err := json.NewDecoder(w.Body).Decode(&tasks_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(tasks_actual)).Equal(3)
    })

    g.It("Should get a specific sheet", func() {

      task_expected, err := stores.Task.Get(1)
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/tasks/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_actual := &model.Task{}
      err = json.NewDecoder(w.Body).Decode(task_actual)
      g.Assert(err).Equal(nil)

      g.Assert(task_actual.ID).Equal(task_expected.ID)
      g.Assert(task_actual.MaxPoints).Equal(task_expected.MaxPoints)
      g.Assert(task_actual.PublicDockerImage).Equal(task_expected.PublicDockerImage)
      g.Assert(task_actual.PrivateDockerImage).Equal(task_expected.PrivateDockerImage)

    })

    g.It("Creating should require claims", func() {
      w := tape.Post("/api/v1/sheets/1/tasks", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.Xit("Creating should require body", func() {
      // TODO empty request with claims
    })

    g.It("Should create valid task", func() {
      tasks_before, err := stores.Task.TasksOfSheet(1, false)
      g.Assert(err).Equal(nil)

      task_sent := model.Task{
        MaxPoints:          88,
        PublicDockerImage:  "TestImage-Public",
        PrivateDockerImage: "TestImage-Private",
      }

      err = task_sent.Validate()
      g.Assert(err).Equal(nil)

      w := tape.PostWithClaims("/api/v1/sheets/1/tasks", helper.ToH(task_sent), 1, true)
      g.Assert(w.Code).Equal(http.StatusCreated)

      task_return := &model.Task{}
      err = json.NewDecoder(w.Body).Decode(&task_return)
      g.Assert(task_return.MaxPoints).Equal(88)
      g.Assert(task_return.PrivateDockerImage).Equal(task_sent.PrivateDockerImage)
      g.Assert(task_return.PublicDockerImage).Equal(task_sent.PublicDockerImage)

      tasks_after, err := stores.Task.TasksOfSheet(1, false)
      g.Assert(err).Equal(nil)
      g.Assert(len(tasks_after)).Equal(len(tasks_before) + 1)
    })

    g.It("Should skip non-existent test files", func() {
      w := tape.GetWithClaims("/api/v1/tasks/1/public_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusNotFound)

      w = tape.GetWithClaims("/api/v1/tasks/1/private_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusNotFound)
    })

    g.It("Should upload public test file", func() {
      defer helper.NewPublicTestFileHandle(1).Delete()
      defer helper.NewPrivateTestFileHandle(1).Delete()

      // no files so far
      g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(false)
      g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(false)

      // public test
      filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
      w, err := tape.UploadWithClaims("/api/v1/tasks/1/public_file", filename, "application/zip", 1, true)
      g.Assert(err).Equal(nil)
      g.Assert(w.Code).Equal(http.StatusOK)

      // only public file
      g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(true)
      g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(false)

      w = tape.GetWithClaims("/api/v1/tasks/1/public_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      w = tape.GetWithClaims("/api/v1/tasks/1/private_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusNotFound)
    })

    g.It("Should upload private test file", func() {
      defer helper.NewPublicTestFileHandle(1).Delete()
      defer helper.NewPrivateTestFileHandle(1).Delete()

      // no files so far
      g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(false)
      g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(false)

      // public test
      filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
      w, err := tape.UploadWithClaims("/api/v1/tasks/1/private_file", filename, "application/zip", 1, true)
      g.Assert(err).Equal(nil)
      g.Assert(w.Code).Equal(http.StatusOK)

      // only public file
      g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(false)
      g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(true)

      w = tape.GetWithClaims("/api/v1/tasks/1/public_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusNotFound)

      w = tape.GetWithClaims("/api/v1/tasks/1/private_file", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Changes should require claims", func() {
      w := tape.Put("/api/v1/sheets/1/tasks", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.It("Should perform updates", func() {
      data := H{
        "max_points":           555,
        "public_docker_image":  "new_public",
        "private_docker_image": "new_private",
      }

      w := tape.PutWithClaims("/api/v1/tasks/1", data, 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_after, err := stores.Task.Get(1)
      g.Assert(err).Equal(nil)
      g.Assert(task_after.MaxPoints).Equal(555)
      g.Assert(task_after.PublicDockerImage).Equal("new_public")
      g.Assert(task_after.PrivateDockerImage).Equal("new_private")
    })

    g.It("Should delete when valid access claims", func() {
      entries_before, err := stores.Task.GetAll()
      g.Assert(err).Equal(nil)

      w := tape.Delete("/api/v1/tasks/1")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // verify nothing has changes
      entries_after, err := stores.Task.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before))

      w = tape.DeleteWithClaims("/api/v1/tasks/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      // verify a sheet less exists
      entries_after, err = stores.Task.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before) - 1)
    })

    g.It("Permission test", func() {
      // sheet (id=1) belongs to group(id=1)
      url := "/api/v1/sheets/1/tasks"

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

    g.It("Should get own rating", func() {
      userID := int64(112)
      taskID := int64(1)

      givenRating, err := stores.Task.GetRatingOfTaskByUser(taskID, userID)
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/tasks/1/ratings", userID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_rating_actual := &TaskRatingResponse{}
      err = json.NewDecoder(w.Body).Decode(task_rating_actual)
      g.Assert(err).Equal(nil)

      g.Assert(task_rating_actual.OwnRating).Equal(givenRating.Rating)
      g.Assert(task_rating_actual.TaskID).Equal(taskID)

      // update rating (mock had rating 2)
      w = tape.PostWithClaims("/api/v1/tasks/1/ratings", H{"rating": 4}, userID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // new query
      w = tape.GetWithClaims("/api/v1/tasks/1/ratings", userID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_rating_actual2 := &TaskRatingResponse{}
      err = json.NewDecoder(w.Body).Decode(task_rating_actual2)
      g.Assert(err).Equal(nil)

      g.Assert(task_rating_actual2.OwnRating).Equal(4)
      g.Assert(task_rating_actual2.TaskID).Equal(taskID)
    })

    g.It("Should create own rating", func() {
      userID := int64(112)
      taskID := int64(1)

      // delete and create (see mock.py)
      prevRatingModel, err := stores.Task.GetRatingOfTaskByUser(taskID, userID)
      g.Assert(err).Equal(nil)
      database.Delete(tape.DB, "task_ratings", prevRatingModel.ID)

      w := tape.GetWithClaims("/api/v1/tasks/1/ratings", userID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_rating_actual3 := &TaskRatingResponse{}
      err = json.NewDecoder(w.Body).Decode(task_rating_actual3)
      g.Assert(err).Equal(nil)

      g.Assert(task_rating_actual3.OwnRating).Equal(0)
      g.Assert(task_rating_actual3.TaskID).Equal(taskID)

      // update rating (mock had rating 2)
      w = tape.PostWithClaims("/api/v1/tasks/1/ratings", H{"rating": 4}, userID, false)
      g.Assert(w.Code).Equal(http.StatusCreated)

      // new query
      w = tape.GetWithClaims("/api/v1/tasks/1/ratings", userID, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      task_rating_actual2 := &TaskRatingResponse{}
      err = json.NewDecoder(w.Body).Decode(task_rating_actual2)
      g.Assert(err).Equal(nil)

      g.Assert(task_rating_actual2.OwnRating).Equal(4)
      g.Assert(task_rating_actual2.TaskID).Equal(taskID)
    })

    g.AfterEach(func() {
      tape.AfterEach()
    })

  })

}
