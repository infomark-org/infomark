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
  "net/http"

  "github.com/cgtuebingen/infomark-backend/api/auth"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/render"
)

// AccountResource specifies user management handler.
type AccountResource struct {
  UserStore UserStore
}

// NewAccountResource create and returns a AccountResource.
func NewAccountResource(userStore UserStore) *AccountResource {
  return &AccountResource{
    UserStore: userStore,
  }
}

// .............................................................................

// userAccountRequest is the request payload for account management.
type userAccountRequest struct {
  User    *model.User `json:"user"`
  Account struct {
    Email         string `json:"email"`
    PlainPassword string `json:"plain_password"`
  } `json:"account"`
}

// UsersResponse is the response payload for the User data model.
type userAccountResponse struct {
  User *model.User
}

func newUserAccountResponse(p *model.User) *userAccountResponse {
  return &userAccountResponse{
    User: p,
  }
}

// Bind user request
func (d *userAccountRequest) Bind(r *http.Request) error {
  // sending the id via request is invalid as the id should be submitted in the url
  d.User.ID = 0

  // encrypt password
  hash, err := auth.HashPassword(d.Account.PlainPassword)
  d.User.EncryptedPassword = hash

  return err
}

// render user response
func (u *userAccountResponse) Render(w http.ResponseWriter, r *http.Request) error {
  // nothing to hide
  return nil
}

func (rs *AccountResource) bindValidate(w http.ResponseWriter, r *http.Request) (*userAccountRequest, error) {
  data := &userAccountRequest{}
  if err := render.Bind(r, data); err != nil {
    return nil, err
  }

  if err := data.User.Validate(); err != nil {
    return nil, err
  }

  return data, nil
}

func (rs *AccountResource) Post(w http.ResponseWriter, r *http.Request) {

  data, err := rs.bindValidate(w, r)
  if err != nil {
    render.Render(w, r, ErrBadRequest)
    return
  }

  newUser, err := rs.UserStore.Create(data.User)
  if err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  if err := render.Render(w, r, newUserAccountResponse(newUser)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// update the user with given id
func (rs *AccountResource) Patch(w http.ResponseWriter, r *http.Request) {

  data, err := rs.bindValidate(w, r)
  if err != nil {
    render.Render(w, r, ErrBadRequest)
    return
  }

  // TODO(patwie) verify userID
  if err = rs.UserStore.Update(data.User); err != nil {
    render.Render(w, r, ErrInternalServerError)
    return
  }

  if err := render.Render(w, r, newUserAccountResponse(data.User)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// get a user by id
func (rs *AccountResource) Get(w http.ResponseWriter, r *http.Request) {
  // TODO(patwie): from middleware and webtoken
  user, err := rs.UserStore.Get(1)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  if err := render.Render(w, r, newUserResponse(user)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

}
