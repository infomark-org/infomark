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

package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	null "gopkg.in/guregu/null.v3"
)

// User holds specific application settings linked to an entity, who can login.
type User struct {
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"-" db:"created_at,omitempty"`
	UpdatedAt time.Time `json:"-" db:"updated_at,omitempty"`

	FirstName     string      `json:"first_name" db:"first_name"`
	LastName      string      `json:"last_name" db:"last_name"`
	AvatarUrl     null.String `json:"avatar_url" db:"avatar_url"`
	Email         string      `json:"email" db:"email"`
	StudentNumber string      `json:"student_number" db:"student_number"`
	Semester      int         `json:"semester" db:"semester"`
	Subject       string      `json:"subject" db:"subject"`

	EncryptedPassword  string      `json:"-" db:"encrypted_password"`
	ResetPasswordToken null.String `json:"-" db:"reset_password_token"`
	ConfirmEmailToken  null.String `json:"-" db:"confirm_email_token"`
	Root               bool        `json:"-" db:"root"`
}

func (d *User) Validate() error {

	return validation.ValidateStruct(d,
		validation.Field(
			&d.FirstName,
			validation.Required,
		),
		validation.Field(
			&d.LastName,
			validation.Required,
		),
		validation.Field(
			&d.Email,
			validation.Required,
		),
		validation.Field(
			&d.StudentNumber,
			validation.Required,
		),
		validation.Field(
			&d.Semester,
			validation.Required,
			validation.Min(1),
		),
		validation.Field(
			&d.Subject,
			validation.Required,
		),
	)

}
