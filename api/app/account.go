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
  "io"
  "net/http"
  "os"
  "strconv"
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
  oldUser, err := rs.UserStore.Get(accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  // we gonna alter this struct
  newUser, err := rs.UserStore.Get(accessClaims.LoginID)
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

  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  user, err := rs.UserStore.Get(accessClaims.LoginID)

  avatarPath := "files/avatar.jpg"

  if user.AvatarPath.Valid {
    // Valid is true if String is not NULL
    avatarPath = fmt.Sprintf("files/uploads/%s", user.AvatarPath.String)
  }

  img, err := os.Open(avatarPath)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
  }
  defer img.Close()
  w.Header().Set("Content-Type", "image/jpeg") // <-- set the content-type header
  io.Copy(w, img)

}

func (rs *AccountResource) UpdateAvatarHandler(w http.ResponseWriter, r *http.Request) {

  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  // get current user
  user, err := rs.UserStore.Get(accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  // receive avatar_data frm post request
  if err := r.ParseMultipartForm(32 << 20); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // we are interested in the field "avatar_data"
  file, handler, err := r.FormFile("avatar_data")
  if err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }
  defer file.Close()
  // fmt.Println("Handler Header", handler.Header)
  // fmt.Println("Handler Header", handler.Header["Content-Type"][0])
  // fmt.Println("Handler Header", handler.Filename)

  switch handler.Header["Content-Type"][0] {
  case "image/jpeg", "image/jpg":
    fmt.Println("file ok")

  // case "image/gif":
  // case "image/png":
  // case "application/pdf": // not image, but application !

  default:
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("We support JPG/JPEG only.")))
    return

  }

  if user.AvatarPath.Valid {
    // Valid is true if String is not NULL
    // remove old file
    if err := os.Remove(fmt.Sprintf("files/uploads/%s", user.AvatarPath.String)); err != nil {
      // TODO better logging
      fmt.Println(err)
      // render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      // return
    }
  }

  // the file will live under "files/uploads/"
  targetFile := fmt.Sprintf("avatar-user-%s.jpg", strconv.FormatInt(user.ID, 10))
  targetPath := fmt.Sprintf("./files/uploads/%s", targetFile)

  // try to open new file
  f, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE, 0666)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }
  defer f.Close()

  // copy file from request
  bt, err := io.Copy(f, file)
  fmt.Println(bt)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // update database
  user.AvatarPath = null.StringFrom(targetFile)

  // update database entry
  if err := rs.UserStore.Update(user); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusCreated)

}

func (rs *AccountResource) DeleteAvatarHandler(w http.ResponseWriter, r *http.Request) {
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  // get current user
  user, err := rs.UserStore.Get(accessClaims.LoginID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  if user.AvatarPath.Valid {
    // Valid is true if String is not NULL
    // remove old file
    if err := os.Remove(fmt.Sprintf("files/uploads/%s", user.AvatarPath.String)); err != nil {
      // TODO better logging
      fmt.Println(err)
      // render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      // return
    }
  }

  // update database
  user.AvatarPath = null.String{}

  // update database entry
  if err := rs.UserStore.Update(user); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusOK)
}
