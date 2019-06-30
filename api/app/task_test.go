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
	PrepareTests()
	g := goblin.Goblin(t)
	email.DefaultMail = email.VoidMail

	tape := &Tape{}

	var stores *Stores

	studentJWT := NewJWTRequest(112, false)
	tutorJWT := NewJWTRequest(2, false)
	adminJWT := NewJWTRequest(1, true)
	noAdminJWT := NewJWTRequest(1, false)

	g.Describe("Task", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
			stores = NewStores(tape.DB)
		})

		g.It("Query should require access claims", func() {

			w := tape.Get("/api/v1/courses/1/sheets/1/tasks")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/courses/1/sheets/1/tasks", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Should list all tasks from a sheet", func() {
			w := tape.Get("/api/v1/courses/1/sheets/1/tasks", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			tasksExpected, err := stores.Task.TasksOfSheet(1)
			g.Assert(err).Equal(nil)

			tasksActual := []TaskResponse{}
			err = json.NewDecoder(w.Body).Decode(&tasksActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(tasksActual)).Equal(3)

			for k := range tasksActual {
				g.Assert(tasksExpected[k].ID).Equal(tasksActual[k].ID)
				g.Assert(tasksExpected[k].MaxPoints).Equal(tasksActual[k].MaxPoints)
				g.Assert(tasksExpected[k].Name).Equal(tasksActual[k].Name)
				g.Assert(tasksExpected[k].PublicDockerImage.String).Equal(tasksActual[k].PublicDockerImage.String)
				g.Assert(tasksExpected[k].PrivateDockerImage.String).Equal(tasksActual[k].PrivateDockerImage.String)

				g.Assert(tasksExpected[k].PublicDockerImage.Valid).Equal(true)
				g.Assert(tasksExpected[k].PrivateDockerImage.Valid).Equal(true)
			}
		})

		g.It("Should get a specific task", func() {

			taskExpected, err := stores.Task.Get(1)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/tasks/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			taskActual := &TaskResponse{}
			err = json.NewDecoder(w.Body).Decode(taskActual)
			g.Assert(err).Equal(nil)

			g.Assert(taskActual.ID).Equal(taskExpected.ID)
			g.Assert(taskActual.MaxPoints).Equal(taskExpected.MaxPoints)
			g.Assert(taskActual.Name).Equal(taskExpected.Name)
			g.Assert(taskActual.PublicDockerImage.String).Equal(taskExpected.PublicDockerImage.String)
			g.Assert(taskActual.PrivateDockerImage.String).Equal(taskExpected.PrivateDockerImage.String)

			g.Assert(taskActual.PublicDockerImage.Valid).Equal(true)
			g.Assert(taskActual.PrivateDockerImage.Valid).Equal(true)

		})

		g.It("Creating should require claims", func() {
			w := tape.Post("/api/v1/courses/1/sheets/1/tasks", H{})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.Xit("Creating should require body", func() {
			// TODO empty request with claims
		})

		g.It("Should create valid task", func() {
			tasksBefore, err := stores.Task.TasksOfSheet(1)
			g.Assert(err).Equal(nil)

			taskSent := TaskRequest{
				Name:               "new Task",
				MaxPoints:          88,
				PublicDockerImage:  "testimage_public",
				PrivateDockerImage: "testimage_private",
			}

			err = taskSent.Validate()
			g.Assert(err).Equal(nil)

			w := tape.Post("/api/v1/courses/1/sheets/1/tasks", helper.ToH(taskSent), adminJWT)
			g.Assert(w.Code).Equal(http.StatusCreated)

			taskReturn := &TaskResponse{}
			err = json.NewDecoder(w.Body).Decode(&taskReturn)
			g.Assert(err).Equal(nil)
			g.Assert(taskReturn.Name).Equal("new Task")
			g.Assert(taskReturn.MaxPoints).Equal(88)
			g.Assert(taskReturn.PrivateDockerImage.Valid).Equal(true)
			g.Assert(taskReturn.PrivateDockerImage.String).Equal(taskSent.PrivateDockerImage)
			g.Assert(taskReturn.PublicDockerImage.Valid).Equal(true)
			g.Assert(taskReturn.PublicDockerImage.String).Equal(taskSent.PublicDockerImage)

			tasksAfter, err := stores.Task.TasksOfSheet(1)
			g.Assert(err).Equal(nil)
			g.Assert(len(tasksAfter)).Equal(len(tasksBefore) + 1)
		})

		g.It("Should skip non-existent test files", func() {
			w := tape.Get("/api/v1/courses/1/tasks/1/public_file", adminJWT)
			g.Assert(w.Code).Equal(http.StatusNotFound)

			w = tape.Get("/api/v1/courses/1/tasks/1/private_file", adminJWT)
			g.Assert(w.Code).Equal(http.StatusNotFound)
		})

		g.It("Should upload public test file", func() {
			defer helper.NewPublicTestFileHandle(1).Delete()
			defer helper.NewPrivateTestFileHandle(1).Delete()

			// no files so far
			g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(false)
			g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(false)

			w := tape.Get("/api/v1/courses/1/tasks/1/public_file", adminJWT)
			g.Assert(w.Code).Equal(http.StatusNotFound)

			// public test
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			w, err := tape.Upload("/api/v1/courses/1/tasks/1/public_file", filename, "application/zip", adminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)

			// only public file
			g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(true)
			g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(false)

			w = tape.Get("/api/v1/courses/1/tasks/1/public_file", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			w = tape.Get("/api/v1/courses/1/tasks/1/public_file", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Get("/api/v1/courses/1/tasks/1/public_file", studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)
		})

		g.It("Should upload private test file", func() {
			defer helper.NewPublicTestFileHandle(1).Delete()
			defer helper.NewPrivateTestFileHandle(1).Delete()

			// no files so far
			g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(false)
			g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(false)

			w := tape.Get("/api/v1/courses/1/tasks/1/private_file", adminJWT)
			g.Assert(w.Code).Equal(http.StatusNotFound)

			// public test
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			w, err := tape.Upload("/api/v1/courses/1/tasks/1/private_file", filename, "application/zip", adminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)

			// only public file
			g.Assert(helper.NewPublicTestFileHandle(1).Exists()).Equal(false)
			g.Assert(helper.NewPrivateTestFileHandle(1).Exists()).Equal(true)

			w = tape.Get("/api/v1/courses/1/tasks/1/private_file", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			w = tape.Get("/api/v1/courses/1/tasks/1/private_file", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Get("/api/v1/courses/1/tasks/1/private_file", studentJWT)
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

			w := tape.Put("/api/v1/courses/1/tasks/1", data, adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			taskAfter, err := stores.Task.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(taskAfter.MaxPoints).Equal(555)
			g.Assert(taskAfter.Name).Equal("new blub")
			g.Assert(taskAfter.PublicDockerImage.Valid).Equal(true)
			g.Assert(taskAfter.PublicDockerImage.String).Equal("new_public")
			g.Assert(taskAfter.PrivateDockerImage.Valid).Equal(true)
			g.Assert(taskAfter.PrivateDockerImage.String).Equal("new_private")

			w = tape.Put("/api/v1/courses/1/tasks/1", data, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Put("/api/v1/courses/1/tasks/1", data, studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)
		})

		g.It("Should delete when valid access claims", func() {

			entriesBefore, err := stores.Task.GetAll()
			g.Assert(err).Equal(nil)

			w := tape.Delete("/api/v1/courses/1/tasks/1")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Delete("/api/v1/courses/1/tasks/1", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Delete("/api/v1/courses/1/tasks/1", studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// verify nothing has changes
			entriesAfter, err := stores.Task.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore))

			w = tape.Delete("/api/v1/courses/1/tasks/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// verify a sheet less exists
			entriesAfter, err = stores.Task.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore) - 1)

		})

		g.It("students should see public results", func() {
			w := tape.Get("/api/v1/courses/1/tasks/1/result")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/courses/1/tasks/1/result", noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			w = tape.Get("/api/v1/courses/1/tasks/1/result", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			w = tape.Get("/api/v1/courses/1/tasks/1/result", studentJWT)
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
			w := tape.Get(url, adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// enrolled tutors can access
			w = tape.Get(url, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// enrolled students can access
			w = tape.Get(url, studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// disenroll student
			w = tape.Delete("/api/v1/courses/1/enrollments", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// cannot access anymore
			w = tape.Get(url, studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)
		})

		g.It("Should get all missing tasks", func() {
			// in the mock script each student has a submission to all tasks
			// therefore we need to delete it temporarily
			_, err := tape.DB.Exec("DELETE FROM submissions WHERE task_id = 2")
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/tasks/missing", studentJWT)
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
