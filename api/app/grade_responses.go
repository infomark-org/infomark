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
  "fmt"
  "net/http"
  "strconv"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/auth/authorize"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/render"
  "github.com/spf13/viper"
)

// .............................................................................

// GradeResponse is the response payload for Grade management.
type GradeResponse struct {
  ID                    int64  `json:"id" example:"1"`
  PublicExecutionState  int    `json:"public_execution_state" example:"1"`
  PrivateExecutionState int    `json:"private_execution_state" example:"1"`
  PublicTestLog         string `json:"public_test_log" example:"Lorem Ipsum"`
  PrivateTestLog        string `json:"private_test_log" example:"Lorem Ipsum"`
  PublicTestStatus      int    `json:"public_test_status" example:"1"`
  PrivateTestStatus     int    `json:"private_test_status" example:"0"`
  AcquiredPoints        int    `json:"acquired_points" example:"19"`
  Feedback              string `json:"feedback" example:"Some feedback"`
  TutorID               int64  `json:"tutor_id" example:"2"`
  SubmissionID          int64  `json:"submission_id" example:"31"`
  FileURL               string `json:"file_url" example:"/api/v1/submissions/61/file"`
  User                  *struct {
    ID        int64  `json:"id" example:"1"`
    FirstName string `json:"first_name" example:"Max"`
    LastName  string `json:"last_name" example:"Mustermensch"`
    Email     string `json:"email" example:"test@unit-tuebingen.de"`
  } `json:"user"`
}

// Render post-processes a GradeResponse.
func (body *GradeResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// newGradeResponse creates a response from a Grade model.
func newGradeResponse(p *model.Grade, courseID int64) *GradeResponse {

  fileURL := ""
  if helper.NewSubmissionFileHandle(p.ID).Exists() {
    fileURL = fmt.Sprintf("%s/api/v1/courses/%d/submissions/%d/file",
      viper.GetString("url"),
      courseID,
      p.SubmissionID,
    )
  }

  user := &struct {
    ID        int64  `json:"id" example:"1"`
    FirstName string `json:"first_name" example:"Max"`
    LastName  string `json:"last_name" example:"Mustermensch"`
    Email     string `json:"email" example:"test@unit-tuebingen.de"`
  }{
    ID:        p.UserID,
    FirstName: p.UserFirstName,
    LastName:  p.UserLastName,
    Email:     p.UserEmail,
  }

  return &GradeResponse{
    ID:                    p.ID,
    PublicExecutionState:  p.PublicExecutionState,
    PrivateExecutionState: p.PrivateExecutionState,
    PublicTestLog:         p.PublicTestLog,
    PrivateTestLog:        p.PrivateTestLog,
    PublicTestStatus:      p.PublicTestStatus,
    PrivateTestStatus:     p.PrivateTestStatus,
    AcquiredPoints:        p.AcquiredPoints,
    Feedback:              p.Feedback,
    TutorID:               p.TutorID,
    User:                  user,
    SubmissionID:          p.SubmissionID,
    FileURL:               fileURL,
  }
}

// newGradeListResponse creates a response from a list of Grade models.
func newGradeListResponse(Grades []model.Grade, courseID int64) []render.Renderer {
  list := []render.Renderer{}
  for k := range Grades {
    list = append(list, newGradeResponse(&Grades[k], courseID))
  }
  return list
}

// GradeResponse is the response payload for Grade management.
type MissingGradeResponse struct {
  Grade *struct {
    ID                    int64  `json:"id" example:"1"`
    PublicExecutionState  int    `json:"public_execution_state" example:"1"`
    PrivateExecutionState int    `json:"private_execution_state" example:"1"`
    PublicTestLog         string `json:"public_test_log" example:"Lorem Ipsum"`
    PrivateTestLog        string `json:"private_test_log" example:"Lorem Ipsum"`
    PublicTestStatus      int    `json:"public_test_status" example:"1"`
    PrivateTestStatus     int    `json:"private_test_status" example:"0"`
    AcquiredPoints        int    `json:"acquired_points" example:"19"`
    Feedback              string `json:"feedback" example:"Some feedback"`
    TutorID               int64  `json:"tutor_id" example:"2"`
    SubmissionID          int64  `json:"submission_id" example:"31"`
    FileURL               string `json:"file_url" example:"/api/v1/submissions/61/file"`
    User                  *struct {
      ID        int64  `json:"id" example:"1"`
      FirstName string `json:"first_name" example:"Max"`
      LastName  string `json:"last_name" example:"Mustermensch"`
      Email     string `json:"email" example:"test@unit-tuebingen.de"`
    } `json:"user"`
  } `json:"grade"`
  CourseID int64 `json:"course_id" example:"1"`
  SheetID  int64 `json:"sheet_id" example:"10"`
  TaskID   int64 `json:"task_id" example:"2"`
}

// Render post-processes a MissingGradeResponse.
func (body *MissingGradeResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// newMissingGradeResponse creates a response from a Grade model.
func newMissingGradeResponse(p *model.MissingGrade) *MissingGradeResponse {
  fileURL := ""
  if helper.NewSubmissionFileHandle(p.SubmissionID).Exists() {
    fileURL = fmt.Sprintf("/api/v1/submissions/%s/file", strconv.FormatInt(p.SubmissionID, 10))
  }

  user := &struct {
    ID        int64  `json:"id" example:"1"`
    FirstName string `json:"first_name" example:"Max"`
    LastName  string `json:"last_name" example:"Mustermensch"`
    Email     string `json:"email" example:"test@unit-tuebingen.de"`
  }{
    ID:        p.UserID,
    FirstName: p.UserFirstName,
    LastName:  p.UserLastName,
    Email:     p.UserEmail,
  }

  grade := &struct {
    ID                    int64  `json:"id" example:"1"`
    PublicExecutionState  int    `json:"public_execution_state" example:"1"`
    PrivateExecutionState int    `json:"private_execution_state" example:"1"`
    PublicTestLog         string `json:"public_test_log" example:"Lorem Ipsum"`
    PrivateTestLog        string `json:"private_test_log" example:"Lorem Ipsum"`
    PublicTestStatus      int    `json:"public_test_status" example:"1"`
    PrivateTestStatus     int    `json:"private_test_status" example:"0"`
    AcquiredPoints        int    `json:"acquired_points" example:"19"`
    Feedback              string `json:"feedback" example:"Some feedback"`
    TutorID               int64  `json:"tutor_id" example:"2"`
    SubmissionID          int64  `json:"submission_id" example:"31"`
    FileURL               string `json:"file_url" example:"/api/v1/submissions/61/file"`
    User                  *struct {
      ID        int64  `json:"id" example:"1"`
      FirstName string `json:"first_name" example:"Max"`
      LastName  string `json:"last_name" example:"Mustermensch"`
      Email     string `json:"email" example:"test@unit-tuebingen.de"`
    } `json:"user"`
  }{
    ID:                    p.ID,
    PublicExecutionState:  p.PublicExecutionState,
    PrivateExecutionState: p.PrivateExecutionState,
    PublicTestLog:         p.PublicTestLog,
    PrivateTestLog:        p.PrivateTestLog,
    PublicTestStatus:      p.PublicTestStatus,
    PrivateTestStatus:     p.PrivateTestStatus,
    AcquiredPoints:        p.AcquiredPoints,
    Feedback:              p.Feedback,
    TutorID:               p.TutorID,
    User:                  user,
    SubmissionID:          p.SubmissionID,
    FileURL:               fileURL,
  }

  r := &MissingGradeResponse{
    Grade:    grade,
    CourseID: p.CourseID,
    SheetID:  p.SheetID,
    TaskID:   p.TaskID,
  }

  return r

}

// newMissingGradeListResponse creates a response from a list of Grade models.
func newMissingGradeListResponse(Grades []model.MissingGrade) []render.Renderer {
  list := []render.Renderer{}
  for k := range Grades {
    list = append(list, newMissingGradeResponse(&Grades[k]))
  }
  return list
}

// // MaterialResponse is the response payload for Material management.
// type GradeOverviewResponse struct {
//   UserID  int64  `json:"user_id" example:"112"`
//   SheetID int64  `json:"sheet_id" example:"1"`
//   Name    string `json:"name" example:"Sheet 8"`
//   Points  int64  `json:"points" example:"4"`
// }

// // newGradeOverviewResponse creates a response from a Material model.
// func newGradeOverviewResponse(p *model.OverviewGrade) *GradeOverviewResponse {
//   return &GradeOverviewResponse{
//     UserID:  p.UserID,
//     SheetID: p.SheetID,
//     Name:    p.Name,
//     Points:  p.Points,
//   }
// }

// // newGradeOverviewListResponse creates a response from a list of Material models.
// func newGradeOverviewListResponse(collection []model.OverviewGrade) []render.Renderer {
//   list := []render.Renderer{}
//   for k := range collection {
//     list = append(list, newGradeOverviewResponse(&collection[k]))
//   }
//   return list
// }

// // Render post-processes a GradeOverviewResponse.
// func (body *GradeOverviewResponse) Render(w http.ResponseWriter, r *http.Request) error {
//   return nil
// }

// for the swagger build relying on go.ast we need to duplicate code here
type sheetInfo struct {
  ID   int64  `json:"id" example:"42"`
  Name string `json:"name" example:"sheet 0"`
}

type userInfo struct {
  ID            int64  `json:"id" example:"42"`
  FirstName     string `json:"first_name" example:"max"`
  LastName      string `json:"last_name" example:"mustermensch"`
  StudentNumber string `json:"student_number" example:"0815"`
}

type achievementInfo struct {
  User   userInfo `json:"user_info" example:""`
  Points []int    `json:"points" example:"4"`
}

type GradeOverviewResponse struct {
  Sheets       []sheetInfo       `json:"sheets" example:""`
  Achievements []achievementInfo `json:"achievements" example:""`
}

// newGradeOverviewResponse creates a response from a Material model.
func newGradeOverviewResponse(collection []model.OverviewGrade, sheets []model.Sheet, role authorize.CourseRole) *GradeOverviewResponse {
  obj := &GradeOverviewResponse{}
  // collection is sorted by user_id

  // only do this once
  for _, s := range sheets {
    obj.Sheets = append(obj.Sheets, sheetInfo{s.ID, s.Name})
  }

  if len(collection) > 0 {

    oldUserID := collection[0].UserID
    currentPoints := []int{}
    // sptr := 0 // sheet pointer
    eptr := 0 // entry pointer

    // iterate collection of users
    // {user, sheet, points}
    // {user, sheet, points}
    // This is sparse: Students without submissions for one sheet are not listed.
    // We need to explicitly list them here with "0" points.
    for eptr < len(collection) {

      // new student
      if collection[eptr].UserID != oldUserID {
        // reset
        currentPoints = []int{}
        // sptr = 0
        oldUserID = collection[eptr].UserID
      }

      for sptr := 0; sptr < len(sheets); sptr++ {

        if eptr < len(collection) {

          if collection[eptr].SheetID == sheets[sptr].ID {
            currentPoints = append(currentPoints, collection[eptr].Points)

            // we safely jumpt to next entry
            if eptr+1 < len(collection) {
              if collection[eptr].UserID == collection[eptr+1].UserID {
                eptr++
              } else {
                break
              }
            }

          } else if collection[eptr].SheetID > sheets[sptr].ID {
            currentPoints = append(currentPoints, int(0))
          }

        }

      }

      for i := 0; i < len(obj.Sheets)-len(currentPoints); i++ {
        currentPoints = append(currentPoints, int(0))
      }

      if len(currentPoints) > 0 {
        user := userInfo{
          ID:            collection[eptr].UserID,
          FirstName:     collection[eptr].UserFirstName,
          LastName:      collection[eptr].UserLastName,
          StudentNumber: collection[eptr].UserStudentNumber,
        }

        if role == authorize.TUTOR {
          user.StudentNumber = ""
        }

        eptr = eptr + 1
        obj.Achievements = append(obj.Achievements, achievementInfo{user, currentPoints})
      }
    }

  }
  return obj
}

// Render post-processes a GradeOverviewResponse.
func (body *GradeOverviewResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}
