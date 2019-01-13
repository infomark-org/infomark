// Copyright 2019 ComputerGraphics Tuebingen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ==============================================================================
// Authors: Patrick Wieschollek

package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/cgtuebingen/infomark-backend/router/auth"
	"github.com/cgtuebingen/infomark-backend/router/helper"
	"github.com/cgtuebingen/infomark-backend/store"
	"github.com/go-chi/render"
)

// .............................................................................

// UserRequest is the request payload for User data model.
type LoginRequest struct {
	Email         string `json:"email"`
	PlainPassword string `json:"password"`
	// PasswordHash  string `json:"-"`
}

// LoginResponse is the response payload for the User data model.
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

	// find password for email
	candidate, _ := store.DS().GetUserFromEmail(request_data.Email)

	if candidate.Email == "" {
		render.Render(w, r, helper.NewErrResponse(
			http.StatusBadRequest,
			errors.New("User is unkown.")))
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
			http.StatusBadRequest,
			errors.New("Invalid password.")))
	}

	log.Println(candidate)

	// if err != nil{
	// 	render.Render(w, r, helper.NewErrResponse(http.StatusBadRequest, "err"))
	// }

	// log.Println(request_data)

}
