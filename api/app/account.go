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
  "fmt"
  "io"
  "net/http"
  "os"
  "strings"

  "github.com/cgtuebingen/infomark-backend/auth"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/render"
  "github.com/spf13/viper"
  null "gopkg.in/guregu/null.v3"
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

// Post is the enpoint for creating a new user account.
func (rs *AccountResource) CreateHandler(w http.ResponseWriter, r *http.Request) {
  // start from empty Request
  data := &userAccountRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // validate final model
  if err := data.User.Validate(); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // we will ask the user to confirm their email address
  data.User.ConfirmEmailToken = null.StringFrom(auth.GenerateToken(32))

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

  err = sendConfirmEmailForUser(newUser)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusOK)

}

func sendConfirmEmailForUser(user *model.User) error {
  // send email
  // Send Email to User
  email, err := email.NewEmailFromTemplate(
    user.Email,
    "Confirm Account Instructions",
    "confirm_email.en.txt",
    map[string]string{
      "first_name":          user.FirstName,
      "last_name":           user.LastName,
      "confirm_email_url":   fmt.Sprintf("%s/confirm_email", viper.GetString("url")),
      "confirm_email_token": user.ConfirmEmailToken.String,
    })

  if err != nil {
    return err
  }
  err = email.Send()
  if err != nil {
    return err
  }

  return nil
}

// Patch is the endpoint fro updating the specific account from the requesting
// identity.
func (rs *AccountResource) EditHandler(w http.ResponseWriter, r *http.Request) {

  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  // start from empty Request
  data := &userAccountRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  data.User.ID = accessClaims.LoginID

  // validate final model
  if err := data.User.Validate(); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  dbUser, err := rs.UserStore.Get(data.User.ID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  emailHasChanged := dbUser.Email != data.User.Email

  // make sure email is valid
  if emailHasChanged {
    // we will ask the user to confirm their email address
    data.User.ConfirmEmailToken = null.StringFrom(auth.GenerateToken(32))
  }

  // TODO(patwie) verify userID
  if err := rs.UserStore.Update(data.User); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // make sure email is valid
  if emailHasChanged {
    err = sendConfirmEmailForUser(data.User)
    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }
  }

  if err := render.Render(w, r, newUserAccountResponse(data.User)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// Get is the endpoint for retrieving the specific user account from the requesting
// identity.
func (rs *AccountResource) GetHandler(w http.ResponseWriter, r *http.Request) {
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  user, err := rs.UserStore.Get(accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  if err := render.Render(w, r, newUserResponse(user)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

}

func (rs *AccountResource) GetAvatarHandler(w http.ResponseWriter, r *http.Request) {
  img, err := os.Open("public/avatar.jpg")
  if err != nil {
    panic(err)
    // log.Fatal(err) // perhaps handle this nicer
  }
  defer img.Close()
  w.Header().Set("Content-Type", "image/jpeg") // <-- set the content-type header
  io.Copy(w, img)

}
