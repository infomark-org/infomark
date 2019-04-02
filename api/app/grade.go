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
  "errors"
  "fmt"
  "net/http"
  "strconv"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
)

// GradeResource specifies Grade management handler.
type GradeResource struct {
  Stores *Stores
}

// NewGradeResource create and returns a GradeResource.
func NewGradeResource(stores *Stores) *GradeResource {
  return &GradeResource{
    Stores: stores,
  }
}

// EditHandler is public endpoint for
// URL: /courses/{course_id}/grades/{grade_id}
// URLPARAM: course_id,integer
// URLPARAM: grade_id,integer
// METHOD: put
// TAG: grades
// REQUEST: GradeRequest
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  edit a grade
func (rs *GradeResource) EditHandler(w http.ResponseWriter, r *http.Request) {
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  currentGrade := r.Context().Value("grade").(*model.Grade)
  data := &GradeRequest{}
  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  task, err := rs.Stores.Grade.IdentifyTaskOfGrade(currentGrade.ID)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  if data.AcquiredPoints > task.MaxPoints {
    render.Render(w, r, ErrBadRequestWithDetails(fmt.Errorf("aquired points is larger than max-points %v is more than %v", data.AcquiredPoints, task.MaxPoints)))
    return
  }

  currentGrade.Feedback = data.Feedback
  currentGrade.AcquiredPoints = data.AcquiredPoints

  currentGrade.TutorID = accessClaims.LoginID

  // update database entry
  if err := rs.Stores.Grade.Update(currentGrade); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

// GetByIDHandler is public endpoint for
// URL: /courses/{course_id}/grades/{grade_id}
// URLPARAM: course_id,integer
// URLPARAM: grade_id,integer
// METHOD: get
// TAG: grades
// RESPONSE: 200,GradeResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get a grade
func (rs *GradeResource) GetByIDHandler(w http.ResponseWriter, r *http.Request) {

  currentGrade := r.Context().Value("grade").(*model.Grade)

  // return Material information of created entry
  if err := render.Render(w, r, newGradeResponse(currentGrade)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

// PublicResultEditHandler is public endpoint for
// URL: /courses/{course_id}/grades/{grade_id}/public_result
// URLPARAM: course_id,integer
// URLPARAM: grade_id,integer
// METHOD: post
// TAG: internal
// REQUEST: GradeFromWorkerRequest
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  update information for grade from background worker
func (rs *GradeResource) PublicResultEditHandler(w http.ResponseWriter, r *http.Request) {

  data := &GradeFromWorkerRequest{}
  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  currentGrade := r.Context().Value("grade").(*model.Grade)
  // currentGrade.PublicTestLog = data.Log
  // currentGrade.PublicTestStatus = data.Status
  // currentGrade.PublicExecutionState = 2

  render.Status(r, http.StatusNoContent)

  // update database entry
  if err := rs.Stores.Grade.UpdatePublicTestInfo(currentGrade.ID, data.Log, data.Status); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

}

// PrivateResultEditHandler is public endpoint for
// URL: /courses/{course_id}/grades/{grade_id}/private_result
// URLPARAM: course_id,integer
// URLPARAM: grade_id,integer
// METHOD: post
// TAG: internal
// REQUEST: GradeFromWorkerRequest
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  update information for grade from background worker
func (rs *GradeResource) PrivateResultEditHandler(w http.ResponseWriter, r *http.Request) {

  data := &GradeFromWorkerRequest{}
  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  currentGrade := r.Context().Value("grade").(*model.Grade)
  // currentGrade.PrivateTestLog = data.Log
  // currentGrade.PrivateTestStatus = data.Status
  // currentGrade.PrivateExecutionState = 2

  // fmt.Println(currentGrade.ID)
  // fmt.Println(currentGrade.PrivateTestLog)
  // fmt.Println(currentGrade.PrivateTestStatus)

  render.Status(r, http.StatusNoContent)

  // update database entry
  if err := rs.Stores.Grade.UpdatePrivateTestInfo(currentGrade.ID, data.Log, data.Status); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

}

// IndexHandler is public endpoint for
// URL: /courses/{course_id}/grades
// URLPARAM: course_id,integer
// QUERYPARAM: sheet_id,integer
// QUERYPARAM: task_id,integer
// QUERYPARAM: group_id,integer
// QUERYPARAM: user_id,integer
// QUERYPARAM: tutor_id,integer
// QUERYPARAM: feedback,string
// QUERYPARAM: acquired_points,integer
// QUERYPARAM: public_test_status,integer
// QUERYPARAM: private_test_status,integer
// QUERYPARAM: public_execution_state,integer
// QUERYPARAM: private_execution_state,integer
// METHOD: get
// TAG: grades
// RESPONSE: 200,GradeResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  Query grades in a course
func (rs *GradeResource) IndexHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value("course").(*model.Course)

  filterSheetID := helper.Int64FromUrl(r, "sheet_id", 0)
  filterTaskID := helper.Int64FromUrl(r, "task_id", 0)
  filterGroupID := helper.Int64FromUrl(r, "group_id", 0)

  if filterGroupID == 0 {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("group_id is missing")))
    return
  }

  filterUserID := helper.Int64FromUrl(r, "user_id", 0)
  filterTutorID := helper.Int64FromUrl(r, "tutor_id", 0)
  filterFeedback := helper.StringFromUrl(r, "feedback", "%%")
  filterAcquiredPoints := helper.IntFromUrl(r, "acquired_points", -1)
  filterPublicTestStatus := helper.IntFromUrl(r, "public_test_status", 0)
  filterPrivateTestStatus := helper.IntFromUrl(r, "private_test_status", 0)
  filterPublicExecutationState := helper.IntFromUrl(r, "public_execution_state", -1)
  filterPrivateExecutationState := helper.IntFromUrl(r, "private_execution_state", -1)

  submissions, err := rs.Stores.Grade.GetFiltered(
    course.ID,
    filterSheetID,
    filterTaskID,
    filterGroupID,
    filterUserID,
    filterTutorID,
    filterFeedback,
    filterAcquiredPoints,
    filterPublicTestStatus,
    filterPrivateTestStatus,
    filterPublicExecutationState,
    filterPrivateExecutationState,
  )
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // render JSON reponse
  if err = render.RenderList(w, r, newGradeListResponse(submissions)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)

}

// IndexMissingHandler is public endpoint for
// URL: /courses/{course_id}/grades/missing
// URLPARAM: course_id,integer
// QUERYPARAM: group_id,integer
// METHOD: get
// TAG: grades
// RESPONSE: 200,MissingGradeResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  the missing grades for the request identity
func (rs *GradeResource) IndexMissingHandler(w http.ResponseWriter, r *http.Request) {
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  course := r.Context().Value("course").(*model.Course)

  filterGroupID := helper.Int64FromUrl(r, "group_id", 0)

  grades, err := rs.Stores.Grade.GetAllMissingGrades(course.ID, accessClaims.LoginID, filterGroupID)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // render JSON reponse
  if err = render.RenderList(w, r, newMissingGradeListResponse(grades)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)

}

// .............................................................................
// Context middleware is used to load an Grade object from
// the URL parameter `TaskID` passed through as the request. In case
// the Grade could not be found, we stop here and return a 404.
// We do NOT check whether the identity is authorized to get this Grade.

func (rs *GradeResource) Context(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

    course_from_url := r.Context().Value("course").(*model.Course)

    var gradeID int64
    var err error

    // try to get id from URL
    if gradeID, err = strconv.ParseInt(chi.URLParam(r, "grade_id"), 10, 64); err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // find specific course in database
    grade, err := rs.Stores.Grade.Get(gradeID)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // serve next
    ctx := context.WithValue(r.Context(), "grade", grade)

    // when there is a gradeID in the url, there is NOT a courseID in the url,
    // BUT: when there is a grade, there is a course

    course, err := rs.Stores.Grade.IdentifyCourseOfGrade(grade.ID)
    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }

    if course_from_url.ID != course.ID {
      render.Render(w, r, ErrNotFound)
      return
    }

    ctx = context.WithValue(ctx, "course", course)

    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
