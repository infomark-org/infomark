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
  "errors"
  "fmt"
  "net/http"
  "strconv"
  "strings"

  "github.com/cgtuebingen/infomark-backend/api/helper"
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
  Stores *Stores
}

// NewAccountResource create and returns a AccountResource.
func NewAccountResource(stores *Stores) *AccountResource {
  return &AccountResource{
    Stores: stores,
  }
}

// .............................................................................

type accountInfo struct {
  Email         string `json:"email"`
  PlainPassword string `json:"plain_password"`
}

// userAccountRequest is the request payload for account management chaning name etc....
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
  if d.User != nil {
    d.User.ID = 0

    d.User.FirstName = strings.TrimSpace(d.User.FirstName)
    d.User.LastName = strings.TrimSpace(d.User.LastName)

    // encrypt password
    hash, err := auth.HashPassword(d.Account.PlainPassword)
    d.User.EncryptedPassword = hash
    return err
  }
  return nil
}

// Render post-processes a userAccountResponse.
func (u *userAccountResponse) Render(w http.ResponseWriter, r *http.Request) error {
  // nothing to hide
  return nil
}

// userAccountResponse is the response payload for account management.
type userEnrollmentResponse struct {
  CourseID int64 `json:"course_id"`
  Role     int64 `json:"role"`
}

// newCourseResponse creates a response from a course model.
func (rs *AccountResource) newUserEnrollmentResponse(p *model.Enrollment) *userEnrollmentResponse {

  return &userEnrollmentResponse{
    CourseID: p.CourseID,
    Role:     p.Role,
  }
}

func (rs *AccountResource) newUserEnrollmentsResponse(enrollments []model.Enrollment) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range enrollments {
    list = append(list, rs.newUserEnrollmentResponse(&enrollments[k]))
  }

  return list
}

// Render post-processes a userAccountResponse.
func (u *userEnrollmentResponse) Render(w http.ResponseWriter, r *http.Request) error {
  // nothing to hide
  return nil
}

// .............................................................................

// Post is the enpoint for creating a new user account.
func (rs *AccountResource) CreateHandler(w http.ResponseWriter, r *http.Request) {
  // start from empty Request
  data := &userAccountRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // check password length
  if len(data.Account.PlainPassword) < viper.GetInt("min_password_length") {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("password to short")))
    return
  }

  if data.User == nil {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("user in request is missing")))
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
  newUser, err := rs.Stores.User.Create(data.User)
  if err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusCreated)

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

}

func sendConfirmEmailForUser(user *model.User) error {
  // send email
  // Send Email to User
  msg, err := email.NewEmailFromTemplate(
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
  err = email.DefaultMail.Send(msg)
  if err != nil {
    return err
  }

  return nil
}

// Patch is the endpoint fro updating the specific account from the requesting
// identity.
func (rs *AccountResource) EditHandler(w http.ResponseWriter, r *http.Request) {

  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  // make a backup of old data
  oldUser, err := rs.Stores.User.Get(accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  // we gonna alter this struct
  newUser, err := rs.Stores.User.Get(accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  // start from database data
  data := &userAccountRequest{User: newUser}

  // update struct from JSON request
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // we require the account-part with at least one value
  if data.Account.PlainPassword == "" {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("plain_password in request is missing")))
    return
  }

  // does the submitted password match with the current active password?
  if !auth.CheckPasswordHash(data.Account.PlainPassword, newUser.EncryptedPassword) {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("credentials are wrong")))
    return
  }

  // we require the user-part with at least one value
  if data.User == nil {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("user data in request is missing")))
    return
  }

  data.User.ID = accessClaims.LoginID

  // validate final model
  if err := data.User.Validate(); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  emailHasChanged := newUser.Email != oldUser.Email

  // make sure email is valid
  if emailHasChanged {
    // we will ask the user to confirm their email address
    data.User.ConfirmEmailToken = null.StringFrom(auth.GenerateToken(32))
  }

  // fmt.Println(data.User)

  if err := rs.Stores.User.Update(data.User); err != nil {
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

  render.Status(r, http.StatusNoContent)

  if err := render.Render(w, r, newUserAccountResponse(data.User)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

}

// Get is the endpoint for retrieving the specific user account from the requesting
// identity.
func (rs *AccountResource) GetHandler(w http.ResponseWriter, r *http.Request) {
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  user, err := rs.Stores.User.Get(accessClaims.LoginID)
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

  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  user, err := rs.Stores.User.Get(accessClaims.LoginID)

  if err = helper.NewAvatarFileHandle(user.ID).WriteToBody(w); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
  }

}

func (rs *AccountResource) ChangeAvatarHandler(w http.ResponseWriter, r *http.Request) {

  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  // get current user
  user, err := rs.Stores.User.Get(accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  if err := helper.NewAvatarFileHandle(user.ID).WriteToDisk(r, "file_data"); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
  }

  user.AvatarURL = null.StringFrom(fmt.Sprintf("/api/v1/user/%s/avatar", strconv.FormatInt(user.ID, 10)))
  if err := rs.Stores.User.Update(user); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
  }

  render.Status(r, http.StatusOK)
}

func (rs *AccountResource) DeleteAvatarHandler(w http.ResponseWriter, r *http.Request) {
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  // get current user
  user, err := rs.Stores.User.Get(accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  if err = helper.NewAvatarFileHandle(user.ID).Delete(); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
  }

  render.Status(r, http.StatusNoContent)
}

func (rs *AccountResource) GetEnrollmentsHandler(w http.ResponseWriter, r *http.Request) {
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  // get enrollments
  enrollments, err := rs.Stores.User.GetEnrollments(accessClaims.LoginID)

  // render JSON reponse
  if err = render.RenderList(w, r, rs.newUserEnrollmentsResponse(enrollments)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}
