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

package api

import (
	"errors"
	"net/http"

	"github.com/cgtuebingen/infomark-backend/router/auth"
	"github.com/cgtuebingen/infomark-backend/router/helper"
	"github.com/cgtuebingen/infomark-backend/store"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// .............................................................................

func LoginRoutes() chi.Router {
	r := chi.NewRouter()

	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Post("/", Login)

	return r
}

// .............................................................................

// LoginRequest is the request payload.
type LoginRequest struct {
	Email         string `json:"email"`
	PlainPassword string `json:"password"`
}

// LoginResponse is the response payload.
type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

func (u *LoginRequest) Bind(r *http.Request) error {
	return nil
}

// render user response
func (u *LoginResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// nothing to hide
	return nil
}

// .............................................................................

// UsersIndex returns all Users.
// POST "/login"
// curl -i -X POST -d '{"Email":"peter.zwegat@uni-tuebingen.de","password": "demo123"}' http://localhost:3000/login
// it returns a valid JWT token
func Login(w http.ResponseWriter, r *http.Request) {

	request_data := &LoginRequest{}

	if err := render.Bind(r, request_data); err != nil {
		render.Render(w, r, helper.NewErrResponse(http.StatusBadRequest, err))
		return
	}

	// find password for an email
	candidate, err := store.DS().GetUserFromEmail(request_data.Email)

	if err != nil {
		render.Render(w, r, helper.NewErrResponse(
			http.StatusUnauthorized,
			errors.New("User is unkown."),
			// err
		))
		return
	}

	// user exists let verify the password
	if helper.CheckPasswordHash(request_data.PlainPassword, candidate.PasswordHash) {
		// ok
		success := true

		token, err := auth.EncodeClaims(auth.CreateClaimsForUserID(int(candidate.ID)))

		if err != nil {
			render.Render(w, r,
				helper.NewErrResponse(http.StatusInternalServerError,
					errors.New("Cannot create JWT token.")))
			return
		}

		response := &LoginResponse{Success: success, Token: token}

		if err := render.Render(w, r, response); err != nil {
			render.Render(w, r, helper.ErrRenderResponse(err))
			return
		}
	} else {
		render.Render(w, r, helper.NewErrResponse(
			http.StatusUnauthorized,
			errors.New("Invalid password.")))
	}

}
