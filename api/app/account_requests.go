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
	"errors"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/infomark-org/infomark-backend/auth"
	"github.com/infomark-org/infomark-backend/configuration"
)

// -----------------------------------------------------------------------------

type CreateUserAccountRequest struct {
	User *struct {
		FirstName     string `json:"first_name" example:"Max"`
		LastName      string `json:"last_name" example:"Mustermensch"`
		Email         string `json:"email" example:"test@uni-tuebingen.de"`
		StudentNumber string `json:"student_number" example:"0815"`
		Semester      int    `json:"semester" example:"15" minval:"1"`
		Subject       string `json:"subject" example:"computer science"`
		Language      string `json:"language" example:"en" len:"2"`
	} `json:"user" required:"true"`
	Account *struct {
		Email             string `json:"email" example:"test@uni-tuebingen.de"`
		PlainPassword     string `json:"plain_password" example:"test"`
		EncryptedPassword string `json:"-"`
	} `json:"account" required:"true"`
}

func (body *CreateUserAccountRequest) Validate() error {

	body.User.FirstName = strings.TrimSpace(body.User.FirstName)
	body.User.LastName = strings.TrimSpace(body.User.LastName)

	body.User.Email = strings.TrimSpace(body.User.Email)
	body.User.Email = strings.ToLower(body.User.Email)

	return validation.ValidateStruct(body.User,
		validation.Field(
			&body.User.FirstName,
			validation.Required,
		),
		validation.Field(
			&body.User.LastName,
			validation.Required,
		),
		validation.Field(
			&body.User.Email,
			validation.Required,
			is.Email,
		),
		validation.Field(
			&body.User.StudentNumber,
			validation.Required,
		),
		validation.Field(
			&body.User.Semester,
			validation.Required,
			validation.Min(1),
		),
		validation.Field(
			&body.User.Subject,
			validation.Required,
		),

		validation.Field(
			&body.User.Language,
			validation.Required,
			validation.Length(2, 2),
		),
	)

}

// Bind preprocesses a CreateUserAccountRequest.
func (body *CreateUserAccountRequest) Bind(r *http.Request) (err error) {
	// sending the id via request is invalid as the id should be submitted in the url
	if body.User == nil {
		return errors.New("missing \"user\" data")
	}

	if body.Account == nil {
		return errors.New("missing \"account\" data")
	}

	// check password length
	if len(body.Account.PlainPassword) < configuration.Configuration.Server.Authentication.Password.MinLength {
		return errors.New("password too short")
	}

	// encrypt password
	body.Account.EncryptedPassword, err = auth.HashPassword(body.Account.PlainPassword)
	if err != nil {
		return err
	}

	body.User.Email = strings.TrimSpace(body.User.Email)
	body.User.Email = strings.ToLower(body.User.Email)

	body.Account.Email = strings.TrimSpace(body.Account.Email)
	body.Account.Email = strings.ToLower(body.Account.Email)

	if body.User.Email != body.Account.Email {
		return errors.New("email from user does not match email from account")
	}

	return body.Validate()
}

type AccountRequest struct {
	Account *struct {
		Email             string `json:"email" required:"false" example:"other@example.com"`
		PlainPassword     string `json:"plain_password" required:"false" example:"new_password"`
		EncryptedPassword string `json:"-"`
	} `json:"account" required:"true"`
	OldPlainPassword string `json:"old_plain_password" example:"old_password"`
}

// Bind preprocesses a AccountRequest.
func (body *AccountRequest) Bind(r *http.Request) error {
	// this is the only patch function
	// sending the id via request is invalid as the id should be submitted in the url
	if body.Account == nil {
		return errors.New("missing \"account\" data")
	}

	body.Account.Email = strings.TrimSpace(body.Account.Email)
	body.Account.Email = strings.ToLower(body.Account.Email)

	// encrypt new password, when given
	if body.Account.PlainPassword != "" {
		hash, err := auth.HashPassword(body.Account.PlainPassword)
		body.Account.EncryptedPassword = hash
		return err
	}

	// validate email, when given
	return validation.ValidateStruct(body.Account,
		validation.Field(&body.Account.Email, is.Email),
	)
}
