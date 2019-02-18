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
  "strings"

  "github.com/cgtuebingen/infomark-backend/auth"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/go-chi/jwtauth"
  "github.com/go-chi/render"
  validation "github.com/go-ozzo/ozzo-validation"
  "github.com/go-ozzo/ozzo-validation/is"
  "github.com/spf13/viper"
  null "gopkg.in/guregu/null.v3"
)

// AuthResource specifies user management handler.
type AuthResource struct {
  UserStore UserStore
}

// NewAuthResource create and returns a AuthResource.
func NewAuthResource(userStore UserStore) *AuthResource {
  return &AuthResource{
    UserStore: userStore,
  }
}

// .............................................................................

type authRequest struct {
  Email         string `json:"email"`
  PlainPassword string `json:"plain_password"`
}

type authResponse struct {
  AccessToken  string `json:"access_token,omitempty"`
  RefreshToken string `json:"refresh_token,omitempty"`
}

// Bind preprocesses a authRequest.
func (body *authRequest) Bind(r *http.Request) error {
  body.Email = strings.TrimSpace(body.Email)
  body.Email = strings.ToLower(body.Email)

  err := validation.ValidateStruct(body,
    validation.Field(&body.Email, validation.Required, is.Email),
    validation.Field(&body.PlainPassword,
      validation.Required,
    // validation.Length(7, 0)
    ),
  )
  if err != nil {
    return err
  }

  return nil
}

func (body *authResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// .............................................................................
type loginResponse struct {
  Root bool `json:"root"`
}

func (body *loginResponse) Render(w http.ResponseWriter, r *http.Request) error {
  // nothing to hide
  return nil
}

// .............................................................................
type resetPasswordRequest struct {
  Email string `json:"email"`
}

func (body *resetPasswordRequest) Bind(r *http.Request) error {
  body.Email = strings.TrimSpace(body.Email)
  body.Email = strings.ToLower(body.Email)

  err := validation.ValidateStruct(body,
    validation.Field(&body.Email, validation.Required, is.Email),
  )
  if err != nil {
    return err
  }

  return nil
}

// .............................................................................
type updatePasswordRequest struct {
  Email              string `json:"email"`
  ResetPasswordToken string `json:"reset_password_token"`
  PlainPassword      string `json:"plain_password"`
}

func (body *updatePasswordRequest) Bind(r *http.Request) error {
  body.Email = strings.TrimSpace(body.Email)
  body.Email = strings.ToLower(body.Email)

  err := validation.ValidateStruct(body,
    validation.Field(&body.Email, validation.Required, is.Email),
    validation.Field(&body.ResetPasswordToken, validation.Required),
    validation.Field(&body.PlainPassword,
      validation.Required,
      validation.Length(7, 0)),
  )
  if err != nil {
    return err
  }

  return nil
}

// .............................................................................
type confirmEmailRequest struct {
  Email             string `json:"email"`
  ConfirmEmailToken string `json:"confirmation_token"`
}

func (body *confirmEmailRequest) Bind(r *http.Request) error {
  body.Email = strings.TrimSpace(body.Email)
  body.Email = strings.ToLower(body.Email)
  return nil
}

// .............................................................................

func (rs *AuthResource) LoginHandler(w http.ResponseWriter, r *http.Request) {
  // we are given email-password credentials

  data := &authRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // does such a user exists with request email adress?
  potentialUser, err := rs.UserStore.FindByEmail(data.Email)
  if err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // does the password match?
  if !auth.CheckPasswordHash(data.PlainPassword, potentialUser.EncryptedPassword) {
    render.Render(w, r, ErrBadRequestWithDetails(errors.New("credentials are wrong")))
    return
  }

  // fmt.Println(potentialUser.ConfirmEmailToken)
  // is the email address confirmed?
  if potentialUser.ConfirmEmailToken.Valid {
    // Valid is true if String is not NULL
    // confirm token `potentialUser.ConfirmEmailToken.String` exists
    render.Render(w, r, ErrRender(errors.New("email not confirmed")))
    return
  }

  // user passed all tests
  accessClaims := &authenticate.AccessClaims{
    LoginID: potentialUser.ID,
    Root:    potentialUser.Root,
  }

  // fmt.Println("WRITE accessClaims.LoginID", accessClaims.LoginID)
  // fmt.Println("WRITE accessClaims.Root", accessClaims.Root)

  w = accessClaims.WriteToSession(w, r)

  resp := &loginResponse{Root: potentialUser.Root}
  // return access token only
  if err := render.Render(w, r, resp); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

}

func (rs *AuthResource) LogoutHandler(w http.ResponseWriter, r *http.Request) {
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  accessClaims.DestroyInSession(w, r)
}

func (rs *AuthResource) RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
  data := &resetPasswordRequest{}
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // does such a user exists with request email adress?
  user, err := rs.UserStore.FindByEmail(data.Email)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  user.ResetPasswordToken = null.StringFrom(auth.GenerateToken(32))
  rs.UserStore.Update(user)

  // Send Email to User
  email, err := email.NewEmailFromTemplate(
    user.Email,
    "Password Reset Instructions",
    "request_password_token.en.txt",
    map[string]string{
      "first_name":           user.FirstName,
      "last_name":            user.LastName,
      "reset_password_url":   fmt.Sprintf("%s/reset_password", viper.GetString("url")),
      "reset_password_token": user.ResetPasswordToken.String,
    })
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  err = email.Send()
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusOK)
}

func (rs *AuthResource) UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
  data := &updatePasswordRequest{}
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // does such a user exists with request email adress?
  user, err := rs.UserStore.FindByEmail(data.Email)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  // compare token
  if user.ResetPasswordToken.String != data.ResetPasswordToken {
    render.Render(w, r, ErrBadRequest)
    return
  }

  // token is ok, remove token and set new password
  user.ResetPasswordToken = null.String{}
  user.EncryptedPassword, err = auth.HashPassword(data.PlainPassword)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // fmt.Println(user)
  if err := rs.UserStore.Update(user); err != nil {
    // fmt.Println(err)
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusOK)
}

func (rs *AuthResource) ConfirmEmailHandler(w http.ResponseWriter, r *http.Request) {
  data := &confirmEmailRequest{}
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // does such a user exists with request email adress?
  user, err := rs.UserStore.FindByEmail(data.Email)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  // compare token
  if user.ConfirmEmailToken.String != data.ConfirmEmailToken {
    render.Render(w, r, ErrBadRequest)
    return
  }

  // token is ok
  user.ConfirmEmailToken = null.String{}
  // fmt.Println(user)
  if err := rs.UserStore.Update(user); err != nil {
    fmt.Println(err)
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// Post is endpoint
func (rs *AuthResource) RefreshAccessTokenHandler(w http.ResponseWriter, r *http.Request) {
  // Login with your username and password to get the generated JWT refresh and
  // access tokens. Alternatively, if the refresh token is already present in
  // the header the access token is returned.
  // This is a corner case, so we do not rely on middleware here

  // access the underlying JWT functions
  tokenManager, err := authenticate.NewTokenAuth()
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // we test wether there is already a JWT Token
  if authenticate.HasHeaderToken(r) {

    // parse string from header
    tokenStr := jwtauth.TokenFromHeader(r)

    // ok, there is a token in the header
    refreshClaims := &authenticate.RefreshClaims{}
    err := refreshClaims.ParseRefreshClaimsFromToken(tokenStr)

    if err != nil {
      // something went wrong during getting the claims
      fmt.Println(err)
      render.Render(w, r, ErrUnauthorized)
      return
    }

    fmt.Println("refreshClaims.LoginID", refreshClaims.LoginID)
    fmt.Println("refreshClaims.AccessNotRefresh", refreshClaims.AccessNotRefresh)

    // everything ok
    targetUser, err := rs.UserStore.Get(refreshClaims.LoginID)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // we just need to return an access-token
    accessToken, err := tokenManager.CreateAccessJWT(authenticate.NewAccessClaims(targetUser.ID, targetUser.Root))
    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }

    resp := &authResponse{
      AccessToken: accessToken,
    }

    // return access token only
    if err := render.Render(w, r, resp); err != nil {
      render.Render(w, r, ErrRender(err))
      return
    }

  } else {

    // we are given email-password credentials
    data := &authRequest{}

    // parse JSON request into struct
    if err := render.Bind(r, data); err != nil {
      render.Render(w, r, ErrBadRequestWithDetails(err))
      return
    }

    // does such a user exists with request email adress?
    potentialUser, err := rs.UserStore.FindByEmail(data.Email)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // does the password match?
    if !auth.CheckPasswordHash(data.PlainPassword, potentialUser.EncryptedPassword) {
      render.Render(w, r, ErrNotFound)
      return
    }

    refreshToken, err := tokenManager.CreateRefreshJWT(authenticate.NewRefreshClaims(potentialUser.ID))

    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }

    accessToken, err := tokenManager.CreateAccessJWT(authenticate.NewAccessClaims(potentialUser.ID, potentialUser.Root))

    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }

    resp := &authResponse{
      AccessToken:  accessToken,
      RefreshToken: refreshToken,
    }

    // return user information of created entry
    if err := render.Render(w, r, resp); err != nil {
      render.Render(w, r, ErrRender(err))
      return
    }

  }

}
