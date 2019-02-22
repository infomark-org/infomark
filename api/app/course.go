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
  "net/http"
  "strconv"

  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
  validation "github.com/go-ozzo/ozzo-validation"
)

// CourseStore specifies required database queries for course management.
type CourseStore interface {
  Get(courseID int64) (*model.Course, error)
  Update(p *model.Course) error
  GetAll() ([]model.Course, error)
  Create(p *model.Course) (*model.Course, error)
  Delete(courseID int64) error
  Enroll(courseID int64, userID int64) error
  Disenroll(courseID int64, userID int64) error
  EnrolledUsers(course *model.Course) ([]model.UserCourseEnrollment, error)
}

// CourseResource specifies course management handler.
type CourseResource struct {
  Stores *Stores
}

// NewCourseResource create and returns a CourseResource.
func NewCourseResource(stores *Stores) *CourseResource {
  return &CourseResource{
    Stores: stores,
  }
}

// .............................................................................

// courseRequest is the request payload for course management.
type courseRequest struct {
  *model.Course
  ProtectedID int64 `json:"id"`
}

// courseResponse is the response payload for course management.
type courseResponse struct {
  *model.Course
}

// newCourseResponse creates a response from a course model.
func (rs *CourseResource) newCourseResponse(p *model.Course) *courseResponse {

  return &courseResponse{
    Course: p,
  }
}

// newCourseListResponse creates a response from a list of course models.
func (rs *CourseResource) newCourseListResponse(courses []model.Course) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range courses {
    list = append(list, rs.newCourseResponse(&courses[k]))
  }

  return list
}

// Bind preprocesses a courseRequest.
func (body *courseRequest) Bind(r *http.Request) error {
  // Sending the id via request-body is invalid.
  // The id should be submitted in the url.
  body.ProtectedID = 0

  err := validation.ValidateStruct(body,
    validation.Field(&body.Name, validation.Required),
  )
  return err

}

// Render post-processes a courseResponse.
func (body *courseResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// .............................................................................

// courseResponse is the response payload for course management.
type enrollmentResponse struct {
  Role int64       `json:"role"`
  User *model.User `json:"user"`
}

// Render post-processes a courseResponse.
func (body *enrollmentResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// newCourseResponse creates a response from a course model.
func (rs *CourseResource) newEnrollmentResponse(p *model.UserCourseEnrollment) *enrollmentResponse {

  return &enrollmentResponse{
    Role: p.Role,
    User: &model.User{
      ID:            p.UserID,
      FirstName:     p.FirstName,
      LastName:      p.LastName,
      AvatarPath:    p.AvatarPath,
      Email:         p.Email,
      StudentNumber: p.StudentNumber,
      Semester:      p.Semester,
      Subject:       p.Subject,
      Language:      p.Language,
    },
  }
}

func (rs *CourseResource) newEnrollmentListResponse(enrollments []model.UserCourseEnrollment) []render.Renderer {
  list := []render.Renderer{}
  for k := range enrollments {
    list = append(list, rs.newEnrollmentResponse(&enrollments[k]))
  }

  return list
}

// .............................................................................

// IndexHandler is the enpoint for retrieving all courses if claim.root is true.
func (rs *CourseResource) IndexHandler(w http.ResponseWriter, r *http.Request) {
  // fetch collection of courses from database
  courses, err := rs.Stores.Course.GetAll()

  // render JSON reponse
  if err = render.RenderList(w, r, rs.newCourseListResponse(courses)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// CreateHandler is the enpoint for retrieving all courses if claim.root is true.
func (rs *CourseResource) CreateHandler(w http.ResponseWriter, r *http.Request) {
  // start from empty Request
  data := &courseRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // validate final model
  if err := data.Course.Validate(); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // create course entry in database
  newCourse, err := rs.Stores.Course.Create(data.Course)
  if err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  // return course information of created entry
  if err := render.Render(w, r, rs.newCourseResponse(newCourse)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusCreated)
}

// GetHandler is the enpoint for retrieving a specific course.
func (rs *CourseResource) GetHandler(w http.ResponseWriter, r *http.Request) {
  // `course` is retrieved via middle-ware
  course, ok := r.Context().Value("course").(*model.Course)
  if !ok {
    render.Render(w, r, ErrInternalServerErrorWithDetails(errors.New("course context is missing")))
    return
  }

  // render JSON reponse
  if err := render.Render(w, r, rs.newCourseResponse(course)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// PatchHandler is the endpoint fro updating a specific course with given id.
func (rs *CourseResource) EditHandler(w http.ResponseWriter, r *http.Request) {
  // start from empty Request
  data := &courseRequest{
    Course: r.Context().Value("course").(*model.Course),
  }

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // update database entry
  if err := rs.Stores.Course.Update(data.Course); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

func (rs *CourseResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value("course").(*model.Course)

  // update database entry
  if err := rs.Stores.Course.Delete(course.ID); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusOK)
}

func (rs *CourseResource) IndexEnrollmentsHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value("course").(*model.Course)

  enrolledUsers, err := rs.Stores.Course.EnrolledUsers(course)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // render JSON reponse
  if err = render.RenderList(w, r, rs.newEnrollmentListResponse(enrolledUsers)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

func (rs *CourseResource) EnrollHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value("course").(*model.Course)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  // update database entry
  if err := rs.Stores.Course.Enroll(course.ID, accessClaims.LoginID); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  user, err := rs.Stores.User.Get(accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  resp := &enrollmentResponse{
    Role: 0,
    User: user,
  }

  if err := render.Render(w, r, resp); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

func (rs *CourseResource) DisenrollHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value("course").(*model.Course)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  // update database entry
  if err := rs.Stores.Course.Disenroll(course.ID, accessClaims.LoginID); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// .............................................................................
// Context middleware is used to load an Course object from
// the URL parameter `courseID` passed through as the request. In case
// the Course could not be found, we stop here and return a 404.
// We do NOT check whether the course is authorized to get this course.
func (rs *CourseResource) Context(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // TODO: check permission if inquirer of request is allowed to access this course
    // Should be done via another middleware
    var course_id int64
    var err error

    // try to get id from URL
    if course_id, err = strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64); err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // find specific course in database
    course, err := rs.Stores.Course.Get(course_id)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // serve next
    ctx := context.WithValue(r.Context(), "course", course)
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}