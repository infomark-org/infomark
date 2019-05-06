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
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/cgtuebingen/infomark-backend/symbol"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// GroupResource specifies Group management handler.
type GroupResource struct {
	Stores *Stores
}

// NewGroupResource create and returns a GroupResource.
func NewGroupResource(stores *Stores) *GroupResource {
	return &GroupResource{
		Stores: stores,
	}
}

// .............................................................................

// IndexHandler is public endpoint for
// URL: /courses/{course_id}/groups
// URLPARAM: course_id,integer
// METHOD: get
// TAG: groups
// RESPONSE: 200,GroupResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get all groups in course
// DESCRIPTION:
// The ordering is abitary
func (rs *GroupResource) IndexHandler(w http.ResponseWriter, r *http.Request) {

	var groups []model.GroupWithTutor
	var err error

	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	groups, err = rs.Stores.Group.GroupsOfCourse(course.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// render JSON reponse
	if err = render.RenderList(w, r, rs.newGroupListResponse(groups)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// CreateHandler is public endpoint for
// URL: /courses/{course_id}/groups
// URLPARAM: course_id,integer
// METHOD: post
// TAG: groups
// REQUEST: groupRequest
// RESPONSE: 204,SheetResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  create a new group
func (rs *GroupResource) CreateHandler(w http.ResponseWriter, r *http.Request) {

	// start from empty Request
	data := &groupRequest{}
	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

	group := &model.Group{}
	group.TutorID = data.Tutor.ID
	group.CourseID = course.ID
	group.Description = data.Description

	tutor, err := rs.Stores.User.Get(group.TutorID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// create Group entry in database
	newGroup, err := rs.Stores.Group.Create(group)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusCreated)

	// return Group information of created entry
	if err := render.Render(w, r, rs.newGroupResponse(newGroup, tutor)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

}

// GetHandler is public endpoint for
// URL: /courses/{course_id}/groups/{group_id}
// URLPARAM: course_id,integer
// URLPARAM: group_id,integer
// METHOD: get
// TAG: groups
// RESPONSE: 200,GroupResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get a specific group
func (rs *GroupResource) GetHandler(w http.ResponseWriter, r *http.Request) {
	// `Task` is retrieved via middle-ware
	group := r.Context().Value(symbol.CtxKeyGroup).(*model.Group)

	tutor, err := rs.Stores.User.Get(group.TutorID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// render JSON reponse
	if err := render.Render(w, r, rs.newGroupResponse(group, tutor)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusOK)
}

// GetMineHandler is public endpoint for
// URL: /courses/{course_id}/groups/own
// URLPARAM: course_id,integer
// METHOD: get
// TAG: groups
// RESPONSE: 200,GroupResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get the group the request identity is enrolled in
func (rs *GroupResource) GetMineHandler(w http.ResponseWriter, r *http.Request) {

	// TODO(patwie): handle case when user is tutor in group

	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	courseRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)

	var (
		groups []model.GroupWithTutor
		err    error
	)

	if courseRole == authorize.STUDENT {
		// here catch on the cases, when user is a student and enrolled in a group
		groups, err = rs.Stores.Group.GetInCourseWithUser(accessClaims.LoginID, course.ID)
	} else {
		// must be tutor
		groups, err = rs.Stores.Group.GetOfTutor(accessClaims.LoginID, course.ID)

	}

	// if we cannot find such an entry, this means the user have not been assigned to a group
	if err != nil {
		fmt.Println(err)
		render.Render(w, r, ErrNotFound)
		return
	}

	// // students can be only within one group -> [0] is ok
	// // tutors can be in multiple groups -> [0] is ok (they are the same person)
	// tutor, err := rs.Stores.User.Get(groups[0].TutorID)
	// if err != nil {
	//   render.Render(w, r, ErrInternalServerErrorWithDetails(err))
	//   return
	// }

	if len(groups) == 0 {
		render.Render(w, r, ErrNotFound)
		return
	}

	// render JSON reponse
	if err := render.RenderList(w, r, rs.newGroupListResponse(groups)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusOK)

}

// EditHandler is public endpoint for
// URL: /courses/{course_id}/groups/{group_id}
// URLPARAM: course_id,integer
// URLPARAM: group_id,integer
// METHOD: put
// TAG: groups
// REQUEST: groupRequest
// RESPONSE: 204,NotContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  update a specific group
func (rs *GroupResource) EditHandler(w http.ResponseWriter, r *http.Request) {
	// start from empty Request
	data := &groupRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	group := r.Context().Value(symbol.CtxKeyGroup).(*model.Group)
	group.TutorID = data.Tutor.ID
	group.Description = data.Description

	// update database entry
	if err := rs.Stores.Group.Update(group); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// DeleteHandler is public endpoint for
// URL: /courses/{course_id}/groups/{group_id}
// URLPARAM: course_id,integer
// URLPARAM: group_id,integer
// METHOD: delete
// TAG: groups
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  delete a specific group
func (rs *GroupResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	group := r.Context().Value(symbol.CtxKeyGroup).(*model.Group)

	// update database entry
	if err := rs.Stores.Group.Delete(group.ID); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// IndexEnrollmentsHandler is public endpoint for
// URL: /courses/{course_id}/groups/{group_id}/enrollments
// URLPARAM: course_id,integer
// URLPARAM: group_id,integer
// QUERYPARAM: roles,string
// QUERYPARAM: first_name,string
// QUERYPARAM: last_name,string
// QUERYPARAM: email,string
// QUERYPARAM: subject,string
// QUERYPARAM: language,string
// METHOD: get
// TAG: enrollments
// RESPONSE: 200,enrollmentResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  list all courses
func (rs *GroupResource) IndexEnrollmentsHandler(w http.ResponseWriter, r *http.Request) {
	// /courses/1/enrollments?roles=0,1
	group := r.Context().Value(symbol.CtxKeyGroup).(*model.Group)
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

	// extract filters
	filterRoles := helper.StringArrayFromURL(r, "roles", []string{"0", "1", "2"})
	filterFirstName := helper.StringFromURL(r, "first_name", "%%")
	filterLastName := helper.StringFromURL(r, "last_name", "%%")
	filterEmail := helper.StringFromURL(r, "email", "%%")
	filterSubject := helper.StringFromURL(r, "subject", "%%")
	filterLanguage := helper.StringFromURL(r, "language", "%%")

	givenRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)

	if givenRole == authorize.STUDENT {
		// students cannot query other students
		filterRoles = []string{"1", "2"}
	}

	enrolledUsers, err := rs.Stores.Group.EnrolledUsers(course.ID, group.ID,
		filterRoles, filterFirstName, filterLastName, filterEmail,
		filterSubject, filterLanguage,
	)
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

// EditGroupEnrollmentHandler is public endpoint for
// URL: /courses/{course_id}/groups/{group_id}/enrollments
// URLPARAM: course_id,integer
// URLPARAM: group_id,integer
// METHOD: post
// TAG: groups
// REQUEST: groupEnrollmentRequest
// RESPONSE: 204,NotContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  will assign a given user to a group or change the group assignment
func (rs *GroupResource) EditGroupEnrollmentHandler(w http.ResponseWriter, r *http.Request) {
	// start from empty Request
	data := &groupEnrollmentRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	group := r.Context().Value(symbol.CtxKeyGroup).(*model.Group)
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

	enrollment, err := rs.Stores.Group.GetGroupEnrollmentOfUserInCourse(data.UserID, course.ID)

	if err != nil {
		// does not exists yet

		enrollment := &model.GroupEnrollment{
			UserID:  data.UserID,
			GroupID: group.ID,
		}

		_, err := rs.Stores.Group.CreateGroupEnrollmentOfUserInCourse(enrollment)
		if err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}

	} else {
		// does exists --> simply change it
		enrollment.GroupID = group.ID
		if err := rs.Stores.Group.ChangeGroupEnrollmentOfUserInCourse(enrollment); err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}
	}

	render.Status(r, http.StatusNoContent)
}

// ChangeBidHandler is public endpoint for
// URL: /courses/{course_id}/groups/{group_id}/bids
// URLPARAM: course_id,integer
// URLPARAM: group_id,integer
// METHOD: post
// TAG: groups
// REQUEST: groupBidRequest
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  change or add the bid for enrolling in a group
func (rs *GroupResource) ChangeBidHandler(w http.ResponseWriter, r *http.Request) {

	courseRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)

	if courseRole != authorize.STUDENT {
		render.Render(w, r, ErrBadRequestWithDetails(errors.New("Only students in a course can bid for a group")))
		return
	}

	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)

	// start from empty Request
	group := r.Context().Value(symbol.CtxKeyGroup).(*model.Group)

	data := &groupBidRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	_, existsErr := rs.Stores.Group.GetBidOfUserForGroup(accessClaims.LoginID, group.ID)
	if existsErr == nil {
		// exists
		// update database entry
		if _, err := rs.Stores.Group.UpdateBidOfUserForGroup(accessClaims.LoginID, group.ID, data.Bid); err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}
		render.Status(r, http.StatusNoContent)
	} else {
		// insert
		// insert database entry
		if _, err := rs.Stores.Group.InsertBidOfUserForGroup(accessClaims.LoginID, group.ID, data.Bid); err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}
		render.Status(r, http.StatusCreated)

		resp := &GroupBidResponse{Bid: data.Bid}

		if err := render.Render(w, r, resp); err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}
	}

	render.Status(r, http.StatusNoContent)
}

// SendEmailHandler is public endpoint for
// URL: /courses/{course_id}/groups/{group_id}/emails
// URLPARAM: course_id,integer
// URLPARAM: group_id,integer
// METHOD: post
// TAG: groups
// TAG: email
// REQUEST: EmailRequest
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  send email to entire group
func (rs *GroupResource) SendEmailHandler(w http.ResponseWriter, r *http.Request) {

	group := r.Context().Value(symbol.CtxKeyGroup).(*model.Group)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	accessUser, _ := rs.Stores.User.Get(accessClaims.LoginID)

	data := &EmailRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	recipients, err := rs.Stores.Group.GetMembers(group.ID)

	if err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	msgOwn := email.NewEmailFromUser(
		accessUser.Email,
		data.Subject,
		data.Body,
		accessUser,
	)
	email.OutgoingEmailsChannel <- msg

	for _, recipient := range recipients {
		msg := email.NewEmailFromUser(
			recipient.Email,
			data.Subject,
			data.Body,
			accessUser,
		)

		email.OutgoingEmailsChannel <- msg
	}

}

// .............................................................................

// Context middleware is used to load an group object from
// the URL parameter `TaskID` passed through as the request. In case
// the group could not be found, we stop here and return a 404.
// We do NOT check whether the identity is authorized to get this group.
func (rs *GroupResource) Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		courseFromURL := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

		var groupID int64
		var err error

		// try to get id from URL
		if groupID, err = strconv.ParseInt(chi.URLParam(r, "group_id"), 10, 64); err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// find specific group in database
		group, err := rs.Stores.Group.Get(groupID)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), symbol.CtxKeyGroup, group)

		// when there is a groupID in the url, there is NOT a courseID in the url,
		// BUT: when there is a group, there is a course

		course, err := rs.Stores.Group.IdentifyCourseOfGroup(group.ID)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

		if courseFromURL.ID != course.ID {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx = context.WithValue(ctx, symbol.CtxKeyCourse, course)

		// serve next
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
