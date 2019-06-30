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

	"github.com/cgtuebingen/infomark-backend/auth"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/cgtuebingen/infomark-backend/symbol"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/spf13/viper"
	null "gopkg.in/guregu/null.v3"
)

// AuthResource specifies user management handler.
type AuthResource struct {
	Stores *Stores
}

// NewAuthResource create and returns a AuthResource.
func NewAuthResource(stores *Stores) *AuthResource {
	return &AuthResource{
		Stores: stores,
	}
}

// RefreshAccessTokenHandler is public endpoint for
// URL: /auth/token
// METHOD: post
// TAG: auth
// REQUEST: loginRequest
// RESPONSE: 201,AuthResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  Refresh or Generate Access token
// DESCRIPTION:
// This endpoint will generate the access token without login credentials
// if the refresh token is given.
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
		targetUser, err := rs.Stores.User.Get(refreshClaims.LoginID)
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

		resp := &AuthResponse{}
		resp.Access.Token = accessToken

		// return access token only
		if err := render.Render(w, r, resp); err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}

	} else {

		// we are given email-password credentials
		data := &loginRequest{}

		// parse JSON request into struct
		if err := render.Bind(r, data); err != nil {
			render.Render(w, r, ErrBadRequestWithDetails(err))
			return
		}

		// does such a user exists with request email adress?
		potentialUser, err := rs.Stores.User.FindByEmail(data.Email)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// does the password match?
		if !auth.CheckPasswordHash(data.PlainPassword, potentialUser.EncryptedPassword) {
			render.Render(w, r, ErrNotFound)
			return
		}

		refreshClaims := authenticate.NewRefreshClaims(potentialUser.ID)
		refreshToken, err := tokenManager.CreateRefreshJWT(refreshClaims)

		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

		accessClaims := authenticate.NewAccessClaims(potentialUser.ID, potentialUser.Root)
		accessToken, err := tokenManager.CreateAccessJWT(accessClaims)

		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

		resp := &AuthResponse{}
		resp.Access.Token = accessToken
		resp.Refresh.Token = refreshToken

		// return user information of created entry
		if err := render.Render(w, r, resp); err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}

	}

}

// LoginHandler is public endpoint for
// URL: /auth/sessions
// METHOD: post
// TAG: auth
// REQUEST: loginRequest
// RESPONSE: 200,loginResponse
// RESPONSE: 400,BadRequest
// SUMMARY:  Start a session
// DESCRIPTION:
// This endpoint will generate the access token without login credentials
// if the refresh token is given.
func (rs *AuthResource) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// we are given email-password credentials

	data := &loginRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	// does such a user exists with request email adress?
	potentialUser, err := rs.Stores.User.FindByEmail(data.Email)
	if err != nil {
		render.Render(w, r, ErrBadRequest)
		return
	}

	// does the password match?
	if !auth.CheckPasswordHash(data.PlainPassword, potentialUser.EncryptedPassword) {
		totalFailedLoginsVec.WithLabelValues().Inc()
		render.Render(w, r, ErrBadRequestWithDetails(errors.New("credentials are wrong")))
		return
	}

	// fmt.Println(potentialUser.ConfirmEmailToken)
	// is the email address confirmed?
	if potentialUser.ConfirmEmailToken.Valid {
		// Valid is true if String is not NULL
		// confirm token `potentialUser.ConfirmEmailToken.String` exists
		render.Render(w, r, ErrBadRequestWithDetails(errors.New("email not confirmed")))
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

// LogoutHandler is public endpoint for
// URL: /auth/sessions
// METHOD: delete
// TAG: auth
// RESPONSE: 200,OK
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  Destroy a session
func (rs *AuthResource) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	accessClaims.DestroyInSession(w, r)
}

// RequestPasswordResetHandler is public endpoint for
// URL: /auth/request_password_reset
// METHOD: post
// TAG: auth
// REQUEST: resetPasswordRequest
// RESPONSE: 200,OK
// RESPONSE: 400,BadRequest
// SUMMARY:  will send an email with password reset link
func (rs *AuthResource) RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	data := &resetPasswordRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	// does such a user exists with request email adress?
	user, err := rs.Stores.User.FindByEmail(data.Email)
	if err != nil {
		render.Render(w, r, ErrBadRequest)
		return
	}

	user.ResetPasswordToken = null.StringFrom(auth.GenerateToken(32))
	rs.Stores.User.Update(user)

	// Send Email to User
	// https://infomark-staging.informatik.uni-tuebingen.de/#/password_reset/example@uni-tuebingen.de/af1ecf6f
	msg, err := email.NewEmailFromTemplate(
		user.Email,
		"Password Reset Instructions",
		email.RequestPasswordTokenTemailTemplateEN,
		map[string]string{
			"first_name":           user.FirstName,
			"last_name":            user.LastName,
			"email_address":        user.Email,
			"reset_password_url":   fmt.Sprintf("%s/#/password_reset", viper.GetString("url")),
			"reset_password_token": user.ResetPasswordToken.String,
		})
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	err = email.DefaultMail.Send(msg)
	// err = email.Send()
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusOK)
}

// UpdatePasswordHandler is public endpoint for
// URL: /auth/update_password
// METHOD: post
// TAG: auth
// REQUEST: updatePasswordRequest
// RESPONSE: 200,OK
// RESPONSE: 400,BadRequest
// SUMMARY:  sets a new password
func (rs *AuthResource) UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	data := &updatePasswordRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	// does such a user exists with request email adress?
	user, err := rs.Stores.User.FindByEmail(data.Email)
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
	if err := rs.Stores.User.Update(user); err != nil {
		// fmt.Println(err)
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusOK)
}

// ConfirmEmailHandler is public endpoint for
// URL: /auth/confirm_email
// METHOD: post
// TAG: auth
// REQUEST: confirmEmailRequest
// RESPONSE: 200,OK
// RESPONSE: 400,BadRequest
// SUMMARY:  handles the confirmation link and activate an account
func (rs *AuthResource) ConfirmEmailHandler(w http.ResponseWriter, r *http.Request) {
	data := &confirmEmailRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	// does such a user exists with request email adress?
	user, err := rs.Stores.User.FindByEmail(data.Email)
	if err != nil {
		render.Render(w, r, ErrBadRequest)
		return
	}

	// compare token
	if user.ConfirmEmailToken.String != data.ConfirmEmailToken {
		render.Render(w, r, ErrBadRequest)
		return
	}

	// token is ok
	user.ConfirmEmailToken = null.String{}
	if err := rs.Stores.User.Update(user); err != nil {
		fmt.Println(err)
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusOK)
}
