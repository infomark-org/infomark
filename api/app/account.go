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
  "strings"

  "github.com/cgtuebingen/infomark-backend/auth"
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

type accountInfo struct {
  Email         string `json:"email"`
  PlainPassword string `json:"plain_password"`
}

// userAccountRequest is the request payload for account management.
type userAccountRequest struct {
  User    *model.User `json:"user"`
  Account accountInfo `json:"account"`
}

// userAccountResponse is the response payload for account management.
type userAccountResponse struct {
  User *model.User `json:"user"`
}

// newUserAccountResponse creates a response from a user model.
func newUserAccountResponse(p *model.User) *userAccountResponse {
  return &userAccountResponse{
    User: p,
  }
}

// Bind preprocesses a userAccountRequest.
func (d *userAccountRequest) Bind(r *http.Request) error {
  // sending the id via request is invalid as the id should be submitted in the url
  d.User.ID = 0

  d.User.FirstName = strings.TrimSpace(d.User.FirstName)
  d.User.LastName = strings.TrimSpace(d.User.LastName)

  // encrypt password
  hash, err := auth.HashPassword(d.Account.PlainPassword)
  d.User.EncryptedPassword = hash

  return err
}

// Render post-processes a userAccountResponse.
func (u *userAccountResponse) Render(w http.ResponseWriter, r *http.Request) error {
  // nothing to hide
  return nil
}

// bindValidate jointly binds data from json request and validates the model.
func (rs *AccountResource) bindValidate(w http.ResponseWriter, r *http.Request) (*userAccountRequest, *ErrResponse) {
  // start from empty Request
  // Note, this function is used by both POST and PATCH.
  // We should not assume there is an initial account from middle-ware.
  data := &userAccountRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    return nil, ErrBadRequestWithDetails(err)
  }

  // validate final model
  if err := data.User.Validate(); err != nil {
    return nil, ErrBadRequestWithDetails(err)
  }

  return data, nil
}

// Post is the enpoint for creating a new user account.
func (rs *AccountResource) Post(w http.ResponseWriter, r *http.Request) {

  data, errResponse := rs.bindValidate(w, r)
  if errResponse != nil {
    render.Render(w, r, errResponse)
    return
  }

  // create user entry in database
  newUser, err := rs.UserStore.Create(data.User)
  if err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  // return user information of created entry
  if err := render.Render(w, r, newUserAccountResponse(newUser)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// Patch is the endpoint fro updating a specific account with given id.
func (rs *AccountResource) Patch(w http.ResponseWriter, r *http.Request) {

  data, errResponse := rs.bindValidate(w, r)
  if errResponse != nil {
    render.Render(w, r, errResponse)
    return
  }

  // TODO(patwie) verify userID
  if err := rs.UserStore.Update(data.User); err != nil {
    render.Render(w, r, ErrInternalServerError)
    return
  }

  if err := render.Render(w, r, newUserAccountResponse(data.User)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// Get is the endpoint for retrieving a specific user account.
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
