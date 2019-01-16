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
	"regexp"
	"time"

	"github.com/cgtuebingen/infomark-backend/validation"
)

// validate an email
var reEmail = regexp.MustCompile(`(?m)[^@]+@(?:student\.|)uni-tuebingen.de`)

// User represents a registered user.
type User struct {
	// the id for this user.
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`

	Email          string `json:"email"` // Email is the email address for this user.
	PasswordHash   string `json:"-"`     // PasswordHash is the encrypted password.
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	StudentNumber  string `json:"student_number"`
	Specialization string `json:"specialization"`
	Term           string `json:"term"`
	Avatar         string `json:"avatar_url"`

	ResetPasswordToken  string    `json:"-"`
	ResetPasswordSentAt time.Time `json:"-"`
	ConfirmationToken   string    `json:"-"`
	ConfirmationSentAt  time.Time `json:"-"`
	ConfirmedAt         time.Time `json:"-"`

	CurrentSignInAt time.Time `json:"-"`
}

// Validate validates the required fields and formats.
func (u *User) Validate() (*validation.CheckResponses, error) {

	vals := []validation.Check{
		{
			Field: "last_name",
			Value: u.LastName,
			Rules: []validation.Rule{
				&validation.LengthRule{Min: 1, Max: 250},
			},
		},
		{
			Field: "first_name",
			Value: u.FirstName,
			Rules: []validation.Rule{
				&validation.LengthRule{Min: 1, Max: 250},
			},
		},
		{
			Field: "email",
			Value: u.Email,
			Rules: []validation.Rule{
				&validation.MatchRule{Expr: reEmail},
			},
		},
		{
			Field: "password_hash",
			Value: u.PasswordHash,
			Rules: []validation.Rule{
				&validation.LengthRule{Min: 8, Max: 500},
			},
		},
	}

	return validation.Validate(vals)
}
