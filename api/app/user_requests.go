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
	"net/http"

	"github.com/cgtuebingen/infomark-backend/auth"
	"github.com/cgtuebingen/infomark-backend/model"
)

// userRequest is the request payload for user management.
type userRequest struct {
	*model.User
	ProtectedID     int64  `json:"id"`
	ProtectedAvatar string `json:"avatar_url"`
	PlainPassword   string `json:"plain_password"`
}

// Bind preprocesses a userRequest.
func (body *userRequest) Bind(r *http.Request) error {

	if body.User == nil {
		return errors.New("missing user data")
	}

	if err := body.User.Validate(); err != nil {
		return err
	}

	// Sending the id via request-body is invalid.
	// The id should be submitted in the url.
	body.ProtectedID = 0

	// Encrypt plain password
	hash, err := auth.HashPassword(body.PlainPassword)
	body.User.EncryptedPassword = hash

	return err

}

// -----------------------------------------------------------------------------

// userRequest is the request payload for user management.
type userMeRequest struct {
	*model.User
	ProtectedID     int64  `json:"id"`
	PlainPassword   string `json:"plain_password"`
	ProtectedEmail  string `json:"email"`
	ProtectedAvatar string `json:"avatar_url"`
}

// Bind preprocesses a userMeRequest.
func (body *userMeRequest) Bind(r *http.Request) error {

	if body.User == nil {
		return errors.New("missing user data")
	}

	if err := body.User.Validate(); err != nil {
		return err
	}

	// Sending the id via request-body is invalid.
	// The id should be submitted in the url.
	body.ProtectedID = 0

	// Encrypt plain password
	hash, err := auth.HashPassword(body.PlainPassword)
	body.EncryptedPassword = hash

	return err

}
