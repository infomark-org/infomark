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
	"strings"

	"github.com/cgtuebingen/infomark-backend/auth"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/spf13/viper"
)

// -----------------------------------------------------------------------------

type createUserAccountRequest struct {
	User *struct {
		FirstName     string `json:"first_name"  example:"Max"`
		LastName      string `json:"last_name"  example:"Mustermensch"`
		Email         string `json:"email"  example:"test@uni-tuebingen.de"`
		StudentNumber string `json:"student_number"  example:"0815"`
		Semester      int    `json:"semester"  example:"15"`
		Subject       string `json:"subject"  example:"computer science"`
		Language      string `json:"language"  example:"en"`
	} `json:"user" required:"true"`
	Account *struct {
		Email             string `json:"email" example:"test@uni-tuebingen.de"`
		PlainPassword     string `json:"plain_password" example:"test"`
		EncryptedPassword string `json:"-"`
	} `json:"account" required:"true"`
}

func (m *createUserAccountRequest) Validate() error {

	m.User.FirstName = strings.TrimSpace(m.User.FirstName)
	m.User.LastName = strings.TrimSpace(m.User.LastName)

	m.User.Email = strings.TrimSpace(m.User.Email)
	m.User.Email = strings.ToLower(m.User.Email)

	return validation.ValidateStruct(m.User,
		validation.Field(
			&m.User.FirstName,
			validation.Required,
		),
		validation.Field(
			&m.User.LastName,
			validation.Required,
		),
		validation.Field(
			&m.User.Email,
			validation.Required,
			is.Email,
		),
		validation.Field(
			&m.User.StudentNumber,
			validation.Required,
		),
		validation.Field(
			&m.User.Semester,
			validation.Required,
			validation.Min(1),
		),
		validation.Field(
			&m.User.Subject,
			validation.Required,
		),

		validation.Field(
			&m.User.Language,
			validation.Required,
			validation.Length(2, 2),
		),
	)

}

// Bind preprocesses a createUserAccountRequest.
func (body *createUserAccountRequest) Bind(r *http.Request) (err error) {
	// sending the id via request is invalid as the id should be submitted in the url
	if body.User == nil {
		return errors.New("missing \"user\" data")
	}

	if body.Account == nil {
		return errors.New("missing \"account\" data")
	}

	// check password length
	if len(body.Account.PlainPassword) < viper.GetInt("min_password_length") {
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

type accountRequest struct {
	Account *struct {
		Email             string `json:"email" required:"false"`
		PlainPassword     string `json:"plain_password" required:"false"`
		EncryptedPassword string `json:"-"`
	} `json:"account"`
	OldPlainPassword string `json:"old_plain_password"`
}

// Bind preprocesses a accountRequest.
func (body *accountRequest) Bind(r *http.Request) error {
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
