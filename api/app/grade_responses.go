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
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/render"
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
func newGradeResponse(p *model.Grade) *GradeResponse {

  fileURL := ""
  if helper.NewSubmissionFileHandle(p.ID).Exists() {
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
func newGradeListResponse(Grades []model.Grade) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range Grades {
    list = append(list, newGradeResponse(&Grades[k]))
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
