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
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/franela/goblin"
	"github.com/spf13/viper"
	null "gopkg.in/guregu/null.v3"
)

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func TestGrade(t *testing.T) {
	PrepareTests()
	g := goblin.Goblin(t)
	email.DefaultMail = email.VoidMail

	tape := &Tape{}

	var stores *Stores

	studentJWT := NewJWTRequest(112, false)
	tutorJWT := NewJWTRequest(2, false)
	adminJWT := NewJWTRequest(1, true)
	noAdminJWT := NewJWTRequest(1, false)

	g.Describe("Grade", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
			stores = NewStores(tape.DB)
			_ = stores
		})

		g.It("Query should require access claims", func() {
			url := "/api/v1/courses/1/grades?group_id=1"
			w := tape.Get(url)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get(url, adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Should get a specific grade", func() {

			w := tape.Get("/api/v1/courses/1/grades/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			gradeActual := &GradeResponse{}
			err := json.NewDecoder(w.Body).Decode(gradeActual)
			g.Assert(err).Equal(nil)

			hnd := helper.NewSubmissionFileHandle(gradeActual.SubmissionID)
			g.Assert(hnd.Exists()).Equal(false)
			gradeExpected, err := stores.Grade.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(gradeActual.ID).Equal(gradeExpected.ID)
			g.Assert(gradeActual.PublicExecutionState).Equal(gradeExpected.PublicExecutionState)
			g.Assert(gradeActual.PrivateExecutionState).Equal(gradeExpected.PrivateExecutionState)
			g.Assert(gradeActual.PublicTestLog).Equal(gradeExpected.PublicTestLog)
			g.Assert(gradeActual.PrivateTestLog).Equal(gradeExpected.PrivateTestLog)
			g.Assert(gradeActual.PublicTestStatus).Equal(gradeExpected.PublicTestStatus)
			g.Assert(gradeActual.PrivateTestStatus).Equal(gradeExpected.PrivateTestStatus)
			g.Assert(gradeActual.AcquiredPoints).Equal(gradeExpected.AcquiredPoints)
			g.Assert(gradeActual.Feedback).Equal(gradeExpected.Feedback)
			g.Assert(gradeActual.TutorID).Equal(gradeExpected.TutorID)
			g.Assert(gradeActual.User.ID).Equal(gradeExpected.UserID)
			g.Assert(gradeActual.User.FirstName).Equal(gradeExpected.UserFirstName)
			g.Assert(gradeActual.User.LastName).Equal(gradeExpected.UserLastName)
			g.Assert(gradeActual.User.Email).Equal(gradeExpected.UserEmail)
			g.Assert(gradeActual.SubmissionID).Equal(gradeExpected.SubmissionID)
			g.Assert(gradeActual.FileURL).Equal("")

			defer hnd.Delete()
			// now file exists
			src := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
			dest := fmt.Sprintf("%s/submissions/%s.zip", viper.GetString("uploads_dir"), strconv.FormatInt(gradeActual.SubmissionID, 10))
			copyFile(src, dest)

			g.Assert(hnd.Exists()).Equal(true)

			w = tape.Get("/api/v1/courses/1/grades/1", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			err = json.NewDecoder(w.Body).Decode(gradeActual)
			g.Assert(err).Equal(nil)

			g.Assert(gradeActual.ID).Equal(gradeExpected.ID)
			g.Assert(gradeActual.PublicExecutionState).Equal(gradeExpected.PublicExecutionState)
			g.Assert(gradeActual.PrivateExecutionState).Equal(gradeExpected.PrivateExecutionState)
			g.Assert(gradeActual.PublicTestLog).Equal(gradeExpected.PublicTestLog)
			g.Assert(gradeActual.PrivateTestLog).Equal(gradeExpected.PrivateTestLog)
			g.Assert(gradeActual.PublicTestStatus).Equal(gradeExpected.PublicTestStatus)
			g.Assert(gradeActual.PrivateTestStatus).Equal(gradeExpected.PrivateTestStatus)
			g.Assert(gradeActual.AcquiredPoints).Equal(gradeExpected.AcquiredPoints)
			g.Assert(gradeActual.Feedback).Equal(gradeExpected.Feedback)
			g.Assert(gradeActual.TutorID).Equal(gradeExpected.TutorID)
			g.Assert(gradeActual.User.ID).Equal(gradeExpected.UserID)
			g.Assert(gradeActual.User.FirstName).Equal(gradeExpected.UserFirstName)
			g.Assert(gradeActual.User.LastName).Equal(gradeExpected.UserLastName)
			g.Assert(gradeActual.User.Email).Equal(gradeExpected.UserEmail)
			g.Assert(gradeActual.SubmissionID).Equal(gradeExpected.SubmissionID)

			url := viper.GetString("url")

			g.Assert(gradeActual.FileURL).Equal(fmt.Sprintf("%s/api/v1/courses/1/submissions/1/file", url))

		})

		g.It("Should list all grades of a group", func() {
			url := "/api/v1/courses/1/grades?group_id=1"

			gradesExpected, err := stores.Grade.GetFiltered(1, 0, 0, 1, 0, 0, "%%", -1, -1, -1, -1, -1)
			g.Assert(err).Equal(nil)

			w := tape.Get(url, adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			gradesActual := []GradeResponse{}
			err = json.NewDecoder(w.Body).Decode(&gradesActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(gradesActual)).Equal(len(gradesExpected))
		})

		g.It("Should list all grades of a group with some filters", func() {

			w := tape.Get("/api/v1/courses/1/grades?group_id=1&public_test_status=0", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			gradesActual := []GradeResponse{}
			err := json.NewDecoder(w.Body).Decode(&gradesActual)
			g.Assert(err).Equal(nil)
			for _, el := range gradesActual {
				g.Assert(el.PublicTestStatus).Equal(0)
			}

			w = tape.Get("/api/v1/courses/1/grades?group_id=1&private_test_status=0", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&gradesActual)
			g.Assert(err).Equal(nil)
			for _, el := range gradesActual {
				g.Assert(el.PrivateTestStatus).Equal(0)
			}

			w = tape.Get("/api/v1/courses/1/grades?group_id=1&tutor_id=3", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&gradesActual)
			g.Assert(err).Equal(nil)
			for _, el := range gradesActual {
				g.Assert(el.TutorID).Equal(int64(3))
			}
		})

		g.It("Should perform updates", func() {

			data := H{
				"acquired_points": 3,
				"feedback":        "Lorem Ipsum_update",
			}

			w := tape.Put("/api/v1/courses/1/grades/1", data)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// students
			w = tape.Put("/api/v1/courses/1/grades/1", data, studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Put("/api/v1/courses/1/grades/1", data, noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// tutors
			w = tape.Put("/api/v1/courses/1/grades/1", data, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			entryAfter, err := stores.Grade.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(entryAfter.Feedback).Equal("Lorem Ipsum_update")
			g.Assert(entryAfter.AcquiredPoints).Equal(3)
			g.Assert(entryAfter.TutorID).Equal(tutorJWT.Claims.LoginID)
		})

		g.It("Should perform updates when zero points", func() {

			data := H{
				"acquired_points": 0,
				"feedback":        "Lorem Ipsum_update",
			}

			w := tape.Put("/api/v1/courses/1/grades/1", data)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// students
			w = tape.Put("/api/v1/courses/1/grades/1", data, studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Put("/api/v1/courses/1/grades/1", data, noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			// tutors
			w = tape.Put("/api/v1/courses/1/grades/1", data, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			entryAfter, err := stores.Grade.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(entryAfter.Feedback).Equal("Lorem Ipsum_update")
			g.Assert(entryAfter.AcquiredPoints).Equal(0)
			g.Assert(entryAfter.TutorID).Equal(tutorJWT.Claims.LoginID)
		})

		g.Xit("Should not perform updates when missing points", func() {
			// todo difference between "0" and None
			data := H{
				// "acquired_points": 0,
				"feedback": "Lorem Ipsum_update",
			}

			w := tape.Put("/api/v1/courses/1/grades/1", data)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// students
			w = tape.Put("/api/v1/courses/1/grades/1", data, studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Put("/api/v1/courses/1/grades/1", data, noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			// tutors
			w = tape.Put("/api/v1/courses/1/grades/1", data, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

		})

		g.It("Should not perform updates (too many points)", func() {

			task, err := stores.Grade.IdentifyTaskOfGrade(1)
			g.Assert(err).Equal(nil)

			entryBefore, err := stores.Grade.Get(1)
			g.Assert(err).Equal(nil)

			data := H{
				"acquired_points": task.MaxPoints + 10,
				"feedback":        "Lorem Ipsum_update",
			}

			w := tape.Put("/api/v1/courses/1/grades/1", data)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// students
			w = tape.Put("/api/v1/courses/1/grades/1", data, studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Put("/api/v1/courses/1/grades/1", data, noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			// tutors
			w = tape.Put("/api/v1/courses/1/grades/1", data, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			entryAfter, err := stores.Grade.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(entryAfter.Feedback).Equal(entryBefore.Feedback)
			g.Assert(entryAfter.AcquiredPoints).Equal(entryBefore.AcquiredPoints)
			g.Assert(entryAfter.TutorID).Equal(entryBefore.TutorID)
		})

		g.It("Should list missing grades", func() {
			gradesActual := []MissingGradeResponse{}
			// students have no missing data
			// but we do not know if a user is student in a course
			w := tape.Get("/api/v1/courses/1/grades/missing", studentJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			err := json.NewDecoder(w.Body).Decode(&gradesActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(gradesActual)).Equal(0)

			// admin (mock creates feed back for all submissions)
			w = tape.Get("/api/v1/courses/1/grades/missing", noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&gradesActual)
			g.Assert(err).Equal(nil)

			gradesExpected, err := stores.Grade.GetAllMissingGrades(1, noAdminJWT.Claims.LoginID, 0)
			g.Assert(err).Equal(nil)
			g.Assert(len(gradesActual)).Equal(len(gradesExpected))

			// tutors (mock creates feed back for all submissions)
			w = tape.Get("/api/v1/courses/1/grades/missing", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&gradesActual)
			g.Assert(err).Equal(nil)

			gradesExpected, err = stores.Grade.GetAllMissingGrades(1, tutorJWT.Claims.LoginID, 0)
			g.Assert(err).Equal(nil)
			g.Assert(len(gradesActual)).Equal(len(gradesExpected))

			_, err = tape.DB.Exec("UPDATE grades SET feedback='' WHERE tutor_id = 3 ")
			g.Assert(err).Equal(nil)

			// tutors (mock creates feed back for all submissions)
			w = tape.Get("/api/v1/courses/1/grades/missing", tutorJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&gradesActual)
			g.Assert(err).Equal(nil)

			gradesExpected, err = stores.Grade.GetAllMissingGrades(1, tutorJWT.Claims.LoginID, 0)
			g.Assert(err).Equal(nil)

			// see mock.py
			g.Assert(len(gradesActual)).Equal(len(gradesExpected))
			for k, el := range gradesActual {
				g.Assert(el.Grade.ID).Equal(gradesExpected[k].Grade.ID)
				g.Assert(el.Grade.PublicExecutionState).Equal(gradesExpected[k].Grade.PublicExecutionState)
				g.Assert(el.Grade.PrivateExecutionState).Equal(gradesExpected[k].Grade.PrivateExecutionState)
				g.Assert(el.Grade.PublicTestLog).Equal(gradesExpected[k].Grade.PublicTestLog)
				g.Assert(el.Grade.PrivateTestLog).Equal(gradesExpected[k].Grade.PrivateTestLog)
				g.Assert(el.Grade.PublicTestStatus).Equal(gradesExpected[k].Grade.PublicTestStatus)
				g.Assert(el.Grade.PrivateTestStatus).Equal(gradesExpected[k].Grade.PrivateTestStatus)
				g.Assert(el.Grade.AcquiredPoints).Equal(gradesExpected[k].Grade.AcquiredPoints)
				g.Assert(el.Grade.PrivateTestLog).Equal(gradesExpected[k].Grade.PrivateTestLog)
				g.Assert(el.Grade.Feedback).Equal("")
				g.Assert(el.Grade.TutorID).Equal(tutorJWT.Claims.LoginID)
				g.Assert(el.Grade.SubmissionID).Equal(gradesExpected[k].Grade.SubmissionID)

				g.Assert(el.Grade.User.ID).Equal(gradesExpected[k].Grade.UserID)
				g.Assert(el.Grade.User.FirstName).Equal(gradesExpected[k].Grade.UserFirstName)
				g.Assert(el.Grade.User.LastName).Equal(gradesExpected[k].Grade.UserLastName)
				g.Assert(el.Grade.User.Email).Equal(gradesExpected[k].Grade.UserEmail)
			}
		})

		g.It("Should handle feedback from public tests", func() {

			url := "/api/v1/courses/1/grades/1/public_result"

			data := H{
				"log":    "some new logs",
				"status": 2,
			}

			w := tape.Post(url, data)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// students
			w = tape.Post(url, data, studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Post(url, data, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Post(url, data, noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			entryAfter, err := stores.Grade.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(entryAfter.PublicTestLog).Equal("some new logs")
			g.Assert(entryAfter.PublicTestStatus).Equal(2)

		})

		g.It("Should handle feedback from private tests", func() {

			url := "/api/v1/courses/1/grades/1/private_result"

			data := H{
				"log":    "some new logs",
				"status": 2,
			}

			w := tape.Post(url, data)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// students
			w = tape.Post(url, data, studentJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.Post(url, data, tutorJWT)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.Post(url, data, noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			entryAfter, err := stores.Grade.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(entryAfter.PrivateTestLog).Equal("some new logs")
			g.Assert(entryAfter.PrivateTestStatus).Equal(2)

		})

		g.It("Should show correct overview", func() {

			course, err := stores.Course.Get(1)
			g.Assert(err).Equal(nil)

			_, err = tape.DB.Exec("TRUNCATE TABLE tasks CASCADE;")
			g.Assert(err).Equal(nil)

			_, err = tape.DB.Exec("TRUNCATE TABLE sheets CASCADE;")
			g.Assert(err).Equal(nil)

			_, err = tape.DB.Exec("TRUNCATE TABLE task_sheet CASCADE;")
			g.Assert(err).Equal(nil)

			_, err = tape.DB.Exec("TRUNCATE TABLE sheet_course CASCADE;")
			g.Assert(err).Equal(nil)

			_, err = tape.DB.Exec("TRUNCATE TABLE grades CASCADE;")
			g.Assert(err).Equal(nil)

			_, err = tape.DB.Exec("TRUNCATE TABLE submissions CASCADE;")
			g.Assert(err).Equal(nil)

			tasks, err := stores.Task.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(tasks)).Equal(0)

			// create Sheets in database
			sheet1, err := stores.Sheet.Create(&model.Sheet{
				Name:      "1",
				PublishAt: NowUTC(),
				DueAt:     NowUTC(),
			}, course.ID)
			g.Assert(err).Equal(nil)

			sheet2, err := stores.Sheet.Create(&model.Sheet{
				Name:      "2",
				PublishAt: NowUTC(),
				DueAt:     NowUTC(),
			}, course.ID)
			g.Assert(err).Equal(nil)

			// fmt.Println("sheet 1", sheet1.ID)
			// fmt.Println("sheet 2", sheet2.ID)

			// create tasks
			task1, err := stores.Task.Create(&model.Task{
				Name:               "1",
				MaxPoints:          30,
				PublicDockerImage:  null.StringFrom("ff"),
				PrivateDockerImage: null.StringFrom("ff"),
			}, sheet1.ID)
			g.Assert(err).Equal(nil)

			task2, err := stores.Task.Create(&model.Task{
				Name:               "2",
				MaxPoints:          31,
				PublicDockerImage:  null.StringFrom("ff"),
				PrivateDockerImage: null.StringFrom("ff"),
			}, sheet2.ID)
			g.Assert(err).Equal(nil)

			uid1 := int64(42)
			uid2 := int64(43)

			user1, err := stores.User.Get(uid1)
			g.Assert(err).Equal(nil)
			user2, err := stores.User.Get(uid2)
			g.Assert(err).Equal(nil)

			sub1, err := stores.Submission.Create(&model.Submission{UserID: uid1, TaskID: task1.ID})
			g.Assert(err).Equal(nil)

			// test empty grades
			response := GradeOverviewResponse{}
			w := tape.Get("/api/v1/courses/1/grades/summary", adminJWT)
			// fmt.Println(w.Body)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&response)
			g.Assert(err).Equal(nil)
			g.Assert(len(response.Sheets)).Equal(2)
			g.Assert(len(response.Achievements)).Equal(0)

			grade1 := &model.Grade{
				PublicExecutionState:  0,
				PrivateExecutionState: 0,
				PublicTestLog:         "empty",
				PrivateTestLog:        "empty",
				PublicTestStatus:      0,
				PrivateTestStatus:     0,
				AcquiredPoints:        5,
				Feedback:              "",
				TutorID:               1,
				SubmissionID:          sub1.ID,
			}

			_, err = stores.Grade.Create(grade1)
			g.Assert(err).Equal(nil)

			p := []model.Grade{}
			err = tape.DB.Select(&p, "SELECT * FROM grades;")
			g.Assert(err).Equal(nil)
			g.Assert(len(p)).Equal(1)

			response = GradeOverviewResponse{}
			w = tape.Get("/api/v1/courses/1/grades/summary", adminJWT)
			// fmt.Println(w.Body)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&response)
			g.Assert(err).Equal(nil)
			g.Assert(len(response.Sheets)).Equal(2)
			g.Assert(len(response.Achievements)).Equal(1)
			g.Assert(len(response.Achievements[0].Points)).Equal(2)
			g.Assert(response.Achievements[0].Points[0]).Equal(grade1.AcquiredPoints)
			g.Assert(response.Achievements[0].Points[1]).Equal(0)
			g.Assert(response.Achievements[0].User.ID).Equal(user1.ID)
			g.Assert(response.Achievements[0].User.Email).Equal(user1.Email)

			//  ---------------------
			sub2, err := stores.Submission.Create(&model.Submission{UserID: uid2, TaskID: task2.ID})
			g.Assert(err).Equal(nil)
			grade2 := &model.Grade{
				PublicExecutionState:  0,
				PrivateExecutionState: 0,
				PublicTestLog:         "empty",
				PrivateTestLog:        "empty",
				PublicTestStatus:      0,
				PrivateTestStatus:     0,
				AcquiredPoints:        7,
				Feedback:              "",
				TutorID:               1,
				SubmissionID:          sub2.ID,
			}

			_, err = stores.Grade.Create(grade2)
			g.Assert(err).Equal(nil)

			p = []model.Grade{}
			err = tape.DB.Select(&p, "SELECT * FROM grades;")
			g.Assert(err).Equal(nil)
			g.Assert(len(p)).Equal(2)

			response = GradeOverviewResponse{}
			w = tape.Get("/api/v1/courses/1/grades/summary", adminJWT)
			// fmt.Println(w.Body)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&response)
			g.Assert(err).Equal(nil)
			g.Assert(len(response.Sheets)).Equal(2)
			g.Assert(len(response.Achievements)).Equal(2)
			g.Assert(len(response.Achievements[0].Points)).Equal(2)
			g.Assert(response.Achievements[0].Points[0]).Equal(grade1.AcquiredPoints)
			g.Assert(response.Achievements[0].Points[1]).Equal(0)
			g.Assert(response.Achievements[0].User.ID).Equal(user1.ID)
			g.Assert(response.Achievements[0].User.Email).Equal(user1.Email)

			g.Assert(len(response.Achievements[1].Points)).Equal(2)
			g.Assert(response.Achievements[1].Points[0]).Equal(0)
			g.Assert(response.Achievements[1].Points[1]).Equal(grade2.AcquiredPoints)
			g.Assert(response.Achievements[1].User.ID).Equal(user2.ID)
			g.Assert(response.Achievements[1].User.Email).Equal(user2.Email)

			//  ---------------------
			sub3, err := stores.Submission.Create(&model.Submission{UserID: uid2, TaskID: task1.ID})
			g.Assert(err).Equal(nil)
			grade3 := &model.Grade{
				PublicExecutionState:  0,
				PrivateExecutionState: 0,
				PublicTestLog:         "empty",
				PrivateTestLog:        "empty",
				PublicTestStatus:      0,
				PrivateTestStatus:     0,
				AcquiredPoints:        8,
				Feedback:              "",
				TutorID:               1,
				SubmissionID:          sub3.ID,
			}

			_, err = stores.Grade.Create(grade3)
			g.Assert(err).Equal(nil)

			p = []model.Grade{}
			err = tape.DB.Select(&p, "SELECT * FROM grades;")
			g.Assert(err).Equal(nil)
			g.Assert(len(p)).Equal(3)

			response = GradeOverviewResponse{}
			w = tape.Get("/api/v1/courses/1/grades/summary", adminJWT)
			// fmt.Println(w.Body)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&response)
			g.Assert(err).Equal(nil)
			g.Assert(len(response.Sheets)).Equal(2)
			g.Assert(len(response.Achievements)).Equal(2)
			g.Assert(len(response.Achievements[0].Points)).Equal(2)
			g.Assert(response.Achievements[0].Points[0]).Equal(grade1.AcquiredPoints)
			g.Assert(response.Achievements[0].Points[1]).Equal(0)
			g.Assert(response.Achievements[0].User.ID).Equal(user1.ID)
			g.Assert(response.Achievements[0].User.Email).Equal(user1.Email)

			g.Assert(len(response.Achievements[1].Points)).Equal(2)
			g.Assert(response.Achievements[1].Points[0]).Equal(grade3.AcquiredPoints)
			g.Assert(response.Achievements[1].Points[1]).Equal(grade2.AcquiredPoints)
			g.Assert(response.Achievements[1].User.ID).Equal(user2.ID)
			g.Assert(response.Achievements[1].User.Email).Equal(user2.Email)

		})

		g.AfterEach(func() {
			tape.AfterEach()
		})
	})

}
