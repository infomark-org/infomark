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
)

// UserRequest is the request payload for user management.
type UserRequest struct {
	FirstName     string `json:"first_name" example:"Max"`
	LastName      string `json:"last_name" example:"Mustermensch"`
	Email         string `json:"email" example:"test@unit-tuebingen.de"`
	StudentNumber string `json:"student_number" example:"0815"`
	Semester      int    `json:"semester" example:"2" minval:"1"`
	Subject       string `json:"subject" example:"bio informatics"`
	Language      string `json:"language" example:"en" len:"2"`
	PlainPassword string `json:"plain_password" example:"new_password" required:"false"`
}

// Bind preprocesses a UserRequest.
func (body *UserRequest) Bind(r *http.Request) error {

	if body == nil {
		return errors.New("missing \"user\" data")
	}

	return body.Validate()

}

func (body *UserRequest) Validate() error {

	body.FirstName = strings.TrimSpace(body.FirstName)
	body.LastName = strings.TrimSpace(body.LastName)

	body.Email = strings.TrimSpace(body.Email)
	body.Email = strings.ToLower(body.Email)

	return validation.ValidateStruct(body,
		validation.Field(
			&body.FirstName,
			validation.Required,
		),
		validation.Field(
			&body.LastName,
			validation.Required,
		),
		validation.Field(
			&body.Email,
			validation.Required,
			is.Email,
		),
		validation.Field(
			&body.StudentNumber,
			validation.Required,
		),
		validation.Field(
			&body.Semester,
			validation.Required,
			validation.Min(1),
		),
		validation.Field(
			&body.Subject,
			validation.Required,
		),

		validation.Field(
			&body.Language,
			validation.Required,
			validation.Length(2, 2),
		),
	)

}

// UserMeRequest is the request payload for user management.
type UserMeRequest struct {
	FirstName string `json:"first_name" example:"Max"`
	LastName  string `json:"last_name" example:"Mustermensch"`
	// Email         string `json:"email" example:"test@unit-tuebingen.de"`
	StudentNumber string `json:"student_number" example:"0815"`
	Semester      int    `json:"semester" example:"2" minval:"1"`
	Subject       string `json:"subject" example:"bio informatics"`
	Language      string `json:"language" example:"en" len:"2"`
	// PlainPassword string `json:"plain_password" example:"new_password"`
}

// Bind preprocesses a UserMeRequest.
func (body *UserMeRequest) Bind(r *http.Request) error {

	if body == nil {
		return errors.New("missing \"user\" data")
	}

	return body.Validate()

}

func (body *UserMeRequest) Validate() error {

	body.FirstName = strings.TrimSpace(body.FirstName)
	body.LastName = strings.TrimSpace(body.LastName)

	return validation.ValidateStruct(body,
		validation.Field(
			&body.FirstName,
			validation.Required,
		),
		validation.Field(
			&body.LastName,
			validation.Required,
		),

		validation.Field(
			&body.StudentNumber,
			validation.Required,
		),
		validation.Field(
			&body.Semester,
			validation.Required,
			validation.Min(1),
		),
		validation.Field(
			&body.Subject,
			validation.Required,
		),

		validation.Field(
			&body.Language,
			validation.Required,
			validation.Length(2, 2),
		),
	)

}
