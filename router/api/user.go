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
		// TODO: check permission if inquirer of request is allowed to access this user

		user, err := store.DS().GetUserFromIdString(chi.URLParam(r, "userID"))

		if err != nil {
			render.Render(w, r, helper.ErrNotFoundResponse)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// .............................................................................

func UserRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", UsersIndex)  // curl -i -X GET http://localhost:3000/api/users
	r.Post("/", UserCreate) // curl -i -X POST -d '{"id":33,"last_name":"awesomeness"}' http://localhost:3000/api/users/
	r.Route("/{userID}", func(r chi.Router) {
		r.Use(UserCtx)
		r.Get("/", UserGet)       // curl -i -X GET http://localhost:3000/api/users/1
		r.Put("/", UserUpdate)    // curl -i -X PUT -d '{"first_name":"dude"}' http://localhost:3000/api/users/1
		r.Delete("/", UserDelete) // curl -i -X DELETE http://localhost:3000/api/users/1
	})

	return r
}

// .............................................................................

// UserRequest is the request payload for User data model.
type UserRequest struct {
	*model.User
	ProtectedId int `json:"id"`
}

// UserResponse is the response payload for the User data model.
type UserResponse struct {
	*model.User
}

type UserListResponse []*UserResponse

func NewUserResponse(u *model.User) *UserResponse {
	resp := &UserResponse{User: u}
	return resp
}

func NewUserListResponse(users []model.User) []render.Renderer {
	// https://stackoverflow.com/a/36463641/7443104
	list := []render.Renderer{}
	for k, _ := range users {
		list = append(list, NewUserResponse(&users[k]))
	}

	return list
}

// Bind user request
func (u *UserRequest) Bind(r *http.Request) error {
	// sending the id via request is invalid as the id should be submitted in the url
	u.ProtectedId = 0
	return nil
}

// render user response
func (u *UserResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// nothing to hide
	return nil
}

// .............................................................................

// UsersIndex returns all Users.
// GET "/users"
func UsersIndex(w http.ResponseWriter, r *http.Request) {

	users := &[]model.User{}

	if err := store.ORM().Find(users).Error; err != nil {
		render.Render(w, r, helper.ErrDatabaseResponse(err))
		return
	}

	if err := render.RenderList(w, r, NewUserListResponse(*users)); err != nil {
		render.Render(w, r, helper.ErrRenderResponse(err))
		return
	}

}

// UserGet returns the specific User.
// GET "/users/{userID}"
func UserGet(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*model.User)

	if err := render.Render(w, r, NewUserResponse(user)); err != nil {
		render.Render(w, r, helper.ErrRenderResponse(err))
		return
	}

}

// UserCreate persists the posted User and returns it
// back to the client as an acknowledgement.
// POST "/users"
func UserCreate(w http.ResponseWriter, r *http.Request) {
	data := &UserRequest{}

	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, helper.NewErrResponse(http.StatusBadRequest, err))
		return
	}

	if hints, err := data.User.Validate(); err != nil {
		helper.RenderValidation(hints, w, r)
		return
	}

	if err := store.ORM().Create(&data.User).Error; err != nil {
		render.Render(w, r, helper.ErrDatabaseResponse(err))
		return
	}
	render.Status(r, http.StatusCreated)

	if err := render.Render(w, r, NewUserResponse(data.User)); err != nil {
		render.Render(w, r, helper.ErrRenderResponse(err))
		return
	}
}

// UserUpdate changes the user information for a given user.
// PUT "/users/{userID}"
func UserUpdate(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*model.User)

	data := &UserRequest{User: user}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, helper.NewErrResponse(http.StatusBadRequest, err))
		return
	}

	if hints, err := data.User.Validate(); err != nil {
		helper.RenderValidation(hints, w, r)
		return
	}

	if err := store.ORM().Model(&user).Updates(data.User).Error; err != nil {
		render.Render(w, r, helper.ErrDatabaseResponse(err))
		return
	}

	if err := render.Render(w, r, NewUserResponse(user)); err != nil {
		render.Render(w, r, helper.ErrRenderResponse(err))
		return
	}
}

// UserDelete removes an user from a database.
// DELETE "/users/{userID}"
func UserDelete(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*model.User)

	if err := store.ORM().Delete(&user).Error; err != nil {
		render.Render(w, r, helper.ErrDatabaseResponse(err))
		return
	}

	if err := render.Render(w, r, NewUserResponse(user)); err != nil {
		render.Render(w, r, helper.ErrRenderResponse(err))
		return
	}
}
