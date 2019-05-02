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
	"fmt"
	"time"

	null "gopkg.in/guregu/null.v3"
)

// User holds specific application settings linked to an entity, who can login.
type User struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at,omitempty"`

	FirstName     string      `db:"first_name"`
	LastName      string      `db:"last_name"`
	AvatarURL     null.String `db:"avatar_url"`
	Email         string      `db:"email"`
	StudentNumber string      `db:"student_number"`
	Semester      int         `db:"semester"`
	Subject       string      `db:"subject"`
	Language      string      `db:"language"`

	EncryptedPassword  string      `db:"encrypted_password"`
	ResetPasswordToken null.String `db:"reset_password_token"`
	ConfirmEmailToken  null.String `db:"confirm_email_token"`
	Root               bool        `db:"root"`
}

// FullName is a wrapper for returning the fullname of a user
func (m *User) FullName() string {
	return fmt.Sprintf("%s %s", m.FirstName, m.LastName)
}
