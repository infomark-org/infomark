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
	"github.com/cgtuebingen/infomark-backend/auth"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/cgtuebingen/infomark-backend/symbol"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// UserResource specifies user management handler.
type UserResource struct {
	Stores *Stores
}

// NewUserResource create and returns a UserResource.
func NewUserResource(stores *Stores) *UserResource {
	return &UserResource{
		Stores: stores,
	}
}

// .............................................................................

// IndexHandler is public endpoint for
// URL: /users
// METHOD: get
// TAG: users
// RESPONSE: 200,userResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  Get own user details (requires root)
func (rs *UserResource) IndexHandler(w http.ResponseWriter, r *http.Request) {

	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)

	if !accessClaims.Root {
		render.Render(w, r, ErrUnauthorized)
		return
	}

	// fetch collection of users from database
	users, err := rs.Stores.User.GetAll()
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// render JSON reponse
	if err = render.RenderList(w, r, newUserListResponse(users)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// GetMeHandler is public endpoint for
// URL: /me
// METHOD: get
// TAG: users
// RESPONSE: 200,userResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  Get own user details
func (rs *UserResource) GetMeHandler(w http.ResponseWriter, r *http.Request) {
	// `user` is retrieved via middle-ware
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	user, err := rs.Stores.User.Get(accessClaims.LoginID)

	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// render JSON reponse
	if err := render.Render(w, r, newUserResponse(user)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// Find is public endpoint for
// URL: /users/find
// QUERYPARAM: query,string
// METHOD: get
// TAG: users
// RESPONSE: 200,userResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  Query a specific user
func (rs *UserResource) Find(w http.ResponseWriter, r *http.Request) {
	query := helper.StringFromURL(r, "query", "%%")
	if query != "%%" {
		query = fmt.Sprintf("%%%s%%", query)
	}
	users, err := rs.Stores.User.Find(query)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	// render JSON reponse
	if err = render.RenderList(w, r, newUserListResponse(users)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// GetHandler is public endpoint for
// URL: /users/{user_id}
// URLPARAM: user_id,integer
// METHOD: get
// TAG: users
// RESPONSE: 200,userResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  Get user details
func (rs *UserResource) GetHandler(w http.ResponseWriter, r *http.Request) {
	// `user` is retrieved via middle-ware
	user := r.Context().Value(symbol.CtxKeyUser).(*model.User)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)

	// is request identity allowed to get informaition about this user
	if user.ID != accessClaims.LoginID {
		if !accessClaims.Root {
			render.Render(w, r, ErrUnauthorized)
			return
		}
	}

	// render JSON reponse
	if err := render.Render(w, r, newUserResponse(user)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// GetAvatarHandler is public endpoint for
// URL: /users/{user_id}/avatar
// URLPARAM: user_id,integer
// METHOD: get
// TAG: users
// RESPONSE: 200,userResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  Get user details
func (rs *UserResource) GetAvatarHandler(w http.ResponseWriter, r *http.Request) {
	// `user` is retrieved via middle-ware
	user := r.Context().Value(symbol.CtxKeyUser).(*model.User)

	file := helper.NewAvatarFileHandle(user.ID)

	if !file.Exists() {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := file.WriteToBody(w); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
	}
}

// EditMeHandler is public endpoint for
// URL: /me
// METHOD: put
// TAG: users
// REQUEST: userMeRequest
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  updating a the user record of the request identity
func (rs *UserResource) EditMeHandler(w http.ResponseWriter, r *http.Request) {

	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)

	data := &userMeRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	// user is not allowed to change all entries, we use the database entry as a starting point
	user, err := rs.Stores.User.Get(accessClaims.LoginID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	user.FirstName = data.FirstName
	user.LastName = data.LastName
	// no email update here
	user.StudentNumber = data.StudentNumber
	user.Semester = data.Semester
	user.Subject = data.Subject
	user.Language = data.Language

	// update database entry
	if err := rs.Stores.User.Update(user); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// EditHandler is public endpoint for
// URL: /users/{user_id}
// URLPARAM: user_id,integer
// METHOD: put
// TAG: users
// REQUEST: userRequest
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  updating a specific user with given id.
func (rs *UserResource) EditHandler(w http.ResponseWriter, r *http.Request) {

	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)

	if !accessClaims.Root {
		render.Render(w, r, ErrUnauthorized)
		return
	}

	data := &userRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	user := r.Context().Value(symbol.CtxKeyUser).(*model.User)

	user.FirstName = data.FirstName
	user.LastName = data.LastName
	user.Email = data.Email
	user.StudentNumber = data.StudentNumber
	user.Semester = data.Semester
	user.Subject = data.Subject
	user.Language = data.Language

	// all identities allowed to this endpoint are allowed to change the password
	if data.PlainPassword != "" {
		var err error
		user.EncryptedPassword, err = auth.HashPassword(data.PlainPassword)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}
	}

	// update database entry
	if err := rs.Stores.User.Update(user); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// SendEmailHandler is public endpoint for
// URL: /users/{user_id}/emails
// URLPARAM: user_id,integer
// METHOD: post
// TAG: users
// TAG: email
// REQUEST: EmailRequest
// RESPONSE: 200,OK
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  send email to a specific user
func (rs *UserResource) SendEmailHandler(w http.ResponseWriter, r *http.Request) {

	user := r.Context().Value(symbol.CtxKeyUser).(*model.User)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	accessUser, _ := rs.Stores.User.Get(accessClaims.LoginID)

	data := &EmailRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	// add sender identity
	msg := email.NewEmailFromUser(
		user.Email,
		data.Subject,
		data.Body,
		accessUser,
	)

	email.OutgoingEmailsChannel <- msg

}

// DeleteHandler is public endpoint for
// URL: /users/{user_id}
// URLPARAM: user_id,integer
// METHOD: delete
// TAG: users
// REQUEST: userRequest
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  updating a specific user with given id.
func (rs *UserResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(symbol.CtxKeyUser).(*model.User)

	// update database entry
	if err := rs.Stores.User.Delete(user.ID); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// .............................................................................

// Context middleware is used to load an User object from
// the URL parameter `userID` passed through as the request. In case
// the User could not be found, we stop here and return a 404.
// We do NOT check whether the user is authorized to get this user.
func (rs *UserResource) Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: check permission if inquirer of request is allowed to access this user
		// Should be done via another middleware
		var userID int64
		var err error

		// try to get id from URL
		if userID, err = strconv.ParseInt(chi.URLParam(r, "user_id"), 10, 64); err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// find specific user in database
		user, err := rs.Stores.User.Get(userID)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// serve next
		ctx := context.WithValue(r.Context(), symbol.CtxKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
