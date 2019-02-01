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

  "github.com/cgtuebingen/infomark-backend/api/auth"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
)

type UserStore interface {
  Get(userID int64) (*model.User, error)
  Update(p *model.User) error
  GetAll() ([]model.User, error)
}

// UserResource implements user management handler.
type UserResource struct {
  UserStore UserStore
}

// NewUserResource creates and returns a user resource.
func NewUserResource(userStore UserStore) *UserResource {
  return &UserResource{
    UserStore: userStore,
  }
}

// .............................................................................

// UsersRequest is the request payload for User data model.
type userRequest struct {
  *model.User
  ProtectedID   int64  `json:"id"`
  PlainPassword string `json:"plain_password"`
}

// UsersResponse is the response payload for the User data model.
type userResponse struct {
  *model.User
}

func newUserResponse(p *model.User) *userResponse {
  return &userResponse{
    User: p,
  }
}

func newUserListResponse(users []model.User) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range users {
    list = append(list, newUserResponse(&users[k]))
  }

  return list
}

// Bind user request
func (d *userRequest) Bind(r *http.Request) error {
  // sending the id via request is invalid as the id should be submitted in the url
  d.ProtectedID = 0

  // encrypt password
  hash, err := auth.HashPassword(d.PlainPassword)
  d.EncryptedPassword = hash

  return err
}

// render user response
func (u *userResponse) Render(w http.ResponseWriter, r *http.Request) error {
  // nothing to hide
  return nil
}

// render user response

// get all users
func (rs *UserResource) Index(w http.ResponseWriter, r *http.Request) {
  users, err := rs.UserStore.GetAll()

  if err = render.RenderList(w, r, newUserListResponse(users)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// get a user by id
func (rs *UserResource) Get(w http.ResponseWriter, r *http.Request) {
  user := r.Context().Value("user").(*model.User)

  if err := render.Render(w, r, newUserResponse(user)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

}

// update the user with given id
func (rs *UserResource) Patch(w http.ResponseWriter, r *http.Request) {

  data := &userRequest{User: r.Context().Value("user").(*model.User)}

  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequest)
    return
  }

  if err := data.User.Validate(); err != nil {
    render.Render(w, r, ErrBadRequest)
    return
  }

  if err := rs.UserStore.Update(data.User); err != nil {
    render.Render(w, r, ErrInternalServerError)
    return
  }

  render.Status(r, http.StatusNoContent)
}

// .............................................................................
// UsersCtx middleware is used to load an User object from
// the URL parameters passed through as the request. In case
// the User could not be found, we stop here and return a 404.
func (d *UserResource) Context(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // TODO: check permission if inquirer of request is allowed to access this user
    var user_id int64
    var err error

    if user_id, err = strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64); err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    user, err := d.UserStore.Get(user_id)

    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    ctx := context.WithValue(r.Context(), "user", user)
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
