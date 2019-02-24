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
	_ "fmt"
	"time"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/cgtuebingen/infomark-backend/logging"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/franela/goblin"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"

	// "github.com/spf13/viper"
	"net/http"
	"testing"
)

func SetSheetContext(sheet *model.Sheet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "sheet", sheet)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TestSheet(t *testing.T) {

	email.DefaultMail = email.VoidMail

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	stores := NewStores(db)
	rs := NewSheetResource(stores)

	g.Describe("Sheet Query", func() {

		course_active, err := stores.Course.Get(1)
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

		g.It("Should list all sheets from a course", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.IndexHandler,
				authenticate.RequiredValidAccessClaims,
				SetCourseContext(course_active),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			sheets_actual := []model.Course{}

			err = json.NewDecoder(w.Body).Decode(&sheets_actual)
			g.Assert(err).Equal(nil)
			g.Assert(len(sheets_actual)).Equal(10)

		})

		g.It("Should get a specific sheet", func() {

			sheet_expected, err := stores.Sheet.Get(1)
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
				SetSheetContext(sheet_expected),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			sheet_actual := &model.Sheet{}
			err = json.NewDecoder(w.Body).Decode(sheet_actual)
			g.Assert(err).Equal(nil)

			g.Assert(sheet_actual.ID).Equal(sheet_expected.ID)
			g.Assert(sheet_actual.Name).Equal(sheet_expected.Name)
			g.Assert(sheet_actual.PublishAt.Equal(sheet_expected.PublishAt)).Equal(true)
			g.Assert(sheet_actual.DueAt.Equal(sheet_expected.DueAt)).Equal(true)

		})

	})

}

func TestSheetCreation(t *testing.T) {

	email.DefaultMail = email.VoidMail

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	stores := NewStores(db)
	rs := NewSheetResource(stores)

	// delete fixture file
	defer helper.NewSheetFileHandle(1).Delete()

	g.Describe("Sheet Creation", func() {

		course_active, err := stores.Course.Get(1)
		g.Assert(err).Equal(nil)

		g.It("Should require claims", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data:   helper.H{},
					Method: "POST",
				},
				rs.CreateHandler,
				authenticate.RequiredValidAccessClaims,
				SetCourseContext(course_active),
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
				SetCourseContext(course_active),
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should create valid sheet", func() {

			sheets_before, err := rs.Stores.Sheet.SheetsOfCourse(course_active, false)
			g.Assert(err).Equal(nil)

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"name":       "Sheet_new",
						"publish_at": "2019-02-01T01:02:03Z",
						"due_at":     "2019-07-30T23:59:59Z",
					},
					Method:       "POST",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.CreateHandler,
				authenticate.RequiredValidAccessClaims,
				SetCourseContext(course_active),
			)
			g.Assert(w.Code).Equal(http.StatusCreated)

			expectedPublishAt, _ := time.Parse(time.RFC3339, "2019-02-01T01:02:03Z")
			expectedDueAt, _ := time.Parse(time.RFC3339, "2019-07-30T23:59:59Z")

			sheet_return := &model.Sheet{}
			err = json.NewDecoder(w.Body).Decode(&sheet_return)
			g.Assert(sheet_return.Name).Equal("Sheet_new")
			g.Assert(sheet_return.PublishAt.Equal(expectedPublishAt)).Equal(true)
			g.Assert(sheet_return.DueAt.Equal(expectedDueAt)).Equal(true)

			sheets_after, err := rs.Stores.Sheet.SheetsOfCourse(course_active, false)
			g.Assert(err).Equal(nil)
			g.Assert(len(sheets_after)).Equal(len(sheets_before) + 1)
		})

		g.It("Should skip non-existent sheet file", func() {

			sheet_active, err := stores.Sheet.Get(1)
			g.Assert(err).Equal(nil)

			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.GetFileHandler,
				authenticate.RequiredValidAccessClaims,
				SetSheetContext(sheet_active),
			)
			g.Assert(w.Code).Equal(http.StatusNotFound)
		})

		g.It("Should upload sheet file", func() {

			sheet_active, err := stores.Sheet.Get(1)
			g.Assert(err).Equal(nil)

			hnd := helper.NewSheetFileHandle(sheet_active.ID)
			g.Assert(hnd.Exists()).Equal(false)

			w := helper.SimulateFileRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "POST",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir")),
				"file_data",
				rs.ChangeFileHandler,
				authenticate.RequiredValidAccessClaims,
				SetSheetContext(sheet_active),
			)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

	})

}

func TestSheetChanges(t *testing.T) {

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	stores := NewStores(db)
	rs := NewSheetResource(stores)

	g.Describe("Sheet Changes", func() {

		course_active, err := stores.Course.Get(1)
		g.Assert(err).Equal(nil)

		all_sheets_before, err := rs.Stores.Sheet.GetAll()
		g.Assert(err).Equal(nil)

		sheets_before, err := rs.Stores.Sheet.SheetsOfCourse(course_active, false)
		g.Assert(err).Equal(nil)

		sheet_before, err := stores.Sheet.Get(sheets_before[0].ID)
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
						"name":       "Sheet_update",
						"publish_at": "2023-02-01T01:02:03Z",
						"due_at":     "2023-07-30T23:59:59Z",
					},
					Method:       "PATCH",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.EditHandler,
				authenticate.RequiredValidAccessClaims,
				SetSheetContext(sheet_before),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			expectedBegin, _ := time.Parse(time.RFC3339, "2023-02-01T01:02:03Z")
			expectedEnd, _ := time.Parse(time.RFC3339, "2023-07-30T23:59:59Z")

			sheet_after, err := stores.Sheet.Get(sheets_before[0].ID)
			g.Assert(err).Equal(nil)
			g.Assert(sheet_after.Name).Equal("Sheet_update")
			g.Assert(sheet_after.PublishAt.Equal(expectedBegin)).Equal(true)
			g.Assert(sheet_after.DueAt.Equal(expectedEnd)).Equal(true)
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
				SetSheetContext(sheet_before),
			)
			// TODO()
			g.Assert(w.Code).Equal(http.StatusOK)

			all_sheets_after, err := rs.Stores.Sheet.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(all_sheets_after)).Equal(len(all_sheets_before) - 1)
		})
	})

}
