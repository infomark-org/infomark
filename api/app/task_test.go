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
	"context"
	"encoding/json"
	"fmt"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/logging"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/franela/goblin"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"

	// "github.com/spf13/viper"
	"net/http"
	"testing"
)

func SetTaskContext(task *model.Task) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "task", task)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TestTask(t *testing.T) {

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	stores := NewStores(db)
	rs := NewTaskResource(stores)

	g.Describe("Task Query", func() {

		sheet_active, err := stores.Sheet.Get(1)
		g.Assert(err).Equal(nil)

		g.It("Should require claims", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data:   helper.H{},
					Method: "GET",
				},
				rs.IndexHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should list all tasks from a sheet", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.IndexHandler,
				authenticate.RequiredValidAccessClaims,
				SetSheetContext(sheet_active),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			tasks_actual := []model.Task{}

			err = json.NewDecoder(w.Body).Decode(&tasks_actual)
			g.Assert(err).Equal(nil)
			g.Assert(len(tasks_actual)).Equal(3)

		})

		g.It("Should get a specific sheet", func() {

			task_expected, err := stores.Task.Get(1)
			g.Assert(err).Equal(nil)

			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.GetHandler,
				// set course
				authenticate.RequiredValidAccessClaims,
				SetTaskContext(task_expected),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			task_actual := &model.Task{}
			err = json.NewDecoder(w.Body).Decode(task_actual)
			g.Assert(err).Equal(nil)

			g.Assert(task_actual.ID).Equal(task_expected.ID)
			g.Assert(task_actual.MaxPoints).Equal(task_expected.MaxPoints)
			g.Assert(task_actual.PublicDockerImage).Equal(task_expected.PublicDockerImage)
			g.Assert(task_actual.PrivateDockerImage).Equal(task_expected.PrivateDockerImage)

		})

	})

}

func TestTaskCreation(t *testing.T) {

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	stores := NewStores(db)
	rs := NewTaskResource(stores)

	// delete fixture file
	defer helper.NewPublicTestFileHandle(1).Delete()
	defer helper.NewPrivateTestFileHandle(1).Delete()

	g.Describe("Task Creation", func() {

		sheet_active, err := stores.Sheet.Get(1)
		g.Assert(err).Equal(nil)

		g.It("Should require claims", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data:   helper.H{},
					Method: "POST",
				},
				rs.CreateHandler,
				authenticate.RequiredValidAccessClaims,
				SetSheetContext(sheet_active),
			)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should require body", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "POST",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.CreateHandler,
				authenticate.RequiredValidAccessClaims,
				SetSheetContext(sheet_active),
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should create valid task", func() {

			tasks_before, err := rs.Stores.Task.TasksOfSheet(sheet_active, false)
			g.Assert(err).Equal(nil)

			task_sent := model.Task{
				MaxPoints:          88,
				PublicDockerImage:  "TestImage-Public",
				PrivateDockerImage: "TestImage-Private",
			}

			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.ToH(task_sent),
					Method:       "POST",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.CreateHandler,
				authenticate.RequiredValidAccessClaims,
				SetSheetContext(sheet_active),
			)
			g.Assert(w.Code).Equal(http.StatusCreated)

			task_return := &model.Task{}
			err = json.NewDecoder(w.Body).Decode(&task_return)

			g.Assert(task_return.MaxPoints).Equal(88)
			g.Assert(task_return.PrivateDockerImage).Equal(task_sent.PrivateDockerImage)
			g.Assert(task_return.PublicDockerImage).Equal(task_sent.PublicDockerImage)

			tasks_after, err := rs.Stores.Task.TasksOfSheet(sheet_active, false)
			g.Assert(err).Equal(nil)
			g.Assert(len(tasks_after)).Equal(len(tasks_before) + 1)
		})

		g.It("Should skip non-existent test files", func() {

			task_active, err := stores.Task.Get(1)
			g.Assert(err).Equal(nil)

			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.GetPublicTestFileHandler,
				authenticate.RequiredValidAccessClaims,
				SetTaskContext(task_active),
			)

			g.Assert(w.Code).Equal(http.StatusNotFound)

			w = helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.GetPrivateTestFileHandler,
				authenticate.RequiredValidAccessClaims,
				SetTaskContext(task_active),
			)

			g.Assert(w.Code).Equal(http.StatusNotFound)
		})

		g.It("Should upload public test file", func() {

			task_active, err := stores.Task.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(helper.NewPublicTestFileHandle(task_active.ID).Exists()).Equal(false)
			g.Assert(helper.NewPrivateTestFileHandle(task_active.ID).Exists()).Equal(false)

			hnd := helper.NewPublicTestFileHandle(task_active.ID)
			g.Assert(hnd.Exists()).Equal(false)

			// send public testfile
			w := helper.SimulateFileRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "POST",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir")),
				"file_data",
				rs.ChangePublicTestFileHandler,
				authenticate.RequiredValidAccessClaims,
				SetTaskContext(task_active),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			// public test file should be there
			g.Assert(helper.NewPublicTestFileHandle(task_active.ID).Exists()).Equal(true)

			// check request as well
			w = helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.GetPublicTestFileHandler,
				authenticate.RequiredValidAccessClaims,
				SetTaskContext(task_active),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			// private test file should be not there
			g.Assert(helper.NewPrivateTestFileHandle(task_active.ID).Exists()).Equal(false)

			// check request as well
			w = helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.GetPrivateTestFileHandler,
				authenticate.RequiredValidAccessClaims,
				SetTaskContext(task_active),
			)
			g.Assert(w.Code).Equal(http.StatusNotFound)

			// send private testfile
			w = helper.SimulateFileRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "POST",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir")),
				"file_data",
				rs.ChangePrivateTestFileHandler,
				authenticate.RequiredValidAccessClaims,
				SetTaskContext(task_active),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			g.Assert(helper.NewPublicTestFileHandle(task_active.ID).Exists()).Equal(true)
			g.Assert(helper.NewPrivateTestFileHandle(task_active.ID).Exists()).Equal(true)

			// check requests as well
			w = helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.GetPrivateTestFileHandler,
				authenticate.RequiredValidAccessClaims,
				SetTaskContext(task_active),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			w = helper.SimulateFileRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "POST",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir")),
				"file_data",
				rs.ChangePrivateTestFileHandler,
				authenticate.RequiredValidAccessClaims,
				SetTaskContext(task_active),
			)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

	})

}

func TestTaskChanges(t *testing.T) {

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	stores := NewStores(db)
	rs := NewTaskResource(stores)

	g.Describe("Task Changes", func() {

		sheet_active, err := stores.Sheet.Get(1)
		g.Assert(err).Equal(nil)

		all_tasks_before, err := rs.Stores.Task.GetAll()
		g.Assert(err).Equal(nil)

		tasks_before, err := rs.Stores.Task.TasksOfSheet(sheet_active, false)
		g.Assert(err).Equal(nil)

		task_before, err := stores.Task.Get(tasks_before[0].ID)
		g.Assert(err).Equal(nil)

		g.It("Should require claims", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data:   helper.H{},
					Method: "PATCH",
				},
				rs.EditHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should perform updates", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{

						"max_points":           555,
						"public_docker_image":  "new_public",
						"private_docker_image": "new_private",
					},
					Method:       "PATCH",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.EditHandler,
				authenticate.RequiredValidAccessClaims,
				SetTaskContext(task_before),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			task_after, err := stores.Task.Get(tasks_before[0].ID)
			g.Assert(err).Equal(nil)
			g.Assert(task_after.MaxPoints).Equal(555)
			g.Assert(task_after.PublicDockerImage).Equal("new_public")
			g.Assert(task_after.PrivateDockerImage).Equal("new_private")
		})

		g.It("Should delete", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "DELETE",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.DeleteHandler,
				authenticate.RequiredValidAccessClaims,
				SetTaskContext(task_before),
			)
			// TODO()
			g.Assert(w.Code).Equal(http.StatusOK)

			all_tasks_after, err := rs.Stores.Task.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(all_tasks_after)).Equal(len(all_tasks_before) - 1)
		})
	})

}
