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
	"time"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/franela/goblin"
	"github.com/spf13/viper"
)

func TestSheet(t *testing.T) {
	PrepareTests()
	g := goblin.Goblin(t)
	email.DefaultMail = email.VoidMail

	tape := &Tape{}

	var stores *Stores

	studentJWT := NewJWTRequest(112, false)
	tutorJWT := NewJWTRequest(2, false)
	adminJWT := NewJWTRequest(1, true)
	noAdminJWT := NewJWTRequest(1, false)

	g.Describe("Sheet", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
			stores = NewStores(tape.DB)
			_ = stores
		})

		g.It("Query should require access claims", func() {

			w := tape.Get("/api/v1/courses/1/sheets")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/courses/1/sheets", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Should list all sheets a course", func() {

			w := tape.Get("/api/v1/courses/1/sheets", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			sheetsActual := []SheetResponse{}
			err := json.NewDecoder(w.Body).Decode(&sheetsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(sheetsActual)).Equal(10)
		})

		g.It("Should get a specific sheet", func() {
			sheetExpected, err := stores.Sheet.Get(1)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/sheets/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			sheetActual := &SheetResponse{}
			err = json.NewDecoder(w.Body).Decode(sheetActual)
			g.Assert(err).Equal(nil)

			g.Assert(sheetActual.ID).Equal(sheetExpected.ID)
			g.Assert(sheetActual.Name).Equal(sheetExpected.Name)
			g.Assert(sheetActual.PublishAt.Equal(sheetExpected.PublishAt)).Equal(true)
			g.Assert(sheetActual.DueAt.Equal(sheetExpected.DueAt)).Equal(true)
		})

		g.It("Creating a sheet should require access claims", func() {
			w := tape.Post("/api/v1/courses/1/sheets", H{})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Creating a sheet should require access body", func() {
			w := tape.Post("/api/v1/courses/1/sheets", H{}, adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should not create sheet with missing data", func() {
			data := H{
				"name":       "Sheet_new",
				"publish_at": "2019-02-01T01:02:03Z",
				// "due_at" is be missing
			}

			w := tape.Post("/api/v1/courses/1/sheets", data, adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should not create sheet with wrong times", func() {
			data := H{
				"name":       "Sheet_new",
				"publish_at": "2019-02-01T01:02:03Z",
				"due_at":     "2018-02-01T01:02:03Z", // time before publish
			}

			w := tape.Post("/api/v1/courses/1/sheets", data, adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should create valid sheet", func() {
			sheetsBefore, err := stores.Sheet.SheetsOfCourse(1)
			g.Assert(err).Equal(nil)

			sheetSent := SheetRequest{
				Name:      "Sheet_new",
				PublishAt: helper.Time(time.Now()),
				DueAt:     helper.Time(time.Now()),
			}

			// students
			w := tape.Post("/api/v1/courses/1/sheets", tape.ToH(sheetSent), studentJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Post("/api/v1/courses/1/sheets", tape.ToH(sheetSent), tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Post("/api/v1/courses/1/sheets", tape.ToH(sheetSent), noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusCreated)

			sheetReturn := &SheetResponse{}
			err = json.NewDecoder(w.Body).Decode(&sheetReturn)
			g.Assert(err).Equal(nil)
			g.Assert(sheetReturn.Name).Equal("Sheet_new")
			g.Assert(sheetReturn.PublishAt.Equal(sheetSent.PublishAt)).Equal(true)
			g.Assert(sheetReturn.DueAt.Equal(sheetSent.DueAt)).Equal(true)

			sheetsAfter, err := stores.Sheet.SheetsOfCourse(1)
			g.Assert(err).Equal(nil)
			g.Assert(len(sheetsAfter)).Equal(len(sheetsBefore) + 1)
		})

		g.It("Should skip non-existent sheet file", func() {
			w := tape.Get("/api/v1/courses/1/sheets/1/file", adminJWT)
			g.Assert(w.Code).Equal(http.StatusNotFound)
		})

		g.It("Should upload sheet file", func() {
			defer helper.NewSheetFileHandle(1).Delete()

			// no file so far
			g.Assert(helper.NewSheetFileHandle(1).Exists()).Equal(false)
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))

			// students
			w, err := tape.Upload("/api/v1/courses/1/sheets/1/file", filename, "application/zip", studentJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w, err = tape.Upload("/api/v1/courses/1/sheets/1/file", filename, "application/zip", tutorJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w, err = tape.Upload("/api/v1/courses/1/sheets/1/file", filename, "application/zip", noAdminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)

			// check disk
			g.Assert(helper.NewSheetFileHandle(1).Exists()).Equal(true)

			// a file should be now served
			w = tape.Get("/api/v1/courses/1/sheets/1/file", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Changes should require claims", func() {
			w := tape.Put("/api/v1/courses/1/sheets", H{})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should perform updates", func() {

			sheetSent := SheetRequest{
				Name:      "Sheet_update",
				PublishAt: helper.Time(time.Now()),
				DueAt:     helper.Time(time.Now()),
			}

			// students
			w := tape.Put("/api/v1/courses/1/sheets/1", tape.ToH(sheetSent), studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Put("/api/v1/courses/1/sheets/1", tape.ToH(sheetSent), tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Put("/api/v1/courses/1/sheets/1", tape.ToH(sheetSent), adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			sheetAfter, err := stores.Sheet.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(sheetAfter.Name).Equal("Sheet_update")
			g.Assert(sheetAfter.PublishAt.Equal(sheetSent.PublishAt)).Equal(true)
			g.Assert(sheetAfter.DueAt.Equal(sheetSent.DueAt)).Equal(true)
		})

		g.It("Should delete when valid access claims", func() {
			entriesBefore, err := stores.Sheet.GetAll()
			g.Assert(err).Equal(nil)

			w := tape.Delete("/api/v1/courses/1/sheets/1")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// students
			w = tape.Delete("/api/v1/courses/1/sheets/1", studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Delete("/api/v1/courses/1/sheets/1", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// verify nothing has changes
			entriesAfter, err := stores.Sheet.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore))

			// admin
			w = tape.Delete("/api/v1/courses/1/sheets/1", noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// verify a sheet less exists
			entriesAfter, err = stores.Sheet.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore) - 1)
		})

		g.It("Should see points for a sheet", func() {
			w := tape.Get("/api/v1/courses/1/sheets/1/points")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/courses/1/sheets/1/points", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

		})

		g.It("Permission test", func() {
			url := "/api/v1/courses/1/sheets"

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

		g.AfterEach(func() {
			tape.AfterEach()
		})
	})

}
