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
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

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

      w = tape.GetWithClaims(url, 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Should get a specific grade", func() {

      w := tape.GetWithClaims("/api/v1/courses/1/grades/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      grade_actual := &GradeResponse{}
      err := json.NewDecoder(w.Body).Decode(grade_actual)
      g.Assert(err).Equal(nil)

      hnd := helper.NewSubmissionFileHandle(grade_actual.SubmissionID)
      g.Assert(hnd.Exists()).Equal(false)
      grade_expected, err := stores.Grade.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(grade_actual.ID).Equal(grade_expected.ID)
      g.Assert(grade_actual.PublicExecutionState).Equal(grade_expected.PublicExecutionState)
      g.Assert(grade_actual.PrivateExecutionState).Equal(grade_expected.PrivateExecutionState)
      g.Assert(grade_actual.PublicTestLog).Equal(grade_expected.PublicTestLog)
      g.Assert(grade_actual.PrivateTestLog).Equal(grade_expected.PrivateTestLog)
      g.Assert(grade_actual.PublicTestStatus).Equal(grade_expected.PublicTestStatus)
      g.Assert(grade_actual.PrivateTestStatus).Equal(grade_expected.PrivateTestStatus)
      g.Assert(grade_actual.AcquiredPoints).Equal(grade_expected.AcquiredPoints)
      g.Assert(grade_actual.Feedback).Equal(grade_expected.Feedback)
      g.Assert(grade_actual.TutorID).Equal(grade_expected.TutorID)
      g.Assert(grade_actual.User.ID).Equal(grade_expected.UserID)
      g.Assert(grade_actual.User.FirstName).Equal(grade_expected.UserFirstName)
      g.Assert(grade_actual.User.LastName).Equal(grade_expected.UserLastName)
      g.Assert(grade_actual.User.Email).Equal(grade_expected.UserEmail)
      g.Assert(grade_actual.SubmissionID).Equal(grade_expected.SubmissionID)
      g.Assert(grade_actual.FileURL).Equal("")

      defer hnd.Delete()
      // now file exists
      src := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
      dest := fmt.Sprintf("%s/submissions/%s.zip", viper.GetString("uploads_dir"), strconv.FormatInt(grade_actual.SubmissionID, 10))
      copyFile(src, dest)

      g.Assert(hnd.Exists()).Equal(true)

      w = tape.GetWithClaims("/api/v1/courses/1/grades/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      err = json.NewDecoder(w.Body).Decode(grade_actual)
      g.Assert(err).Equal(nil)

      g.Assert(grade_actual.ID).Equal(grade_expected.ID)
      g.Assert(grade_actual.PublicExecutionState).Equal(grade_expected.PublicExecutionState)
      g.Assert(grade_actual.PrivateExecutionState).Equal(grade_expected.PrivateExecutionState)
      g.Assert(grade_actual.PublicTestLog).Equal(grade_expected.PublicTestLog)
      g.Assert(grade_actual.PrivateTestLog).Equal(grade_expected.PrivateTestLog)
      g.Assert(grade_actual.PublicTestStatus).Equal(grade_expected.PublicTestStatus)
      g.Assert(grade_actual.PrivateTestStatus).Equal(grade_expected.PrivateTestStatus)
      g.Assert(grade_actual.AcquiredPoints).Equal(grade_expected.AcquiredPoints)
      g.Assert(grade_actual.Feedback).Equal(grade_expected.Feedback)
      g.Assert(grade_actual.TutorID).Equal(grade_expected.TutorID)
      g.Assert(grade_actual.User.ID).Equal(grade_expected.UserID)
      g.Assert(grade_actual.User.FirstName).Equal(grade_expected.UserFirstName)
      g.Assert(grade_actual.User.LastName).Equal(grade_expected.UserLastName)
      g.Assert(grade_actual.User.Email).Equal(grade_expected.UserEmail)
      g.Assert(grade_actual.SubmissionID).Equal(grade_expected.SubmissionID)

      url := viper.GetString("url")

      g.Assert(grade_actual.FileURL).Equal(fmt.Sprintf("%s/api/v1/courses/1/submissions/1/file", url))

    })

    g.It("Should list all grades of a group", func() {
      url := "/api/v1/courses/1/grades?group_id=1"

      grades_expected, err := stores.Grade.GetFiltered(1, 0, 0, 1, 0, 0, "%%", -1, 0, 0, -1, -1)
      g.Assert(err).Equal(nil)

      w := tape.GetWithClaims(url, 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      grades_actual := []GradeResponse{}
      err = json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(grades_actual)).Equal(len(grades_expected))
    })

    g.It("Should list all grades of a group with some filters", func() {

      w := tape.GetWithClaims("/api/v1/courses/1/grades?group_id=1&public_test_status=0", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
      grades_actual := []GradeResponse{}
      err := json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)
      for _, el := range grades_actual {
        g.Assert(el.PublicTestStatus).Equal(0)
      }

      w = tape.GetWithClaims("/api/v1/courses/1/grades?group_id=1&private_test_status=0", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
      err = json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)
      for _, el := range grades_actual {
        g.Assert(el.PrivateTestStatus).Equal(0)
      }

      w = tape.GetWithClaims("/api/v1/courses/1/grades?group_id=1&tutor_id=3", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
      err = json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)
      for _, el := range grades_actual {
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
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // tutors
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 3, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_after, err := stores.Grade.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(entry_after.Feedback).Equal("Lorem Ipsum_update")
      g.Assert(entry_after.AcquiredPoints).Equal(3)
      g.Assert(entry_after.TutorID).Equal(int64(3))
    })

    g.It("Should perform updates when zero points", func() {

      data := H{
        "acquired_points": 0,
        "feedback":        "Lorem Ipsum_update",
      }

      w := tape.Put("/api/v1/courses/1/grades/1", data)
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // students
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      // tutors
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 3, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_after, err := stores.Grade.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(entry_after.Feedback).Equal("Lorem Ipsum_update")
      g.Assert(entry_after.AcquiredPoints).Equal(0)
      g.Assert(entry_after.TutorID).Equal(int64(3))
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
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 1, false)
      g.Assert(w.Code).Equal(http.StatusBadRequest)

      // tutors
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 3, false)
      g.Assert(w.Code).Equal(http.StatusBadRequest)

    })

    g.It("Should not perform updates (too many points)", func() {

      task, err := stores.Grade.IdentifyTaskOfGrade(1)
      g.Assert(err).Equal(nil)

      entry_before, err := stores.Grade.Get(1)
      g.Assert(err).Equal(nil)

      data := H{
        "acquired_points": task.MaxPoints + 10,
        "feedback":        "Lorem Ipsum_update",
      }

      w := tape.Put("/api/v1/courses/1/grades/1", data)
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      // students
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 1, false)
      g.Assert(w.Code).Equal(http.StatusBadRequest)

      // tutors
      w = tape.PutWithClaims("/api/v1/courses/1/grades/1", data, 3, false)
      g.Assert(w.Code).Equal(http.StatusBadRequest)

      entry_after, err := stores.Grade.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(entry_after.Feedback).Equal(entry_before.Feedback)
      g.Assert(entry_after.AcquiredPoints).Equal(entry_before.AcquiredPoints)
      g.Assert(entry_after.TutorID).Equal(entry_before.TutorID)
    })

    g.It("Should list missing grades", func() {
      grades_actual := []MissingGradeResponse{}
      // students have no missing data
      // but we do not know if a user is student in a course
      w := tape.GetWithClaims("/api/v1/courses/1/grades/missing", 112, false)
      g.Assert(w.Code).Equal(http.StatusOK)
      err := json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(grades_actual)).Equal(0)

      // admin (mock creates feed back for all submissions)
      w = tape.GetWithClaims("/api/v1/courses/1/grades/missing", 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)
      err = json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)

      expected_grades, err := stores.Grade.GetAllMissingGrades(1, 1, 0)
      g.Assert(err).Equal(nil)
      g.Assert(len(grades_actual)).Equal(len(expected_grades))

      // tutors (mock creates feed back for all submissions)
      w = tape.GetWithClaims("/api/v1/courses/1/grades/missing", 3, false)
      g.Assert(w.Code).Equal(http.StatusOK)
      err = json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)

      expected_grades, err = stores.Grade.GetAllMissingGrades(1, 3, 0)
      g.Assert(err).Equal(nil)
      g.Assert(len(grades_actual)).Equal(len(expected_grades))

      _, err = tape.DB.Exec("UPDATE grades SET feedback='' WHERE tutor_id = 3 ")
      g.Assert(err).Equal(nil)

      // tutors (mock creates feed back for all submissions)
      w = tape.GetWithClaims("/api/v1/courses/1/grades/missing", 3, false)
      g.Assert(w.Code).Equal(http.StatusOK)
      err = json.NewDecoder(w.Body).Decode(&grades_actual)
      g.Assert(err).Equal(nil)

      grades_expected, err := stores.Grade.GetAllMissingGrades(1, 3, 0)
      g.Assert(err).Equal(nil)

      // see mock.py
      g.Assert(len(grades_actual)).Equal(len(grades_expected))
      for k, el := range grades_actual {
        g.Assert(el.Grade.ID).Equal(grades_expected[k].Grade.ID)
        g.Assert(el.Grade.PublicExecutionState).Equal(grades_expected[k].Grade.PublicExecutionState)
        g.Assert(el.Grade.PrivateExecutionState).Equal(grades_expected[k].Grade.PrivateExecutionState)
        g.Assert(el.Grade.PublicTestLog).Equal(grades_expected[k].Grade.PublicTestLog)
        g.Assert(el.Grade.PrivateTestLog).Equal(grades_expected[k].Grade.PrivateTestLog)
        g.Assert(el.Grade.PublicTestStatus).Equal(grades_expected[k].Grade.PublicTestStatus)
        g.Assert(el.Grade.PrivateTestStatus).Equal(grades_expected[k].Grade.PrivateTestStatus)
        g.Assert(el.Grade.AcquiredPoints).Equal(grades_expected[k].Grade.AcquiredPoints)
        g.Assert(el.Grade.PrivateTestLog).Equal(grades_expected[k].Grade.PrivateTestLog)
        g.Assert(el.Grade.Feedback).Equal("")
        g.Assert(el.Grade.TutorID).Equal(int64(3))
        g.Assert(el.Grade.SubmissionID).Equal(grades_expected[k].Grade.SubmissionID)

        g.Assert(el.Grade.User.ID).Equal(grades_expected[k].Grade.UserID)
        g.Assert(el.Grade.User.FirstName).Equal(grades_expected[k].Grade.UserFirstName)
        g.Assert(el.Grade.User.LastName).Equal(grades_expected[k].Grade.UserLastName)
        g.Assert(el.Grade.User.Email).Equal(grades_expected[k].Grade.UserEmail)
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
      w = tape.PostWithClaims(url, data, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.PostWithClaims(url, data, 3, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PostWithClaims(url, data, 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_after, err := stores.Grade.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(entry_after.PublicTestLog).Equal("some new logs")
      g.Assert(entry_after.PublicTestStatus).Equal(2)

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
      w = tape.PostWithClaims(url, data, 112, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // tutors
      w = tape.PostWithClaims(url, data, 3, false)
      g.Assert(w.Code).Equal(http.StatusForbidden)

      // admin
      w = tape.PostWithClaims(url, data, 1, false)
      g.Assert(w.Code).Equal(http.StatusOK)

      entry_after, err := stores.Grade.Get(1)
      g.Assert(err).Equal(nil)

      g.Assert(entry_after.PrivateTestLog).Equal("some new logs")
      g.Assert(entry_after.PrivateTestStatus).Equal(2)

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

      task2, err := stores.Task.Create(&model.Task{
        Name:               "2",
        MaxPoints:          31,
        PublicDockerImage:  null.StringFrom("ff"),
        PrivateDockerImage: null.StringFrom("ff"),
      }, sheet2.ID)

      _ = task1
      _ = task2

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
      w := tape.GetWithClaims("/api/v1/courses/1/grades/summary", 1, true)
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
      w = tape.GetWithClaims("/api/v1/courses/1/grades/summary", 1, true)
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
      w = tape.GetWithClaims("/api/v1/courses/1/grades/summary", 1, true)
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
      w = tape.GetWithClaims("/api/v1/courses/1/grades/summary", 1, true)
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
