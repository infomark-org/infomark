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
  "github.com/cgtuebingen/infomark-backend/auth/authorize"
  "github.com/cgtuebingen/infomark-backend/common"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
)

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

// IndexHandler is public endpoint for
// URL: /courses
// METHOD: get
// TAG: courses
// RESPONSE: 200,courseResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  list all courses
func (rs *CourseResource) IndexHandler(w http.ResponseWriter, r *http.Request) {
  // fetch collection of courses from database
  courses, err := rs.Stores.Course.GetAll()

  // render JSON reponse
  if err = render.RenderList(w, r, rs.newCourseListResponse(courses)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// CreateHandler is public endpoint for
// URL: /courses
// METHOD: post
// TAG: courses
// REQUEST: courseRequest
// RESPONSE: 204,courseResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  create a new course
func (rs *CourseResource) CreateHandler(w http.ResponseWriter, r *http.Request) {
  // start from empty Request
  data := &courseRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  course := &model.Course{}
  course.Name = data.Name
  course.Description = data.Description
  course.BeginsAt = data.BeginsAt
  course.EndsAt = data.EndsAt
  course.RequiredPercentage = data.RequiredPercentage

  // create course entry in database
  newCourse, err := rs.Stores.Course.Create(course)
  if err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusCreated)

  // return course information of created entry
  if err := render.Render(w, r, rs.newCourseResponse(newCourse)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

}

// GetHandler is public endpoint for
// URL: /courses/{course_id}
// URLPARAM: course_id,integer
// METHOD: get
// TAG: courses
// RESPONSE: 200,courseResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get a specific course
func (rs *CourseResource) GetHandler(w http.ResponseWriter, r *http.Request) {
  // `course` is retrieved via middle-ware
  course, ok := r.Context().Value(common.CtxKeyCourse).(*model.Course)
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

// EditHandler is public endpoint for
// URL: /courses/{course_id}
// URLPARAM: course_id,integer
// METHOD: put
// TAG: courses
// REQUEST: courseRequest
// RESPONSE: 204,NotContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  update a specific course
func (rs *CourseResource) EditHandler(w http.ResponseWriter, r *http.Request) {
  // start from empty Request
  data := &courseRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  course := r.Context().Value(common.CtxKeyCourse).(*model.Course)
  course.Name = data.Name
  course.Description = data.Description
  course.BeginsAt = data.BeginsAt
  course.EndsAt = data.EndsAt
  course.RequiredPercentage = data.RequiredPercentage

  // update database entry
  if err := rs.Stores.Course.Update(course); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // TODO(patwie): change StatusNoContent
  render.Status(r, http.StatusNoContent)
}

// DeleteHandler is public endpoint for
// URL: /courses/{course_id}
// URLPARAM: course_id,integer
// METHOD: delete
// TAG: courses
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  delete a specific course
func (rs *CourseResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value(common.CtxKeyCourse).(*model.Course)

  // Warning: There is more to do! Currently we just dis-enroll all students,
  // remove all sheets and delete the course it self FROM THE DATABASE.
  // This does not remove gradings and the sheets or touches any file!

  // update database entry
  if err := rs.Stores.Course.Delete(course.ID); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

// IndexEnrollmentsHandler is public endpoint for
// URL: /courses/{course_id}/enrollments
// URLPARAM: course_id,integer
// QUERYPARAM: roles,string
// QUERYPARAM: first_name,string
// QUERYPARAM: last_name,string
// QUERYPARAM: email,string
// QUERYPARAM: subject,string
// QUERYPARAM: language,string
// QUERYPARAM: q,string
// METHOD: get
// TAG: enrollments
// RESPONSE: 200,enrollmentResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  list all courses
// DESCRIPTION:
// If the query 'q' parameter is given this endpoints returns all users which matches the query
// by first_name, last_name or email. The 'q' does not need be wrapped by '%'. But all other query strings
// do need to be wrapped by '%' to indicated end and start of a string.
func (rs *CourseResource) IndexEnrollmentsHandler(w http.ResponseWriter, r *http.Request) {
  // /courses/1/enrollments?roles=0,1
  course := r.Context().Value(common.CtxKeyCourse).(*model.Course)

  filterQuery := helper.StringFromUrl(r, "q", "")

  // extract filters
  filterRoles := helper.StringArrayFromUrl(r, "roles", []string{"0", "1", "2"})
  filterFirstName := helper.StringFromUrl(r, "first_name", "%%")
  filterLastName := helper.StringFromUrl(r, "last_name", "%%")
  filterEmail := helper.StringFromUrl(r, "email", "%%")
  filterSubject := helper.StringFromUrl(r, "subject", "%%")
  filterLanguage := helper.StringFromUrl(r, "language", "%%")

  givenRole := r.Context().Value(common.CtxKeyCourseRole).(authorize.CourseRole)

  if givenRole == authorize.STUDENT {
    // students cannot query other students
    filterRoles = []string{"1", "2"}
  }

  var (
    enrolledUsers []model.UserCourse
    err           error
  )

  if filterQuery != "" {
    filterQuery = fmt.Sprintf("%%%s%%", filterQuery)
    enrolledUsers, err = rs.Stores.Course.FindEnrolledUsers(course.ID,
      filterRoles, filterQuery,
    )
  } else {
    enrolledUsers, err = rs.Stores.Course.EnrolledUsers(course.ID,
      filterRoles, filterFirstName, filterLastName, filterEmail,
      filterSubject, filterLanguage,
    )

  }

  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  enrolledUsers = EnsurePrivacyInEnrollments(enrolledUsers, givenRole)

  // render JSON reponse
  if err = render.RenderList(w, r, newEnrollmentListResponse(enrolledUsers)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// GetUserEnrollmentHandler is public endpoint for
// URL: /courses/{course_id}/enrollments/{user_id}
// URLPARAM: course_id,integer
// URLPARAM: user_id,integer
// METHOD: get
// TAG: enrollments
// RESPONSE: 200,enrollmentResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  give enrollment of a specific user in a specific course
func (rs *CourseResource) GetUserEnrollmentHandler(w http.ResponseWriter, r *http.Request) {
  // /courses/1/enrollments?roles=0,1
  course := r.Context().Value(common.CtxKeyCourse).(*model.Course)
  user := r.Context().Value(common.CtxKeyUser).(*model.User)

  // find role in the course

  userEnrollment, err := rs.Stores.Course.GetUserEnrollment(course.ID, user.ID)
  if err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  resp := newEnrollmentResponse(userEnrollment)

  // render JSON reponse
  if err = render.Render(w, r, resp); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// DeleteUserEnrollmentHandler is public endpoint for
// URL: /courses/{course_id}/enrollments/{user_id}
// URLPARAM: course_id,integer
// URLPARAM: user_id,integer
// METHOD: delete
// TAG: enrollments
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  give enrollment of a specific user in a specific course
func (rs *CourseResource) DeleteUserEnrollmentHandler(w http.ResponseWriter, r *http.Request) {
  // /courses/1/enrollments?roles=0,1
  course := r.Context().Value(common.CtxKeyCourse).(*model.Course)
  user := r.Context().Value(common.CtxKeyUser).(*model.User)

  // find role in the course

  userEnrollment, err := rs.Stores.Course.GetUserEnrollment(course.ID, user.ID)
  if err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  if int64(userEnrollment.Role) > int64(authorize.STUDENT) {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("Cannot disenroll tutors")))
    return
  }

  if err := rs.Stores.Course.Disenroll(course.ID, user.ID); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

// ChangeRole is public endpoint for
// URL: /courses/{course_id}/enrollments/{user_id}
// URLPARAM: course_id,integer
// URLPARAM: user_id,integer
// METHOD: put
// TAG: enrollments
// REQUEST: changeRoleInCourseRequest
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  change role of specific user
func (rs *CourseResource) ChangeRole(w http.ResponseWriter, r *http.Request) {
  // /courses/1/enrollments?roles=0,1

  course := r.Context().Value(common.CtxKeyCourse).(*model.Course)
  user := r.Context().Value(common.CtxKeyUser).(*model.User)

  data := &changeRoleInCourseRequest{}
  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // update database entry
  if err := rs.Stores.Course.UpdateRole(course.ID, user.ID, data.Role); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// EnrollHandler is public endpoint for
// URL: /courses/{course_id}/enrollments
// URLPARAM: course_id,integer
// METHOD: post
// TAG: enrollments
// REQUEST: Empty
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  enroll a user into a course
func (rs *CourseResource) EnrollHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value(common.CtxKeyCourse).(*model.Course)
  accessClaims := r.Context().Value(common.CtxKeyAccessClaims).(*authenticate.AccessClaims)

  role := int64(0)
  if accessClaims.Root {
    role = int64(2)
  }

  // update database entry
  if err := rs.Stores.Course.Enroll(course.ID, accessClaims.LoginID, role); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  userEnrollment, err := rs.Stores.Course.GetUserEnrollment(course.ID, accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  render.Status(r, http.StatusCreated)

  if err := render.Render(w, r, newEnrollmentResponse(userEnrollment)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

}

// DisenrollHandler is public endpoint for
// URL: /courses/{course_id}/enrollments
// URLPARAM: course_id,integer
// METHOD: delete
// TAG: enrollments
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  disenroll a user from a course
func (rs *CourseResource) DisenrollHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value(common.CtxKeyCourse).(*model.Course)
  accessClaims := r.Context().Value(common.CtxKeyAccessClaims).(*authenticate.AccessClaims)

  givenRole := r.Context().Value(common.CtxKeyCourseRole).(authorize.CourseRole)

  if givenRole == authorize.TUTOR {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("tutors cannot disenroll from a course")))
    return
  }

  // update database entry
  if err := rs.Stores.Course.Disenroll(course.ID, accessClaims.LoginID); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

// SendEmailHandler is public endpoint for
// URL: /courses/{course_id}/emails
// URLPARAM: course_id,integer
// QUERYPARAM: roles,string
// QUERYPARAM: first_name,string
// QUERYPARAM: last_name,string
// QUERYPARAM: email,string
// QUERYPARAM: subject,string
// QUERYPARAM: language,string
// METHOD: post
// TAG: courses
// TAG: email
// REQUEST: EmailRequest
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  send email to entire course filtered
func (rs *CourseResource) SendEmailHandler(w http.ResponseWriter, r *http.Request) {

  course := r.Context().Value(common.CtxKeyCourse).(*model.Course)

  accessClaims := r.Context().Value(common.CtxKeyAccessClaims).(*authenticate.AccessClaims)
  accessUser, _ := rs.Stores.User.Get(accessClaims.LoginID)

  data := &EmailRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // extract filters
  filterRoles := helper.StringArrayFromUrl(r, "roles", []string{"0", "1", "2"})
  filterFirstName := "%%"
  filterLastName := "%%"
  filterEmail := "%%"
  filterSubject := "%%"
  filterLanguage := "%%"

  recipients, err := rs.Stores.Course.EnrolledUsers(course.ID,
    filterRoles, filterFirstName, filterLastName, filterEmail,
    filterSubject, filterLanguage,
  )

  if err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  for _, recipient := range recipients {
    // add sender identity
    msg := email.NewEmailFromUser(
      recipient.Email,
      data.Subject,
      data.Body,
      accessUser,
    )

    email.OutgoingEmailsChannel <- msg
  }

}

// PointsHandler is public endpoint for
// URL: /courses/{course_id}/points
// URLPARAM: course_id,integer
// METHOD: get
// TAG: courses
// RESPONSE: 200,SheetPointsResponseList
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get all points for the request identity
func (rs *CourseResource) PointsHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value(common.CtxKeyCourse).(*model.Course)
  accessClaims := r.Context().Value(common.CtxKeyAccessClaims).(*authenticate.AccessClaims)

  sheetPoints, err := rs.Stores.Course.PointsForUser(accessClaims.LoginID, course.ID)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // resp := &SheetPointsResponse{SheetPoints: sheetPoints}

  if err := render.RenderList(w, r, newSheetPointsListResponse(sheetPoints)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// BidsHandler is public endpoint for
// URL: /courses/{course_id}/bids
// URLPARAM: course_id,integer
// METHOD: get
// TAG: courses
// RESPONSE: 200,groupBidsResponseList
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get all bids for the request identity in a course
func (rs *CourseResource) BidsHandler(w http.ResponseWriter, r *http.Request) {
  course := r.Context().Value(common.CtxKeyCourse).(*model.Course)
  accessClaims := r.Context().Value(common.CtxKeyAccessClaims).(*authenticate.AccessClaims)

  givenRole := r.Context().Value(common.CtxKeyCourseRole).(authorize.CourseRole)

  var bids []model.GroupBid
  var err error

  if givenRole == authorize.TUTOR {
    // tutors see nothing
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("tutors cannot have bids for a group in a course")))
    return

  }

  if givenRole == authorize.STUDENT {
    // students only see their own bids
    bids, err = rs.Stores.Group.GetBidsForCourseForUser(course.ID, accessClaims.LoginID)
  } else {
    // admins see all (to later setup the bid)
    bids, err = rs.Stores.Group.GetBidsForCourse(course.ID)
  }

  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  if err := render.RenderList(w, r, newGroupBidsListResponse(bids)); err != nil {
    render.Render(w, r, ErrRender(err))
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
    var courseID int64
    var err error

    // try to get id from URL
    if courseID, err = strconv.ParseInt(chi.URLParam(r, "course_id"), 10, 64); err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // find specific course in database
    course, err := rs.Stores.Course.Get(courseID)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // serve next
    ctx := context.WithValue(r.Context(), common.CtxKeyCourse, course)
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}

// RoleContext middleware extracts the role of an identity in a given course
func (rs *CourseResource) RoleContext(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    course := r.Context().Value(common.CtxKeyCourse).(*model.Course)
    accessClaims := r.Context().Value(common.CtxKeyAccessClaims).(*authenticate.AccessClaims)

    // find role in the course
    courseRole, err := rs.Stores.Course.RoleInCourse(accessClaims.LoginID, course.ID)
    if err != nil {
      render.Render(w, r, ErrBadRequestWithDetails(err))
      return
    }

    if accessClaims.Root {
      courseRole = authorize.ADMIN
    }

    // serve next
    ctx := context.WithValue(r.Context(), common.CtxKeyCourseRole, courseRole)
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
