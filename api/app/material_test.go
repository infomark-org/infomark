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
	"mime"
	"net/http"
	"testing"
	"time"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/auth/authorize"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/franela/goblin"
	"github.com/spf13/viper"
)

func TestMaterial(t *testing.T) {
	PrepareTests()
	g := goblin.Goblin(t)
	email.DefaultMail = email.VoidMail

	tape := &Tape{}

	var stores *Stores

	studentJWT := NewJWTRequest(112, false)
	tutorJWT := NewJWTRequest(2, false)
	adminJWT := NewJWTRequest(1, true)
	noAdminJWT := NewJWTRequest(1, false)

	g.Describe("Material", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
			stores = NewStores(tape.DB)
		})

		g.It("Query should require access claims", func() {

			w := tape.Get("/api/v1/courses/1/materials")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/courses/1/materials", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Should list all materials a course (student)", func() {
			materialsExpected, err := stores.Material.MaterialsOfCourse(1, authorize.STUDENT.ToInt())
			g.Assert(err).Equal(nil)

			for _, mat := range materialsExpected {
				mat.PublishAt = NowUTC().Add(-time.Hour)
				stores.Material.Update(&mat)
			}

			user, err := stores.Course.GetUserEnrollment(1, studentJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)
			g.Assert(user.Role).Equal(int64(0))
			g.Assert(user.ID).Equal(studentJWT.Claims.LoginID)

			w := tape.Get("/api/v1/courses/1/materials", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			materialsActual := []MaterialResponse{}
			err = json.NewDecoder(w.Body).Decode(&materialsActual)
			g.Assert(err).Equal(nil)

			g.Assert(len(materialsActual)).Equal(len(materialsExpected))
		})

		g.It("Should list all materials a course (tutor)", func() {

			user, err := stores.Course.GetUserEnrollment(1, tutorJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)
			g.Assert(user.Role).Equal(int64(1))
			g.Assert(user.ID).Equal(tutorJWT.Claims.LoginID)

			w := tape.Get("/api/v1/courses/1/materials", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			materialsActual := []MaterialResponse{}
			err = json.NewDecoder(w.Body).Decode(&materialsActual)
			g.Assert(err).Equal(nil)

			materialsExpected, err := stores.Material.MaterialsOfCourse(1, authorize.TUTOR.ToInt())
			g.Assert(err).Equal(nil)
			g.Assert(len(materialsActual)).Equal(len(materialsExpected))
		})

		g.It("Should list all materials a course (admin)", func() {

			user, err := stores.Course.GetUserEnrollment(1, adminJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)
			g.Assert(user.Role).Equal(int64(2))
			g.Assert(user.ID).Equal(adminJWT.Claims.LoginID)

			w := tape.Get("/api/v1/courses/1/materials", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			materialsActual := []MaterialResponse{}
			err = json.NewDecoder(w.Body).Decode(&materialsActual)
			g.Assert(err).Equal(nil)

			materialsExpected, err := stores.Material.MaterialsOfCourse(1, authorize.ADMIN.ToInt())
			g.Assert(err).Equal(nil)
			g.Assert(len(materialsActual)).Equal(len(materialsExpected))
		})

		g.It("Should get a specific material", func() {
			materialExpected, err := stores.Material.Get(1)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/materials/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			materialActual := &MaterialResponse{}
			err = json.NewDecoder(w.Body).Decode(materialActual)
			g.Assert(err).Equal(nil)

			g.Assert(materialActual.ID).Equal(materialExpected.ID)
			g.Assert(materialActual.Name).Equal(materialExpected.Name)
			g.Assert(materialActual.PublishAt.Equal(materialExpected.PublishAt)).Equal(true)
			g.Assert(materialActual.LectureAt.Equal(materialExpected.LectureAt)).Equal(true)
		})

		g.It("Should not get a specific material (unpublish)", func() {
			materialExpected, err := stores.Material.Get(1)
			g.Assert(err).Equal(nil)
			materialExpected.RequiredRole = 0
			stores.Material.Update(materialExpected)

			materialExpected.PublishAt = NowUTC().Add(time.Hour)
			err = stores.Material.Update(materialExpected)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/materials/1", studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			materialExpected.PublishAt = NowUTC().Add(-time.Hour)
			err = stores.Material.Update(materialExpected)
			g.Assert(err).Equal(nil)

			w = tape.Get("/api/v1/courses/1/materials/1", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

		})

		g.It("Should create valid material for tutors", func() {

			materialsBeforeStudent, err := stores.Material.MaterialsOfCourse(1, authorize.STUDENT.ToInt())
			g.Assert(err).Equal(nil)

			materialsBeforeTutor, err := stores.Material.MaterialsOfCourse(1, authorize.TUTOR.ToInt())
			g.Assert(err).Equal(nil)

			materialsBeforeAdmin, err := stores.Material.MaterialsOfCourse(1, authorize.ADMIN.ToInt())
			g.Assert(err).Equal(nil)

			materialSent := MaterialRequest{
				Name:         "Material_new",
				Kind:         0,
				RequiredRole: authorize.TUTOR.ToInt(),
				PublishAt:    helper.Time(time.Now()),
				LectureAt:    helper.Time(time.Now()),
			}

			g.Assert(materialSent.Validate()).Equal(nil)

			w := tape.Post("/api/v1/courses/1/materials", helper.ToH(materialSent))
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Post("/api/v1/courses/1/materials", helper.ToH(materialSent), studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Post("/api/v1/courses/1/materials", helper.ToH(materialSent), tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Post("/api/v1/courses/1/materials", helper.ToH(materialSent), noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusCreated)

			materialsAfterStudent, err := stores.Material.MaterialsOfCourse(1, authorize.STUDENT.ToInt())
			g.Assert(err).Equal(nil)
			materialsAfterTutor, err := stores.Material.MaterialsOfCourse(1, authorize.TUTOR.ToInt())
			g.Assert(err).Equal(nil)
			materialsAfterAdmin, err := stores.Material.MaterialsOfCourse(1, authorize.ADMIN.ToInt())
			g.Assert(err).Equal(nil)

			materialReturn := &MaterialResponse{}
			err = json.NewDecoder(w.Body).Decode(&materialReturn)
			g.Assert(err).Equal(nil)
			g.Assert(materialReturn.Name).Equal(materialSent.Name)
			g.Assert(materialReturn.Kind).Equal(materialSent.Kind)
			g.Assert(materialReturn.RequiredRole).Equal(materialSent.RequiredRole)
			g.Assert(materialReturn.PublishAt.Equal(materialSent.PublishAt)).Equal(true)
			g.Assert(materialReturn.LectureAt.Equal(materialSent.LectureAt)).Equal(true)

			g.Assert(len(materialsAfterStudent)).Equal(len(materialsBeforeStudent))
			g.Assert(len(materialsAfterTutor)).Equal(len(materialsBeforeTutor) + 1)
			g.Assert(len(materialsAfterAdmin)).Equal(len(materialsBeforeAdmin) + 1)
		})

		g.It("Should create valid material for admins", func() {

			materialsBeforeStudent, err := stores.Material.MaterialsOfCourse(1, authorize.STUDENT.ToInt())
			g.Assert(err).Equal(nil)

			materialsBeforeTutor, err := stores.Material.MaterialsOfCourse(1, authorize.TUTOR.ToInt())
			g.Assert(err).Equal(nil)

			materialsBeforeAdmin, err := stores.Material.MaterialsOfCourse(1, authorize.ADMIN.ToInt())
			g.Assert(err).Equal(nil)

			materialSent := MaterialRequest{
				Name:         "Material_new",
				Kind:         0,
				RequiredRole: authorize.ADMIN.ToInt(),
				PublishAt:    helper.Time(time.Now()),
				LectureAt:    helper.Time(time.Now()),
			}

			g.Assert(materialSent.Validate()).Equal(nil)

			w := tape.Post("/api/v1/courses/1/materials", helper.ToH(materialSent))
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Post("/api/v1/courses/1/materials", helper.ToH(materialSent), studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Post("/api/v1/courses/1/materials", helper.ToH(materialSent), tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Post("/api/v1/courses/1/materials", helper.ToH(materialSent), noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusCreated)

			materialsAfterStudent, err := stores.Material.MaterialsOfCourse(1, authorize.STUDENT.ToInt())
			g.Assert(err).Equal(nil)
			materialsAfterTutor, err := stores.Material.MaterialsOfCourse(1, authorize.TUTOR.ToInt())
			g.Assert(err).Equal(nil)
			materialsAfterAdmin, err := stores.Material.MaterialsOfCourse(1, authorize.ADMIN.ToInt())
			g.Assert(err).Equal(nil)

			materialReturn := &MaterialResponse{}
			err = json.NewDecoder(w.Body).Decode(&materialReturn)
			g.Assert(err).Equal(nil)
			g.Assert(materialReturn.Name).Equal(materialSent.Name)
			g.Assert(materialReturn.Kind).Equal(materialSent.Kind)
			g.Assert(materialReturn.RequiredRole).Equal(materialSent.RequiredRole)
			g.Assert(materialReturn.PublishAt.Equal(materialSent.PublishAt)).Equal(true)
			g.Assert(materialReturn.LectureAt.Equal(materialSent.LectureAt)).Equal(true)

			g.Assert(len(materialsAfterStudent)).Equal(len(materialsBeforeStudent))
			g.Assert(len(materialsAfterTutor)).Equal(len(materialsBeforeTutor))
			g.Assert(len(materialsAfterAdmin)).Equal(len(materialsBeforeAdmin) + 1)
		})

		g.It("Creating a material should require body", func() {
			w := tape.Post("/api/v1/courses/1/materials", H{}, adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should not create material with missing data", func() {
			data := H{
				"name":       "Sheet_new",
				"publish_at": "2019-02-01T01:02:03Z",
				// "lecture_at" is be missing
			}

			w := tape.Post("/api/v1/courses/1/materials", data, adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should skip non-existent material file", func() {

			hnd := helper.NewMaterialFileHandle(1)
			g.Assert(hnd.Exists()).Equal(false)
			g.Assert(hnd.Exists()).Equal(false)

			w := tape.Get("/api/v1/courses/1/materials/1/file", adminJWT)
			g.Assert(w.Code).Equal(http.StatusNotFound)
		})

		g.It("Should upload material file", func() {
			defer helper.NewMaterialFileHandle(1).Delete()

			// set to publish
			material, err := stores.Material.Get(1)
			g.Assert(err).Equal(nil)
			material.PublishAt = NowUTC().Add(-2 * time.Hour)
			err = stores.Material.Update(material)
			g.Assert(err).Equal(nil)

			// no file so far
			g.Assert(helper.NewMaterialFileHandle(1).Exists()).Equal(false)
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))

			// students
			w, err := tape.Upload("/api/v1/courses/1/materials/1/file", filename, "application/zip", studentJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w, err = tape.Upload("/api/v1/courses/1/materials/1/file", filename, "application/zip", tutorJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w, err = tape.Upload("/api/v1/courses/1/materials/1/file", filename, "application/zip", noAdminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)

			// check disk
			hnd := helper.NewMaterialFileHandle(1)
			g.Assert(hnd.Exists()).Equal(true)

			// a file should be now served
			w = tape.Get("/api/v1/courses/1/materials/1/file", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Should upload material file (zip)", func() {
			defer helper.NewMaterialFileHandle(1).Delete()
			filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			// admin
			w, err := tape.Upload("/api/v1/courses/1/materials/1/file", filename, "application/zip", noAdminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)

			// check disk
			hnd := helper.NewMaterialFileHandle(1)
			g.Assert(hnd.Exists()).Equal(true)

			// a file should be now served
			w = tape.Get("/api/v1/courses/1/materials/1/file", adminJWT)
			g.Assert(w.Header().Get("Content-Type")).Equal("application/zip")
			g.Assert(w.Code).Equal(http.StatusOK)

			course, err := stores.Material.IdentifyCourseOfMaterial(1)
			g.Assert(err).Equal(nil)

			_, params, err := mime.ParseMediaType(w.Header().Get("Content-Disposition"))
			g.Assert(err).Equal(nil)
			g.Assert(params["filename"]).Equal(fmt.Sprintf("%s-empty.zip", course.Name))
		})

		g.It("Should upload material file (pdf)", func() {
			defer helper.NewMaterialFileHandle(1).Delete()
			filename := fmt.Sprintf("%s/empty.pdf", viper.GetString("fixtures_dir"))
			// admin
			w, err := tape.Upload("/api/v1/courses/1/materials/1/file", filename, "application/pdf", noAdminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)

			// check disk
			hnd := helper.NewMaterialFileHandle(1)
			g.Assert(hnd.Exists()).Equal(true)

			// a file should be now served
			w = tape.Get("/api/v1/courses/1/materials/1/file", adminJWT)
			g.Assert(w.Header().Get("Content-Type")).Equal("application/pdf")
			g.Assert(w.Code).Equal(http.StatusOK)

			course, err := stores.Material.IdentifyCourseOfMaterial(1)
			g.Assert(err).Equal(nil)

			_, params, err := mime.ParseMediaType(w.Header().Get("Content-Disposition"))
			g.Assert(err).Equal(nil)
			g.Assert(params["filename"]).Equal(fmt.Sprintf("%s-empty.pdf", course.Name))
		})

		g.It("Changes should require claims", func() {
			w := tape.Put("/api/v1/courses/1/materials", H{})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should perform updates", func() {

			// set to publish
			material, err := stores.Material.Get(1)
			g.Assert(err).Equal(nil)
			material.PublishAt = NowUTC().Add(-2 * time.Hour)
			err = stores.Material.Update(material)
			g.Assert(err).Equal(nil)

			materialSent := MaterialRequest{
				Name:      "Material_new",
				Kind:      0,
				PublishAt: helper.Time(time.Now()),
				LectureAt: helper.Time(time.Now()),
			}

			// students
			w := tape.Put("/api/v1/courses/1/materials/1", tape.ToH(materialSent), studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Put("/api/v1/courses/1/materials/1", tape.ToH(materialSent), tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Put("/api/v1/courses/1/materials/1", tape.ToH(materialSent), adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			materialAfter, err := stores.Material.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(materialAfter.Name).Equal(materialSent.Name)
			g.Assert(materialAfter.Kind).Equal(materialSent.Kind)
			g.Assert(materialAfter.PublishAt.Equal(materialSent.PublishAt)).Equal(true)
			g.Assert(materialAfter.LectureAt.Equal(materialSent.LectureAt)).Equal(true)
		})

		g.It("Should delete when valid access claims", func() {

			// set to publish
			material, err := stores.Material.Get(1)
			g.Assert(err).Equal(nil)
			material.PublishAt = NowUTC().Add(-2 * time.Hour)
			err = stores.Material.Update(material)
			g.Assert(err).Equal(nil)

			entriesBefore, err := stores.Material.GetAll()
			g.Assert(err).Equal(nil)

			w := tape.Delete("/api/v1/courses/1/materials/1")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// students
			w = tape.Delete("/api/v1/courses/1/materials/1", studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Delete("/api/v1/courses/1/materials/1", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// verify nothing has changes
			entriesAfter, err := stores.Material.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore))

			// admin
			w = tape.Delete("/api/v1/courses/1/materials/1", noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// verify a sheet less exists
			entriesAfter, err = stores.Material.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore) - 1)
		})

		g.It("Should delete when valid access claims and not published", func() {

			// set to publish
			material, err := stores.Material.Get(1)
			g.Assert(err).Equal(nil)
			material.PublishAt = NowUTC().Add(2 * time.Hour)
			err = stores.Material.Update(material)
			g.Assert(err).Equal(nil)

			entriesBefore, err := stores.Material.GetAll()
			g.Assert(err).Equal(nil)

			// admin
			w := tape.Delete("/api/v1/courses/1/materials/1", noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// verify a sheet less exists
			entriesAfter, err := stores.Material.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore) - 1)
		})

		g.It("Permission test", func() {
			url := "/api/v1/courses/1/materials"

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
