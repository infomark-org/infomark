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
  "fmt"
  "net/http"
  "strconv"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/auth/authorize"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
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

// .............................................................................

// SubmissionResponse is the response payload for Submission management.
type SubmissionResponse struct {
  *model.Submission
  FileURL string `json="file_url`
}

// newSubmissionResponse creates a response from a Submission model.
func (rs *SubmissionResource) newSubmissionResponse(p *model.Submission) *SubmissionResponse {
  return &SubmissionResponse{
    Submission: p,
    FileURL:    fmt.Sprintf("/api/v1/submissions/%s/file", strconv.FormatInt(p.ID, 10)),
  }
}

// newSubmissionListResponse creates a response from a list of Submission models.
func (rs *SubmissionResource) newSubmissionListResponse(Submissions []model.Submission) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range Submissions {
    list = append(list, rs.newSubmissionResponse(&Submissions[k]))
  }
  return list
}

// Render post-processes a SubmissionResponse.
func (body *SubmissionResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// // IndexHandler is the enpoint for retrieving all Submissions if claim.root is true.
// func (rs *SubmissionResource) GetHandler(w http.ResponseWriter, r *http.Request) {

//   task := r.Context().Value("task").(*model.Task)
//   accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

//   submission, err := rs.Stores.Submission.GetByUserAndTask(accessClaims.LoginID, task.ID)
//   if err != nil {
//     render.Render(w, r, ErrNotFound)
//     return
//   }

//   // render JSON reponse
//   if err = render.Render(w, r, rs.newSubmissionResponse(submission)); err != nil {
//     render.Render(w, r, ErrRender(err))
//     return
//   }
// }

func (rs *SubmissionResource) GetFileHandler(w http.ResponseWriter, r *http.Request) {
  task := r.Context().Value("task").(*model.Task)
  // submission := r.Context().Value("submission").(*model.Submission)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  givenRole := r.Context().Value("course_role").(authorize.CourseRole)

  submission, err := rs.Stores.Submission.GetByUserAndTask(accessClaims.LoginID, task.ID)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
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

func (rs *SubmissionResource) GetFileByIdHandler(w http.ResponseWriter, r *http.Request) {

  submission := r.Context().Value("submission").(*model.Submission)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  givenRole := r.Context().Value("course_role").(authorize.CourseRole)

  submission, err := rs.Stores.Submission.Get(submission.ID)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
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

func (rs *SubmissionResource) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
  // will always be a POST
  task := r.Context().Value("task").(*model.Task)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  // todo create submission if not exists

  submission, err := rs.Stores.Submission.GetByUserAndTask(accessClaims.LoginID, task.ID)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // the file will be located
  if err := helper.NewSubmissionFileHandle(submission.ID).WriteToDisk(r, "file_data"); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
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
    if submissionID, err = strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64); err != nil {
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
