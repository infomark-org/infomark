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
  "context"
  "encoding/json"
  "fmt"
  "net/http"
  "strconv"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/api/shared"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/auth/authorize"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
  "github.com/spf13/viper"
)

// SubmissionResource specifies Submission management handler.
type SubmissionResource struct {
  Stores *Stores
}

// NewSubmissionResource create and returns a SubmissionResource.
func NewSubmissionResource(stores *Stores) *SubmissionResource {
  return &SubmissionResource{
    Stores: stores,
  }
}

// GetFileHandler is public endpoint for
// URL: /tasks/{task_id}/submission
// URLPARAM: task_id,integer
// METHOD: get
// TAG: submissions
// RESPONSE: 200,ZipFile
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get the zip file containing the submission of the request identity for a given task
func (rs *SubmissionResource) GetFileHandler(w http.ResponseWriter, r *http.Request) {
  task := r.Context().Value("task").(*model.Task)
  // submission := r.Context().Value("submission").(*model.Submission)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  givenRole := r.Context().Value("course_role").(authorize.CourseRole)

  submission, err := rs.Stores.Submission.GetByUserAndTask(accessClaims.LoginID, task.ID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  // students can only access their own files
  if submission.UserID != accessClaims.LoginID {
    if givenRole == authorize.STUDENT {
      render.Render(w, r, ErrUnauthorized)
      return
    }
  }

  hnd := helper.NewSubmissionFileHandle(submission.ID)

  if !hnd.Exists() {
    render.Render(w, r, ErrNotFound)
    return
  } else {
    if err := hnd.WriteToBody(w); err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }
  }
}

// GetFileByIdHandler is public endpoint for
// URL: /submissions/{submission_id}/file
// URLPARAM: submission_id,integer
// METHOD: get
// TAG: submissions
// RESPONSE: 200,ZipFile
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get the zip file of a specific submission
func (rs *SubmissionResource) GetFileByIdHandler(w http.ResponseWriter, r *http.Request) {

  submission := r.Context().Value("submission").(*model.Submission)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  givenRole := r.Context().Value("course_role").(authorize.CourseRole)

  submission, err := rs.Stores.Submission.Get(submission.ID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  // students can only access their own files
  if submission.UserID != accessClaims.LoginID {
    if givenRole == authorize.STUDENT {
      render.Render(w, r, ErrUnauthorized)
      return
    }
  }

  hnd := helper.NewSubmissionFileHandle(submission.ID)

  if !hnd.Exists() {
    render.Render(w, r, ErrNotFound)
    return
  } else {
    if err := hnd.WriteToBody(w); err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }
  }
}

// UploadFileHandler is public endpoint for
// URL: /tasks/{task_id}/submission
// URLPARAM: task_id,integer
// URLPARAM: sheet_id,integer
// METHOD: post
// TAG: submissions
// REQUEST: zipfile
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  changes the zip file of a submission belonging to the request identity
func (rs *SubmissionResource) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
  task := r.Context().Value("task").(*model.Task)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  // todo create submission if not exists

  var grade *model.Grade

  // create ssubmisison if not exists
  submission, err := rs.Stores.Submission.GetByUserAndTask(accessClaims.LoginID, task.ID)
  if err != nil {
    // no such submission
    submission, err = rs.Stores.Submission.Create(&model.Submission{UserID: accessClaims.LoginID, TaskID: task.ID})
    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }

    // create also empty grade, which will be filled in later
    grade = &model.Grade{
      PublicExecutionState:  0,
      PrivateExecutionState: 0,
      PublicTestLog:         "...",
      PrivateTestLog:        "...",
      PublicTestStatus:      0,
      PrivateTestStatus:     0,
      AcquiredPoints:        0,
      Feedback:              "...",
      TutorID:               1,
      SubmissionID:          submission.ID,
    }

    _, err = rs.Stores.Grade.Create(grade)
    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }
  } else {
    // submission exists, we only need to get the grade
    grade, err = rs.Stores.Grade.GetForSubmission(submission.ID)
    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }
  }

  // the file will be located
  if err := helper.NewSubmissionFileHandle(submission.ID).WriteToDisk(r, "file_data"); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // enqueue file into testing queue
  // By definition user with id 1 is the system itself with root access
  tokenManager, err := authenticate.NewTokenAuth()
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }
  accessToken, err := tokenManager.CreateAccessJWT(
    authenticate.NewAccessClaims(1, true))
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // enqueue public test
  request := &shared.SubmissionAMQPWorkerRequest{
    SubmissionID: submission.ID,
    AccessToken:  accessToken,
    FrameworkFileURL: fmt.Sprintf("%s/api/v1/tasks/%s/public_file",
      viper.GetString("url"),
      strconv.FormatInt(task.ID, 10)),
    SubmissionFileURL: fmt.Sprintf("%s/api/v1/submissions/%s/file",
      viper.GetString("url"),
      strconv.FormatInt(submission.ID, 10)),
    ResultEndpointURL: fmt.Sprintf("%s/api/v1/grades/%s/public_result",
      viper.GetString("url"),
      strconv.FormatInt(grade.ID, 10)),
    DockerImage: task.PublicDockerImage,
  }

  body, err := json.Marshal(request)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  DefaultSubmissionProducer.Publish(body)

  // enqueue private test
  request = &shared.SubmissionAMQPWorkerRequest{
    SubmissionID: submission.ID,
    AccessToken:  accessToken,
    FrameworkFileURL: fmt.Sprintf("%s/api/v1/tasks/%s/private_file",
      viper.GetString("url"),
      strconv.FormatInt(task.ID, 10)),
    SubmissionFileURL: fmt.Sprintf("%s/api/v1/submissions/%s/file",
      viper.GetString("url"),
      strconv.FormatInt(submission.ID, 10)),
    ResultEndpointURL: fmt.Sprintf("%s/api/v1/grades/%s/private_result",
      viper.GetString("url"),
      strconv.FormatInt(grade.ID, 10)),
    DockerImage: task.PrivateDockerImage,
  }

  body, err = json.Marshal(request)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  DefaultSubmissionProducer.Publish(body)

  render.Status(r, http.StatusOK)
}

// IndexHandler is public endpoint for
// URL: /courses/{course_id}/submissions
// URLPARAM: course_id,integer
// QUERYPARAM: sheet_id,integer
// QUERYPARAM: task_id,integer
// QUERYPARAM: group_id,integer
// QUERYPARAM: user_id,integer
// METHOD: get
// TAG: submissions
// RESPONSE: 200,SubmissionResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  Query submissions in a course
func (rs *SubmissionResource) IndexHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value("course").(*model.Course)

  filterGroupID := helper.Int64FromUrl(r, "group_id", 0)
  filterUserID := helper.Int64FromUrl(r, "user_id", 0)
  filterSheetID := helper.Int64FromUrl(r, "sheet_id", 0)
  filterTaskID := helper.Int64FromUrl(r, "task_id", 0)

  submissions, err := rs.Stores.Submission.GetFiltered(course.ID, filterGroupID, filterUserID, filterSheetID, filterTaskID)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // render JSON reponse
  if err = render.RenderList(w, r, newSubmissionListResponse(submissions)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)

}

// .............................................................................
// Context middleware is used to load an Submission object from
// the URL parameter `TaskID` passed through as the request. In case
// the Submission could not be found, we stop here and return a 404.
// We do NOT check whether the identity is authorized to get this Submission.
func (rs *SubmissionResource) Context(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // TODO: check permission if inquirer of request is allowed to access this Submission
    // Should be done via another middleware
    var submissionID int64
    var err error

    // try to get id from URL
    if submissionID, err = strconv.ParseInt(chi.URLParam(r, "submission_id"), 10, 64); err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // find specific Submission in database
    submission, err := rs.Stores.Submission.Get(submissionID)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    ctx := context.WithValue(r.Context(), "submission", submission)

    // when there is a submissionID in the url, there is NOT a taskID in the url,
    // BUT: when there is a Submission, there is a task
    // BUT: when there is a task, there is a course (other middlewarwe)

    // find specific Task in database
    var task *model.Task
    task, err = rs.Stores.Task.Get(submission.TaskID)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    ctx = context.WithValue(ctx, "task", task)

    // find course

    course, err := rs.Stores.Task.IdentifyCourseOfTask(task.ID)
    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }

    ctx = context.WithValue(ctx, "course", course)

    // serve next
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
