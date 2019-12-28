// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019 ComputerGraphics Tuebingen
//               2020-present InfoMark.org
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
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

// LoginRequest is the request for the login process containing the password
// and email as identifiers.
type LoginRequest struct {
	Email         string `json:"email" example:"test@uni-tuebingen.de"`
	PlainPassword string `json:"plain_password" example:"test"`
}

// Bind preprocesses a loginRequest.
func (body *LoginRequest) Bind(r *http.Request) error {
	body.Email = strings.TrimSpace(body.Email)
	body.Email = strings.ToLower(body.Email)

	return validation.ValidateStruct(body,
		validation.Field(&body.Email, validation.Required, is.Email),
		validation.Field(&body.PlainPassword, validation.Required),
	)
}

// -----------------------------------------------------------------------------
// ResetPasswordRequest is the request whenever a user forgot his password and wants
// to receive an email with a new one.
type ResetPasswordRequest struct {
	Email string `json:"email" example:"test@uni-tuebingen.de"`
}

func (body *ResetPasswordRequest) Bind(r *http.Request) error {
	body.Email = strings.TrimSpace(body.Email)
	body.Email = strings.ToLower(body.Email)

	return validation.ValidateStruct(body,
		validation.Field(&body.Email, validation.Required, is.Email),
	)
}

// -----------------------------------------------------------------------------
type UpdatePasswordRequest struct {
	Email              string `json:"email" example:"test@uni-tuebingen.de"`
	ResetPasswordToken string `json:"reset_password_token" example:"SDFOI34FZH4HUFH"`
	PlainPassword      string `json:"plain_password" example:"test"`
}

func (body *UpdatePasswordRequest) Bind(r *http.Request) error {
	body.Email = strings.TrimSpace(body.Email)
	body.Email = strings.ToLower(body.Email)

	return validation.ValidateStruct(body,
		validation.Field(&body.Email, validation.Required, is.Email),
		validation.Field(&body.ResetPasswordToken, validation.Required),
		validation.Field(&body.PlainPassword, validation.Required, validation.Length(7, 0)),
	)
}

// -----------------------------------------------------------------------------
type ConfirmEmailRequest struct {
	Email             string `json:"email" example:"test@uni-tuebingen.de"`
	ConfirmEmailToken string `json:"confirmation_token" example:"SDFOI34FZH4HUFH"`
}

func (body *ConfirmEmailRequest) Bind(r *http.Request) error {
	body.Email = strings.TrimSpace(body.Email)
	body.Email = strings.ToLower(body.Email)

	return validation.ValidateStruct(body,
		validation.Field(&body.Email, validation.Required, is.Email),
		validation.Field(&body.ConfirmEmailToken, validation.Required),
	)

}
