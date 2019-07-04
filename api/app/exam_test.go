// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  Infomark Authors
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
	"net/http"
	"testing"
	"time"

	"github.com/franela/goblin"
	"github.com/infomark-org/infomark-backend/api/helper"
	"github.com/infomark-org/infomark-backend/email"
)

func TestExam(t *testing.T) {
	PrepareTests()
	g := goblin.Goblin(t)
	email.DefaultMail = email.VoidMail

	tape := &Tape{}

	var stores *Stores

	studentJWT := NewJWTRequest(112, false)
	tutorJWT := NewJWTRequest(2, false)
	adminJWT := NewJWTRequest(1, true)

	g.Describe("Exam", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
			stores = NewStores(tape.DB)
			_ = stores
		})

		g.It("Query should require access claims", func() {

			w := tape.Get("/api/v1/courses/1/exams")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/courses/1/exams", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			w = tape.Get("/api/v1/courses/1/exams", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			w = tape.Get("/api/v1/courses/1/exams", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Should list all exams from a exams", func() {
			w := tape.Get("/api/v1/courses/1/exams", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			examsActual := []ExamResponse{}
			err := json.NewDecoder(w.Body).Decode(&examsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(examsActual)).Equal(2)
		})

		g.It("Should get a specific exam", func() {
			entryExpected, err := stores.Exam.Get(1)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/exams/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			entryActual := &ExamResponse{}
			err = json.NewDecoder(w.Body).Decode(entryActual)
			g.Assert(err).Equal(nil)

			g.Assert(entryActual.ID).Equal(entryExpected.ID)
			g.Assert(entryActual.Name).Equal(entryExpected.Name)
			g.Assert(entryActual.Description).Equal(entryExpected.Description)
			g.Assert(entryActual.ExamTime.Equal(entryExpected.ExamTime)).Equal(true)
			g.Assert(entryActual.CourseID).Equal(entryExpected.CourseID)
		})

		g.It("Creating should require claims and admin priviledges", func() {
			w := tape.Post("/api/v1/courses/1/exams", H{})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Post("/api/v1/courses/1/exams", H{}, studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Post("/api/v1/courses/1/exams", H{}, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)
		})

		g.It("Should create valid exam", func() {
			entriesBefore, err := stores.Exam.ExamsOfCourse(1)
			g.Assert(err).Equal(nil)

			entrySent := ExamRequest{
				Name:        "Exam_new",
				Description: "blah blahe",
				ExamTime:    helper.Time(time.Now()),
			}

			err = entrySent.Validate()
			g.Assert(err).Equal(nil)

			// students
			w := tape.Post("/api/v1/courses/1/exams", tape.ToH(entrySent), studentJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Post("/api/v1/courses/1/exams", tape.ToH(entrySent), tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Post("/api/v1/courses/1/exams", tape.ToH(entrySent), adminJWT)
			g.Assert(w.Code).Equal(http.StatusCreated)

			examReturn := &ExamResponse{}
			err = json.NewDecoder(w.Body).Decode(&examReturn)
			g.Assert(err).Equal(nil)
			g.Assert(examReturn.Name).Equal(entrySent.Name)
			g.Assert(examReturn.Description).Equal(entrySent.Description)
			g.Assert(examReturn.ExamTime.Equal(entrySent.ExamTime)).Equal(true)

			entriesAfter, err := stores.Exam.ExamsOfCourse(1)
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore) + 1)

		})

		g.It("Should update a group", func() {
			entrySent := ExamRequest{
				Name:        "Exam_new_updated",
				Description: "blah blahe_updated",
				ExamTime:    helper.Time(time.Now()),
			}

			// students
			w := tape.Put("/api/v1/courses/1/exams/1", tape.ToH(entrySent), studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Put("/api/v1/courses/1/exams/1", tape.ToH(entrySent), tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Put("/api/v1/courses/1/exams/1", tape.ToH(entrySent), adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			entryAfter, err := stores.Exam.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(err).Equal(nil)
			g.Assert(entryAfter.Name).Equal(entrySent.Name)
			g.Assert(entryAfter.Description).Equal(entrySent.Description)
			g.Assert(entryAfter.ExamTime.Equal(entrySent.ExamTime)).Equal(true)
		})

		g.It("Should delete when valid access claims", func() {
			entriesBefore, err := stores.Exam.GetAll()
			g.Assert(err).Equal(nil)

			w := tape.Delete("/api/v1/courses/1/exams/1")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// verify nothing has changes
			entriesAfter, err := stores.Exam.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore))

			// students
			w = tape.Delete("/api/v1/courses/1/exams/1", studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Delete("/api/v1/courses/1/exams/1", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Delete("/api/v1/courses/1/exams/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// verify a sheet less exists
			entriesAfter, err = stores.Exam.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore) - 1)
		})

		g.It("Should be able to fetch enrollments", func() {
			examActive, err := stores.Exam.Get(1)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/exams/1/enrollments", studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Get("/api/v1/courses/1/exams/1/enrollments", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			numberEnrollmentsExpected, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_exam WHERE exam_id = $1",
				examActive.ID,
			)
			g.Assert(err).Equal(nil)

			w = tape.Get("/api/v1/courses/1/exams/1/enrollments", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			enrollmentsActual := []ExamEnrollmentResponse{}
			err = json.NewDecoder(w.Body).Decode(&enrollmentsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(enrollmentsActual)).Equal(numberEnrollmentsExpected)
		})

		g.It("Students should be able to enroll into exam", func() {
			// remove all enrollments from student
			_, err := tape.DB.Exec("DELETE FROM user_exam WHERE user_id = 112;")
			g.Assert(err).Equal(nil)

			examsBefore, err := stores.Exam.GetEnrollmentsOfUser(studentJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)
			g.Assert(len(examsBefore)).Equal(0)

			w := tape.Post("/api/v1/courses/1/exams/1/enrollments", helper.H{}, adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			w = tape.Post("/api/v1/courses/1/exams/1/enrollments", helper.H{}, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			w = tape.Post("/api/v1/courses/1/exams/1/enrollments", helper.H{}, studentJWT)
			g.Assert(w.Code).Equal(http.StatusCreated)

			examsAfter, err := stores.Exam.GetEnrollmentsOfUser(studentJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)
			g.Assert(len(examsAfter)).Equal(1)

		})

		g.It("Students should be able to disenroll from exam", func() {
			// remove all enrollments from student

			examsBefore, err := stores.Exam.GetEnrollmentsOfUser(studentJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)
			g.Assert(len(examsBefore)).Equal(2)

			w := tape.Delete("/api/v1/courses/1/exams/1/enrollments", adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			w = tape.Delete("/api/v1/courses/1/exams/1/enrollments", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			w = tape.Delete("/api/v1/courses/1/exams/1/enrollments", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			examsAfter, err := stores.Exam.GetEnrollmentsOfUser(studentJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)
			g.Assert(len(examsAfter)).Equal(len(examsBefore) - 1)

		})

		g.It("Admins scan update mark and status", func() {
			// remove all enrollments from student

			exam, err := stores.Exam.GetEnrollmentOfUser(int64(1), studentJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)
			g.Assert(exam.Status).Equal(0)

			entrySent := helper.H{
				"user_id": studentJWT.Claims.LoginID,
				"mark":    "passed very good",
				"status":  3,
			}

			w := tape.Put("/api/v1/courses/1/exams/1/enrollments", entrySent, studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Put("/api/v1/courses/1/exams/1/enrollments", entrySent, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Put("/api/v1/courses/1/exams/1/enrollments", entrySent, adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			examAfter, err := stores.Exam.GetEnrollmentOfUser(int64(1), studentJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)
			g.Assert(examAfter.Mark).Equal("passed very good")
			g.Assert(examAfter.Status).Equal(3)

		})

		g.It("Students should not be able to disenroll from exam when status has changed", func() {
			// remove all enrollments from student

			examsBefore, err := stores.Exam.GetEnrollmentsOfUser(studentJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)
			g.Assert(len(examsBefore)).Equal(2)

			exam, err := stores.Exam.GetEnrollmentOfUser(int64(1), studentJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)

			exam.Status = 2
			stores.Exam.UpdateUserExam(exam)
			g.Assert(err).Equal(nil)

			w := tape.Delete("/api/v1/courses/1/exams/1/enrollments", adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			w = tape.Delete("/api/v1/courses/1/exams/1/enrollments", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			w = tape.Delete("/api/v1/courses/1/exams/1/enrollments", studentJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			examsAfter, err := stores.Exam.GetEnrollmentsOfUser(studentJWT.Claims.LoginID)
			g.Assert(err).Equal(nil)
			g.Assert(len(examsAfter)).Equal(len(examsBefore))

		})

		g.AfterEach(func() {
			tape.AfterEach()
		})

	})

}
