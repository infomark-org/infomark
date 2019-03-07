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
  "time"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/franela/goblin"
  "github.com/spf13/viper"
)

func TestMaterial(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

  g.Describe("Material", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      stores = NewStores(tape.DB)
    })

    g.It("Query should require access claims", func() {

      w := tape.Get("/api/v1/courses/1/materials")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.GetWithClaims("/api/v1/courses/1/materials", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Should list all materials a course", func() {

      w := tape.GetWithClaims("/api/v1/courses/1/materials", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      materials_actual := []MaterialResponse{}
      err := json.NewDecoder(w.Body).Decode(&materials_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(materials_actual)).Equal(10)
    })

    g.It("Should get a specific material", func() {
      material_expected, err := stores.Material.Get(1)
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims("/api/v1/materials/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      material_actual := &MaterialResponse{}
      err = json.NewDecoder(w.Body).Decode(material_actual)
      g.Assert(err).Equal(nil)

      g.Assert(material_actual.ID).Equal(material_expected.ID)
      // g.Assert(material_actual.Name).Equal(material_expected.Name)
      // g.Assert(material_actual.PublishAt.Equal(material_expected.PublishAt)).Equal(true)
      // g.Assert(material_actual.DueAt.Equal(material_expected.DueAt)).Equal(true)
    })

    g.It("Should create valid material", func() {

      materials_before, err := stores.Material.MaterialsOfCourse(1, false)
      g.Assert(err).Equal(nil)

      material_sent := MaterialRequest{
        Name:      "Material_new",
        Filename:  "Filename",
        Kind:      0,
        PublishAt: helper.Time(time.Now()),
        LectureAt: helper.Time(time.Now()),
      }

      g.Assert(material_sent.Validate()).Equal(nil)

      w := tape.Post("/api/v1/courses/1/materials", helper.ToH(material_sent))
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.PostWithClaims("/api/v1/courses/1/materials", helper.ToH(material_sent), 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      w = tape.PostWithClaims("/api/v1/courses/1/materials", helper.ToH(material_sent), 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      w = tape.PostWithClaims("/api/v1/courses/1/materials", helper.ToH(material_sent), 1, false)
      g.Assert(w.Code).Equal(http.StatusCreated)

      material_return := &MaterialResponse{}
      err = json.NewDecoder(w.Body).Decode(&material_return)
      g.Assert(err).Equal(nil)
      g.Assert(material_return.Name).Equal(material_sent.Name)
      g.Assert(material_return.Filename).Equal(material_sent.Filename)
      g.Assert(material_return.Kind).Equal(material_sent.Kind)
      g.Assert(material_return.PublishAt.Equal(material_sent.PublishAt)).Equal(true)
      g.Assert(material_return.LectureAt.Equal(material_sent.LectureAt)).Equal(true)

      materials_after, err := stores.Material.MaterialsOfCourse(1, false)
      g.Assert(err).Equal(nil)

      g.Assert(len(materials_after)).Equal(len(materials_before) + 1)
    })

    g.It("Creating a material should require body", func() {
      w := tape.PlayDataWithClaims("POST", "/api/v1/courses/1/materials", H{}, 1, true)
      g.Assert(w.Code).Equal(http.StatusBadRequest)
    })

    g.It("Should not create material with missing data", func() {
      data := H{
        "name":       "Sheet_new",
        "publish_at": "2019-02-01T01:02:03Z",
        "lecture_at": "2019-02-01T01:02:03Z",
        // "due_at" is be missing
      }

      w := tape.PlayDataWithClaims("POST", "/api/v1/courses/1/materials", data, 1, true)
      g.Assert(w.Code).Equal(http.StatusBadRequest)
    })

    g.It("Should skip non-existent material file", func() {
      w := tape.GetWithClaims("/api/v1/materials/1/file", 1, true)
      g.Assert(w.Code).Equal(http.StatusNotFound)
    })

    g.It("Should upload material file", func() {
      defer helper.NewMaterialFileHandle(1).Delete()

      // no file so far
      g.Assert(helper.NewMaterialFileHandle(1).Exists()).Equal(false)
      filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))

      // students
      w, err := tape.UploadWithClaims("/api/v1/materials/1/file", filename, "application/zip", 112, false)
      g.Assert(err).Equal(nil)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w, err = tape.UploadWithClaims("/api/v1/materials/1/file", filename, "application/zip", 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w, err = tape.UploadWithClaims("/api/v1/materials/1/file", filename, "application/zip", 1, false)
      g.Assert(err).Equal(nil)
      fmt.Println(w.Body)
      g.Assert(w.Code).Equal(http.StatusOK)

      // check disk
      g.Assert(helper.NewMaterialFileHandle(1).Exists()).Equal(true)

      // a file should be now served
      w = tape.GetWithClaims("/api/v1/materials/1/file", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Changes should require claims", func() {
      w := tape.Put("/api/v1/courses/1/materials", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.It("Should perform updates", func() {

      material_sent := MaterialRequest{
        Name:      "Material_new",
        Filename:  "Filename",
        Kind:      0,
        PublishAt: helper.Time(time.Now()),
        LectureAt: helper.Time(time.Now()),
      }

      // students
      w := tape.PutWithClaims("/api/v1/materials/1", tape.ToH(material_sent), 122, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.PutWithClaims("/api/v1/materials/1", tape.ToH(material_sent), 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PutWithClaims("/api/v1/materials/1", tape.ToH(material_sent), 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      material_after, err := stores.Material.Get(1)
      g.Assert(err).Equal(nil)
      g.Assert(material_after.Name).Equal(material_sent.Name)
      g.Assert(material_after.Filename).Equal(material_sent.Filename)
      g.Assert(material_after.Kind).Equal(material_sent.Kind)
      g.Assert(material_after.PublishAt.Equal(material_sent.PublishAt)).Equal(true)
      g.Assert(material_after.LectureAt.Equal(material_sent.LectureAt)).Equal(true)
    })

    g.It("Should delete when valid access claims", func() {
      entries_before, err := stores.Material.GetAll()
      g.Assert(err).Equal(nil)

      w := tape.Delete("/api/v1/materials/1")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // students
      w = tape.DeleteWithClaims("/api/v1/materials/1", 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.DeleteWithClaims("/api/v1/materials/1", 2, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // verify nothing has changes
      entries_after, err := stores.Material.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before))

      // admin
      w = tape.DeleteWithClaims("/api/v1/materials/1", 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // verify a sheet less exists
      entries_after, err = stores.Material.GetAll()
      g.Assert(err).Equal(nil)
      g.Assert(len(entries_after)).Equal(len(entries_before) - 1)
    })

    g.It("Permission test", func() {
      url := "/api/v1/courses/1/materials"

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

    g.AfterEach(func() {
      tape.AfterEach()
    })
  })

}
