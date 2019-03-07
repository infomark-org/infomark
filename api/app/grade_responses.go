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
  "net/http"

  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/render"
)

// func ExecutionStateToString(state int) string {
//   switch state {
//   case 0:
//     return "pending"
//   case 1:
//     return "running"
//   case 2:
//     return "finished"
//   default:
//     return "pending"
//   }
// }

// func TestStatusToString(state int) string {
//   switch state {
//   case 0:
//     return "passed"
//   case 1:
//     return "failed"
//   default:
//     return "failed"
//   }
// }

// .............................................................................

// GradeResponse is the response payload for Grade management.
type GradeResponse struct {
  ID                int64  `json:"id" example:"1"`
  ExecutionState    int    `json:"execution_state" example:"1"`
  PublicTestLog     string `json:"public_test_log" example:"Lorem Ipsum"`
  PrivateTestLog    string `json:"private_test_log" example:"Lorem Ipsum"`
  PublicTestStatus  int    `json:"public_test_status" example:"1"`
  PrivateTestStatus int    `json:"private_test_status" example:"0"`
  AcquiredPoints    int    `json:"acquired_points" example:"19"`
  Feedback          string `json:"feedback" example:"Some feedback"`
  TutorID           int64  `json:"tutor_id" example:"2"`
  SubmissionID      int64  `json:"submission_id" example:"31"`
}

// Render post-processes a GradeResponse.
func (body *GradeResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// newGradeResponse creates a response from a Grade model.
func newGradeResponse(p *model.Grade) *GradeResponse {
  return &GradeResponse{
    ID:                p.ID,
    ExecutionState:    p.ExecutionState,
    PublicTestLog:     p.PublicTestLog,
    PrivateTestLog:    p.PrivateTestLog,
    PublicTestStatus:  p.PublicTestStatus,
    PrivateTestStatus: p.PrivateTestStatus,
    AcquiredPoints:    p.AcquiredPoints,
    Feedback:          p.Feedback,
    TutorID:           p.TutorID,
    SubmissionID:      p.SubmissionID,
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
  Grade struct {
    ID                int64  `json:"id" example:"1"`
    ExecutionState    int    `json:"execution_state" example:"1"`
    PublicTestLog     string `json:"public_test_log" example:"Lorem Ipsum"`
    PrivateTestLog    string `json:"private_test_log" example:"Lorem Ipsum"`
    PublicTestStatus  int    `json:"public_test_status" example:"1"`
    PrivateTestStatus int    `json:"private_test_status" example:"0"`
    AcquiredPoints    int    `json:"acquired_points" example:"19"`
    Feedback          string `json:"feedback" example:"Some feedback"`
    TutorID           int64  `json:"tutor_id" example:"2"`
    SubmissionID      int64  `json:"submission_id" example:"31"`
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
  r := &MissingGradeResponse{
    // Grade:    p.Grade,
    CourseID: p.CourseID,
    SheetID:  p.SheetID,
    TaskID:   p.TaskID,
  }

  r.Grade.ID = p.Grade.ID
  r.Grade.ExecutionState = p.Grade.ExecutionState
  r.Grade.PublicTestLog = p.Grade.PublicTestLog
  r.Grade.PrivateTestLog = p.Grade.PrivateTestLog
  r.Grade.PublicTestStatus = p.Grade.PublicTestStatus
  r.Grade.PrivateTestStatus = p.Grade.PrivateTestStatus
  r.Grade.AcquiredPoints = p.Grade.AcquiredPoints
  r.Grade.Feedback = p.Grade.Feedback
  r.Grade.TutorID = p.Grade.TutorID
  r.Grade.SubmissionID = p.Grade.SubmissionID
  return r

}

// newMissingGradeListResponse creates a response from a list of Grade models.
func newMissingGradeListResponse(Grades []model.MissingGrade) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range Grades {
    list = append(list, newMissingGradeResponse(&Grades[k]))
  }
  return list
}
