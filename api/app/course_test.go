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
	"net/http"
	"testing"
	"time"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/franela/goblin"
)

func DBGetInt(tape *Tape, stmt string, param1 int64) (int, error) {
	var rsl int
	err := tape.DB.Get(&rsl, stmt, param1)
	return rsl, err
}

func DBGetInt2(tape *Tape, stmt string, param1 int64, param2 int64) (int, error) {
	var rsl int
	err := tape.DB.Get(&rsl, stmt, param1, param2)
	return rsl, err
}

func TestCourse(t *testing.T) {
	PrepareTests()

	g := goblin.Goblin(t)
	email.DefaultMail = email.VoidMail
	// email.DefaultMail = email.TerminalMail
	go email.BackgroundSend(email.OutgoingEmailsChannel)

	tape := &Tape{}

	var stores *Stores

	studentJWT := NewJWTRequest(112, false)
	tutorJWT := NewJWTRequest(2, false)
	adminJWT := NewJWTRequest(1, true)
	noAdminJWT := NewJWTRequest(1, false)

	g.Describe("Course", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
			stores = NewStores(tape.DB)
			_ = stores
		})

		g.It("Query should require claims", func() {

			w := tape.Get("/api/v1/courses")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/courses", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

		})

		g.It("Should list all courses", func() {
			w := tape.Get("/api/v1/courses", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			coursesActual := []model.Course{}
			err := json.NewDecoder(w.Body).Decode(&coursesActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(coursesActual)).Equal(2)
		})

		g.It("Should get a specific course", func() {

			w := tape.Get("/api/v1/courses/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			courseActual := &CourseResponse{}
			err := json.NewDecoder(w.Body).Decode(courseActual)
			g.Assert(err).Equal(nil)

			courseExpected, err := stores.Course.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(courseActual.ID).Equal(courseExpected.ID)
			g.Assert(courseActual.Name).Equal(courseExpected.Name)
			g.Assert(courseActual.Description).Equal(courseExpected.Description)
			g.Assert(courseActual.BeginsAt.Equal(courseExpected.BeginsAt)).Equal(true)
			g.Assert(courseActual.EndsAt.Equal(courseExpected.EndsAt)).Equal(true)
			g.Assert(courseActual.RequiredPercentage).Equal(courseExpected.RequiredPercentage)
		})

		g.It("Should be able to filter enrollments (all)", func() {
			courseActive, err := stores.Course.Get(1)
			g.Assert(err).Equal(nil)

			numberEnrollmentsExpected, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1",
				courseActive.ID,
			)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/enrollments", adminJWT)
			enrollmentsActual := []EnrollmentResponse{}
			err = json.NewDecoder(w.Body).Decode(&enrollmentsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(enrollmentsActual)).Equal(numberEnrollmentsExpected)
		})

		g.It("Should be able to filter enrollments (students only)", func() {
			courseActive, err := stores.Course.Get(1)
			g.Assert(err).Equal(nil)

			numberEnrollmentsExpected, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
				courseActive.ID,
			)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/enrollments?roles=0", adminJWT)
			enrollmentsActual := []EnrollmentResponse{}
			err = json.NewDecoder(w.Body).Decode(&enrollmentsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(enrollmentsActual)).Equal(numberEnrollmentsExpected)
		})

		g.It("Should be able to query enrollments (tutor+admin only)", func() {
			courseActive, err := stores.Course.Get(1)
			g.Assert(err).Equal(nil)

			enrollmentsExpected, err := stores.Course.FindEnrolledUsers(courseActive.ID,
				[]string{"0", "1", "2"}, "%chi%",
			)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/enrollments?q=chi", noAdminJWT)
			enrollmentsActual := []EnrollmentResponse{}
			err = json.NewDecoder(w.Body).Decode(&enrollmentsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(enrollmentsActual)).Equal(len(enrollmentsExpected))
		})

		g.It("Should be able to filter enrollments (tutors only)", func() {
			courseActive, err := stores.Course.Get(1)
			g.Assert(err).Equal(nil)

			numberEnrollmentsExpected, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role = 1",
				courseActive.ID,
			)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/enrollments?roles=1", noAdminJWT)
			enrollmentsActual := []EnrollmentResponse{}
			err = json.NewDecoder(w.Body).Decode(&enrollmentsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(enrollmentsActual)).Equal(numberEnrollmentsExpected)
		})

		g.It("Should be able to filter enrollments (students+tutors only)", func() {
			courseActive, err := stores.Course.Get(1)
			g.Assert(err).Equal(nil)

			numberEnrollmentsExpected, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role IN (0,1)",
				courseActive.ID,
			)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/courses/1/enrollments?roles=0,1", noAdminJWT)
			enrollmentsActual := []EnrollmentResponse{}
			err = json.NewDecoder(w.Body).Decode(&enrollmentsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(enrollmentsActual)).Equal(numberEnrollmentsExpected)
		})

		g.It("Should be able to filter enrollments (but receive only tutors + admins), when role=student", func() {
			courseActive, err := stores.Course.Get(1)
			g.Assert(err).Equal(nil)

			numberEnrollmentsExpected, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role IN (1, 2)",
				courseActive.ID,
			)
			g.Assert(err).Equal(nil)

			// 112 is a student
			w := tape.Get("/api/v1/courses/1/enrollments?roles=0", studentJWT)
			enrollmentsActual := []EnrollmentResponse{}
			err = json.NewDecoder(w.Body).Decode(&enrollmentsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(enrollmentsActual)).Equal(numberEnrollmentsExpected)
		})

		g.It("Should be able to filter enrollments (but not see field protected by privacy), when role=tutor,student", func() {
			w := tape.Get("/api/v1/courses/1/enrollments?roles=0", studentJWT)
			enrollmentsActual := []EnrollmentResponse{}
			err := json.NewDecoder(w.Body).Decode(&enrollmentsActual)
			g.Assert(err).Equal(nil)

			for _, el := range enrollmentsActual {
				g.Assert(el.User.StudentNumber).Equal("")
			}
		})

		g.It("Creating course should require claims", func() {
			w := tape.Post("/api/v1/courses", H{})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Creating course should require body", func() {
			w := tape.Post("/api/v1/courses", make(map[string]interface{}), adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should create valid course", func() {

			coursesBefore, err := stores.Course.GetAll()
			g.Assert(err).Equal(nil)

			entrySent := courseRequest{
				Name:               "Info2_new",
				Description:        "Lorem Ipsum_new",
				BeginsAt:           helper.Time(time.Now()),
				EndsAt:             helper.Time(time.Now().Add(time.Hour * 1)),
				RequiredPercentage: 43,
			}

			g.Assert(entrySent.Validate()).Equal(nil)

			// students
			w := tape.Post("/api/v1/courses", tape.ToH(entrySent), studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Post("/api/v1/courses", tape.ToH(entrySent), tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin in course (cannot be admin, course does not exists yet)
			w = tape.Post("/api/v1/courses", tape.ToH(entrySent), noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Post("/api/v1/courses", tape.ToH(entrySent), adminJWT)
			g.Assert(w.Code).Equal(http.StatusCreated)

			// verify body
			courseReturn := &CourseResponse{}
			err = json.NewDecoder(w.Body).Decode(&courseReturn)
			g.Assert(err).Equal(nil)
			g.Assert(courseReturn.Name).Equal(entrySent.Name)
			g.Assert(courseReturn.Description).Equal(entrySent.Description)
			g.Assert(courseReturn.BeginsAt.Equal(entrySent.BeginsAt)).Equal(true)
			g.Assert(courseReturn.EndsAt.Equal(entrySent.EndsAt)).Equal(true)
			g.Assert(courseReturn.RequiredPercentage).Equal(entrySent.RequiredPercentage)

			// verify database
			courseNew, err := stores.Course.Get(courseReturn.ID)
			g.Assert(err).Equal(nil)
			g.Assert(courseReturn.Name).Equal(courseNew.Name)
			g.Assert(courseReturn.Description).Equal(courseNew.Description)
			g.Assert(courseReturn.BeginsAt.Equal(courseNew.BeginsAt)).Equal(true)
			g.Assert(courseReturn.EndsAt.Equal(courseNew.EndsAt)).Equal(true)
			g.Assert(courseReturn.RequiredPercentage).Equal(courseNew.RequiredPercentage)

			coursesAfter, err := stores.Course.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(coursesAfter)).Equal(len(coursesBefore) + 1)
		})

		g.It("Should send email to all enrolled users", func() {
			w := tape.Post("/api/v1/courses/1/emails", H{
				"subject": "subj",
				"body":    "text",
			}, adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Changes should require access claims", func() {
			w := tape.Put("/api/v1/courses/1", H{})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should perform updates", func() {

			entrySent := courseRequest{
				Name:               "Info2_update",
				Description:        "Lorem Ipsum_update",
				BeginsAt:           helper.Time(time.Now()),
				EndsAt:             helper.Time(time.Now()),
				RequiredPercentage: 99,
			}

			g.Assert(entrySent.Validate()).Equal(nil)

			// students
			w := tape.Put("/api/v1/courses/1", tape.ToH(entrySent), studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Put("/api/v1/courses/1", tape.ToH(entrySent), tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Put("/api/v1/courses/1", tape.ToH(entrySent), adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			entryAfter, err := stores.Course.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(entryAfter.Name).Equal(entrySent.Name)
			g.Assert(entryAfter.Description).Equal(entrySent.Description)
			g.Assert(entryAfter.BeginsAt.Equal(entrySent.BeginsAt)).Equal(true)
			g.Assert(entryAfter.EndsAt.Equal(entrySent.EndsAt)).Equal(true)
			g.Assert(entryAfter.RequiredPercentage).Equal(entrySent.RequiredPercentage)
		})

		g.It("Should delete when valid access claims", func() {
			entriesBefore, err := stores.Course.GetAll()
			g.Assert(err).Equal(nil)

			w := tape.Delete("/api/v1/courses/1")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// verify nothing has changes
			entriesAfter, err := stores.Course.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore))

			// students
			w = tape.Delete("/api/v1/courses/1", studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Delete("/api/v1/courses/1", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Delete("/api/v1/courses/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// verify a course less exists
			entriesAfter, err = stores.Course.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore) - 1)
		})

		g.It("Non-Global root enroll as students", func() {
			courseID := int64(1)

			w := tape.Post("/api/v1/courses/1/enrollments", helper.H{}, studentJWT)
			g.Assert(w.Code).Equal(http.StatusCreated)

			role, err := DBGetInt2(
				tape,
				"SELECT role FROM user_course WHERE course_id = $1 and user_id = $2",
				courseID, studentJWT.Claims.LoginID,
			)
			g.Assert(err).Equal(nil)
			g.Assert(role).Equal(0)

		})

		g.It("Global root enroll as admins", func() {

			courseID := int64(1)
			localAdminJWT := NewJWTRequest(112, true)

			w := tape.Post("/api/v1/courses/1/enrollments", helper.H{}, localAdminJWT)
			g.Assert(w.Code).Equal(http.StatusCreated)

			role, err := DBGetInt2(
				tape,
				"SELECT role FROM user_course WHERE course_id = $1 and user_id = $2",
				courseID, localAdminJWT.Claims.LoginID,
			)
			g.Assert(err).Equal(nil)
			g.Assert(role).Equal(2)

		})

		g.It("Can disenroll from course", func() {

			courseID := int64(1)

			numberEnrollmentsBefore, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
				courseID,
			)
			g.Assert(err).Equal(nil)

			w := tape.Delete("/api/v1/courses/1/enrollments", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			numberEnrollmentsAfter, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
				courseID,
			)
			g.Assert(err).Equal(nil)
			g.Assert(numberEnrollmentsAfter).Equal(numberEnrollmentsBefore - 1)

		})

		g.It("Can disenroll a specific user from course", func() {

			courseID := int64(1)

			numberEnrollmentsBefore, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
				courseID,
			)
			g.Assert(err).Equal(nil)

			// admin
			w := tape.Delete("/api/v1/courses/1/enrollments/113", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			numberEnrollmentsAfter, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
				courseID,
			)
			g.Assert(err).Equal(nil)
			g.Assert(numberEnrollmentsAfter).Equal(numberEnrollmentsBefore - 1)

		})

		g.It("Cannot  disenroll a specific user from course if user is tutor", func() {

			courseID := int64(1)

			numberEnrollmentsBefore, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
				courseID,
			)
			g.Assert(err).Equal(nil)

			// admin
			w := tape.Delete("/api/v1/courses/1/enrollments/2", adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			numberEnrollmentsAfter, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
				courseID,
			)
			g.Assert(err).Equal(nil)
			g.Assert(numberEnrollmentsAfter).Equal(numberEnrollmentsBefore)

		})

		g.It("Cannot disenroll as a tutor from course", func() {
			courseID := int64(1)

			numberEnrollmentsBefore, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
				courseID,
			)
			g.Assert(err).Equal(nil)

			w := tape.Delete("/api/v1/courses/1/enrollments", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			numberEnrollmentsAfter, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_course WHERE course_id = $1 and role = 0",
				courseID,
			)
			g.Assert(err).Equal(nil)
			g.Assert(numberEnrollmentsAfter).Equal(numberEnrollmentsBefore)
		})

		g.It("should see bids in course", func() {

			// tutors cannot use this
			w := tape.Get("/api/v1/courses/1/bids", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			// admins will see all
			w = tape.Get("/api/v1/courses/1/bids", noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// students will see their own
			w = tape.Get("/api/v1/courses/1/bids", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

		})

		g.It("Show user enrollement info", func() {

			w := tape.Get("/api/v1/courses/1/enrollments/2", NewJWTRequest(122, false))
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Get("/api/v1/courses/1/enrollments/2", NewJWTRequest(3, false))
			g.Assert(w.Code).Equal(http.StatusForbidden)

			result := EnrollmentResponse{}

			w = tape.Get("/api/v1/courses/1/enrollments/2", noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			err := json.NewDecoder(w.Body).Decode(&result)
			g.Assert(err).Equal(nil)
			g.Assert(result.User.ID).Equal(int64(2))
			g.Assert(result.Role).Equal(int64(1))

			w = tape.Get("/api/v1/courses/1/enrollments/112", noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&result)
			g.Assert(err).Equal(nil)
			g.Assert(result.User.ID).Equal(int64(112))
			g.Assert(result.Role).Equal(int64(0))

		})

		g.It("Should update role", func() {

			w := tape.Put("/api/v1/courses/1/enrollments/112", H{"role": 1}, studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Put("/api/v1/courses/1/enrollments/112", H{"role": 1}, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			w = tape.Put("/api/v1/courses/1/enrollments/112", H{"role": 1}, noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			w = tape.Get("/api/v1/courses/1/enrollments/112", noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			result := EnrollmentResponse{}
			err := json.NewDecoder(w.Body).Decode(&result)
			g.Assert(err).Equal(nil)
			g.Assert(result.User.ID).Equal(int64(112))
			g.Assert(result.Role).Equal(int64(1))

		})

		g.It("Permission test", func() {
			url := "/api/v1/courses/1"

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
