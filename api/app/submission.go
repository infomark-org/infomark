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
  "net/http"
  "strconv"

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
}

// newSubmissionResponse creates a response from a Submission model.
func (rs *SubmissionResource) newSubmissionResponse(p *model.Submission) *SubmissionResponse {
  return &SubmissionResponse{
    Submission: p,
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

// IndexHandler is the enpoint for retrieving all Submissions if claim.root is true.
func (rs *SubmissionResource) IndexHandler(w http.ResponseWriter, r *http.Request) {

  // var Submissions []model.Submission
  // var err error

  // submission := r.Context().Value("submission").(*model.Course)
  // Submissions, err = rs.Stores.Submission.SubmissionsOfCourse(submission.ID)

  // // render JSON reponse
  // if err = render.RenderList(w, r, rs.newSubmissionListResponse(Submissions)); err != nil {
  //   render.Render(w, r, ErrRender(err))
  //   return
  // }
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
    Submission, err := rs.Stores.Submission.Get(submissionID)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    ctx := context.WithValue(r.Context(), "Submission", Submission)

    // when there is a submissionID in the url, there is NOT a courseID in the url,
    // BUT: when there is a Submission, there is a course

    // course, err := rs.Stores.Submission.IdentifyCourseOfSubmission(Submission.ID)
    // if err != nil {
    //   render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    //   return
    // }

    // ctx = context.WithValue(ctx, "course", course)

    // serve next
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
