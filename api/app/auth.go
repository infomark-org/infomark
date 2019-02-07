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
  "net/http"
  "strings"

  "github.com/cgtuebingen/infomark-backend/auth"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/go-chi/jwtauth"
  "github.com/go-chi/render"
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

// authRequest is the request payload for account management.
type authRequest struct {
  Email         string `json:"email"`
  PlainPassword string `json:"plain_password"`
}

// authResponse is the response payload for account management.
type authResponse struct {
  AccessToken  string `json:"access_token,omitempty"`
  RefreshToken string `json:"refresh_token,omitempty"`
}

// Bind preprocesses a authRequest.
func (body *authRequest) Bind(r *http.Request) error {
  body.Email = strings.TrimSpace(body.Email)
  body.Email = strings.ToLower(body.Email)
  return nil
}

// Render post-processes a authResponse.
func (body *authResponse) Render(w http.ResponseWriter, r *http.Request) error {
  // nothing to hide
  return nil
}

// bindValidate jointly binds data from json request and validates the model.
func (rs *AuthResource) bindValidate(w http.ResponseWriter, r *http.Request) (*authRequest, *ErrResponse) {
  // get user from middle-ware context
  data := &authRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    return nil, ErrBadRequestWithDetails(err)
  }

  // // validate final model
  // if err := data.User.Validate(); err != nil {
  //   return nil, ErrBadRequestWithDetails(err)
  // }

  return data, nil
}

func (rs *AuthResource) LoginHandler(w http.ResponseWriter, r *http.Request) {
  // we are given email-password credentials
  data, errResponse := rs.bindValidate(w, r)
  if errResponse != nil {
    render.Render(w, r, errResponse)
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

  // user passed all tests
  accessClaims := &authenticate.AccessClaims{
    LoginID: potentialUser.ID,
    Root:    potentialUser.Root,
  }

  fmt.Println("WRITE accessClaims.LoginID", accessClaims.LoginID)
  fmt.Println("WRITE accessClaims.Root", accessClaims.Root)

  accessClaims.WriteToSession(w, r)
}

func (rs *AuthResource) LogoutHandler(w http.ResponseWriter, r *http.Request) {

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
    render.Render(w, r, ErrInternalServerError)
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
      render.Render(w, r, ErrInternalServerError)
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

    // TODO refector same as LoginHandler

    // we are given email-password credentials
    data, errResponse := rs.bindValidate(w, r)
    if errResponse != nil {
      render.Render(w, r, errResponse)
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
      render.Render(w, r, ErrInternalServerError)
      return
    }

    accessToken, err := tokenManager.CreateAccessJWT(authenticate.NewAccessClaims(potentialUser.ID, potentialUser.Root))

    if err != nil {
      render.Render(w, r, ErrInternalServerError)
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
