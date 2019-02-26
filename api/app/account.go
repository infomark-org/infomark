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
  validation "github.com/go-ozzo/ozzo-validation"
  "github.com/go-ozzo/ozzo-validation/is"
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

// userAccountRequest is the request payload for account management chaning name etc....
type userAccountRequest struct {
  User    *model.User  `json:"user"`
  Account *accountInfo `json:"account"`
}

// Bind preprocesses a userAccountRequest.
func (body *userAccountRequest) Bind(r *http.Request) error {
  // sending the id via request is invalid as the id should be submitted in the url
  if body.User != nil {
    body.User.ID = 0

    body.User.FirstName = strings.TrimSpace(body.User.FirstName)
    body.User.LastName = strings.TrimSpace(body.User.LastName)

    // encrypt password
    hash, err := auth.HashPassword(body.Account.PlainPassword)
    body.User.EncryptedPassword = hash
    return err
  }
  if body.Account != nil {
    body.Account.Email = strings.TrimSpace(body.Account.Email)
    body.Account.Email = strings.ToLower(body.Account.Email)
  }
  return nil
}

type accountInfo struct {
  Email             string `json:"email"`
  PlainPassword     string `json:"plain_password"`
  EncryptedPassword string `json:"-"`
}

// userAccountRequest is the request payload for account management chaning name etc....
type accountRequest struct {
  Account          *accountInfo `json:"account"`
  OldPlainPassword string       `json:"old_plain_password"`
}

// Bind preprocesses a accountRequest.
func (body *accountRequest) Bind(r *http.Request) error {
  // sending the id via request is invalid as the id should be submitted in the url
  if body.Account != nil {

    body.Account.Email = strings.TrimSpace(body.Account.Email)
    body.Account.Email = strings.ToLower(body.Account.Email)

    // encrypt new password, when given
    if body.Account.PlainPassword != "" {
      hash, err := auth.HashPassword(body.Account.PlainPassword)
      body.Account.EncryptedPassword = hash
      return err
    }

    // validate
    err := validation.ValidateStruct(body.Account,
      validation.Field(&body.Account.Email, is.Email),
    )

    return err

  }

  return nil
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

  data.User.AvatarURL = null.String{}

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
      "first_name":            user.FirstName,
      "last_name":             user.LastName,
      "confirm_email_url":     fmt.Sprintf("%s/#/confirmation", viper.GetString("url")),
      "confirm_email_address": user.Email,
      "confirm_email_token":   user.ConfirmEmailToken.String,
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

  // start from database data
  data := &accountRequest{}

  // update struct from JSON request
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // we require the account-part with at least one value
  if data.OldPlainPassword == "" {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("old_plain_password in request is missing")))
    return
  }

  // does the submitted old password match with the current active password?
  if !auth.CheckPasswordHash(data.OldPlainPassword, oldUser.EncryptedPassword) {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("credentials are wrong")))
    return
  }

  // we require the user-part with at least one value
  if data.Account == nil {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("account data in request is missing")))
    return
  }

  emailHasChanged := false
  if data.Account.Email != "" {
    emailHasChanged = data.Account.Email != oldUser.Email
  }
  // emailHasChanged := data.Account.Email != oldUser.Email
  passwordHasChanged := data.Account.PlainPassword != ""

  // we gonna alter this struct
  newUser, err := rs.Stores.User.Get(accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  // make sure email is valid
  if emailHasChanged {
    // we will ask the user to confirm their email address
    newUser.ConfirmEmailToken = null.StringFrom(auth.GenerateToken(32))
    newUser.Email = data.Account.Email
  }

  if passwordHasChanged {
    newUser.EncryptedPassword = data.Account.EncryptedPassword
  }

  if err := rs.Stores.User.Update(newUser); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // make sure email is valid
  if emailHasChanged {
    err = sendConfirmEmailForUser(newUser)
    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }
  }

  render.Status(r, http.StatusNoContent)

  if err := render.Render(w, r, newUserAccountResponse(newUser)); err != nil {
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
