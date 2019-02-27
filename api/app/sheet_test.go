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
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/franela/goblin"
  "github.com/spf13/viper"
)

func TestSheet(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

  g.Describe("Sheet", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      stores = NewStores(tape.DB)
      _ = stores
    })

    g.It("Query should require access claims", func() {

      w := tape.Get("/api/v1/courses/1/sheets")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses/1/sheets", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Should list all sheets a course", func() {

      w := tape.GetWithClaims("/api/v1/courses/1/sheets", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      sheets_actual := []model.Sheet{}
      err := json.NewDecoder(w.Body).Decode(&sheets_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(sheets_actual)).Equal(10)
    })

    g.It("Should get a specific sheet", func() {

      sheet_expected, err := stores.Sheet.Get(1)
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/sheets/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      sheet_actual := &model.Sheet{}
      err = json.NewDecoder(w.Body).Decode(sheet_actual)
      g.Assert(err).Equal(nil)

      g.Assert(sheet_actual.ID).Equal(sheet_expected.ID)
      g.Assert(sheet_actual.Name).Equal(sheet_expected.Name)
      g.Assert(sheet_actual.PublishAt.Equal(sheet_expected.PublishAt)).Equal(true)
      g.Assert(sheet_actual.DueAt.Equal(sheet_expected.DueAt)).Equal(true)
    })

    g.It("Creating a sheet should require access claims", func() {
      w := tape.Post("/api/v1/courses/1/sheets", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.It("Creating a sheet should require access body", func() {
      w := tape.PlayDataWithClaims("POST", "/api/v1/courses/1/sheets", H{}, 1, true)
      g.Assert(w.Code).Equal(http.StatusBadRequest)
    })

    g.It("Should not create sheet with missing data", func() {
      data := H{
        "name":       "Sheet_new",
        "publish_at": "2019-02-01T01:02:03Z",
        // "due_at" is be missing
      }

      w := tape.PlayDataWithClaims("POST", "/api/v1/courses/1/sheets", data, 1, true)
      g.Assert(w.Code).Equal(http.StatusBadRequest)
    })

    g.It("Should not create sheet with wrong times", func() {
      data := H{
        "name":       "Sheet_new",
        "publish_at": "2019-02-01T01:02:03Z",
        "due_at":     "2018-02-01T01:02:03Z", // time before publish
      }

      w := tape.PlayDataWithClaims("POST", "/api/v1/courses/1/sheets", data, 1, true)
      g.Assert(w.Code).Equal(http.StatusBadRequest)
    })

    g.It("Should create valid sheet", func() {
      sheets_before, err := stores.Sheet.SheetsOfCourse(1, false)
      g.Assert(err).Equal(nil)

      sheet_sent := model.Sheet{
        Name:      "Sheet_new",
        PublishAt: helper.Time(time.Now()),
        DueAt:     helper.Time(time.Now()),
      }

      w := tape.PlayDataWithClaims("POST", "/api/v1/courses/1/sheets",
        tape.ToH(sheet_sent), 1, true)
      g.Assert(w.Code).Equal(http.StatusCreated)

      sheet_return := &model.Sheet{}
      err = json.NewDecoder(w.Body).Decode(&sheet_return)
      g.Assert(sheet_return.Name).Equal("Sheet_new")
      g.Assert(sheet_return.PublishAt.Equal(sheet_sent.PublishAt)).Equal(true)
      g.Assert(sheet_return.DueAt.Equal(sheet_sent.DueAt)).Equal(true)

      sheets_after, err := stores.Sheet.SheetsOfCourse(1, false)
      g.Assert(err).Equal(nil)
      g.Assert(len(sheets_after)).Equal(len(sheets_before) + 1)
    })

    g.It("Should skip non-existent sheet file", func() {
      w := tape.GetWithClaims("/api/v1/sheets/1/file", 1, true)
      g.Assert(w.Code).Equal(http.StatusNotFound)
    })

    g.It("Should upload sheet file", func() {
      defer helper.NewSheetFileHandle(1).Delete()

      // no file so far
      g.Assert(helper.NewSheetFileHandle(1).Exists()).Equal(false)

      // upload file
      filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
      w, err := tape.UploadWithClaims("/api/v1/sheets/1/file", filename, "application/zip", 1, true)
      g.Assert(err).Equal(nil)
      g.Assert(w.Code).Equal(http.StatusOK)

      // check disk
      g.Assert(helper.NewSheetFileHandle(1).Exists()).Equal(true)

      // a file should be now served
      w = tape.GetWithClaims("/api/v1/sheets/1/file", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Changes should require claims", func() {
      w := tape.Put("/api/v1/courses/1/sheets", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.It("Should perform updates", func() {

      sheet_sent := model.Sheet{
        Name:      "Sheet_update",
        PublishAt: helper.Time(time.Now()),
        DueAt:     helper.Time(time.Now()),
      }

      w := tape.PutWithClaims("/api/v1/sheets/1",
        tape.ToH(sheet_sent), 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      sheet_after, err := stores.Sheet.Get(1)
      g.Assert(err).Equal(nil)
      g.Assert(sheet_after.Name).Equal("Sheet_update")
      g.Assert(sheet_after.PublishAt.Equal(sheet_sent.PublishAt)).Equal(true)
      g.Assert(sheet_after.DueAt.Equal(sheet_sent.DueAt)).Equal(true)
    })

    g.It("Should delete when valid access claims", func() {
      entries_before, err := stores.Sheet.GetAll()
      g.Assert(err).Equal(nil)

      w := tape.Delete("/api/v1/sheets/1")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // verify nothing has changes
      entries_after, err := stores.Sheet.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before))

      w = tape.DeleteWithClaims("/api/v1/sheets/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      // verify a sheet less exists
      entries_after, err = stores.Sheet.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before) - 1)
    })

    g.AfterEach(func() {
      tape.AfterEach()
    })
  })

}
