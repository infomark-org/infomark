// Copyright 2019 ComputerGraphics Tuebingen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ==============================================================================
// Authors: Patrick Wieschollek

package api

import (
	"context"
	"net/http"
	"time"

	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/cgtuebingen/infomark-backend/router/helper"
	"github.com/cgtuebingen/infomark-backend/store"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// .............................................................................

// UserCtx middleware is used to load an User object from
// the URL parameters passed through as the request. In case
// the User could not be found, we stop here and return a 404.
func UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// TODO: check permission if request is allowed to access this user

		user, err := store.DS().GetUserFromIdString(chi.URLParam(r, "userID"))

		if err != nil {
			render.Render(w, r, helper.ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// .............................................................................

func UserRoutes() chi.Router {
	r := chi.NewRouter()

	// curl -i -X GET http://localhost:3000/api/users
	r.Get("/", UsersIndex)
	// curl -X POST -d '{"id":33,"last_name":"awesomeness"}' http://localhost:3000/api/users/
	r.Post("/", UserCreate)
	r.Route("/{userID}", func(r chi.Router) {
		r.Use(UserCtx)
		// curl -i -X GET http://localhost:3000/api/users/1
		r.Get("/", UserGet)
		// curl -X PUT -d '{"first_name":"dude"}' http://localhost:3000/api/users/1
		r.Put("/", UserUpdate)
		// curl -i -X DELETE http://localhost:3000/api/users/1
		r.Delete("/", helper.EmptyHandler)
	})

	return r
}

// .............................................................................

// UserResponse is the response payload for the User data model.
type UserResponse struct {
	*model.User
}

// UserRequest is the request payload for User data model.
type UserRequest struct {
	*model.User
}

func NewUserResponse(u *model.User) *UserResponse {
	return &UserResponse{User: u}
}

func (u *UserResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// nothing to hide
	return nil
}

func (u *UserRequest) Bind(r *http.Request) error {
	// sending the id via json is invalid as the id should be submitted in the url
	u.ID = 0
	// seems to be GORM issue/bug
	u.CreatedAt = time.Time{}
	return nil
}

// .............................................................................

// UsersIndex returns all Users.
// GET "/users"
func UsersIndex(w http.ResponseWriter, r *http.Request) {
	users := &[]model.User{}
	store.ORM().Find(users)
	helper.WriteJSON(w, users)
}

// UserGet returns the specific User.
// GET "/users/{userID}"
func UserGet(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*model.User)

	if err := render.Render(w, r, NewUserResponse(user)); err != nil {
		render.Render(w, r, helper.ErrRender(err))
		return
	}

}

// UserCreate persists the posted User and returns it
// back to the client as an acknowledgement.
// POST "/users"
func UserCreate(w http.ResponseWriter, r *http.Request) {
	data := &UserRequest{}

	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, helper.ErrInvalidRequest(err))
		return
	}

	store.ORM().Create(&data.User)
	render.Status(r, http.StatusCreated)
	render.Render(w, r, NewUserResponse(data.User))
}

// UserUpdate changes the user information for a given user.
// PUT "/users/{userID}"
func UserUpdate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*model.User)

	data := &UserRequest{User: user}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, helper.ErrInvalidRequest(err))
		return
	}

	store.ORM().Model(&user).Updates(data.User)
}

// UserDelete removes an user from a database.
// DELETE "/users/{userID}"
func UserDelete(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*model.User)
	db.Delete(&user)
}
