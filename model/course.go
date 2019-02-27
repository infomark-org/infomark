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
	"errors"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

// User holds specific application settings linked to an entity, who can login.
type Course struct {
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"-" db:"created_at,omitempty"`
	UpdatedAt time.Time `json:"-" db:"updated_at,omitempty"`

	Name               string    `json:"name" db:"name"`
	Description        string    `json:"description" db:"description"`
	BeginsAt           time.Time `json:"begins_at" db:"begins_at"`
	EndsAt             time.Time `json:"ends_at" db:"ends_at"`
	RequiredPercentage int       `json:"required_percentage" db:"required_percentage"`
}

func (m *Course) Validate() error {
	if m.EndsAt.Sub(m.BeginsAt).Seconds() < 0 {
		return errors.New("ends_at should be later than begins_at")
	}

	return validation.ValidateStruct(m,
		validation.Field(
			&m.Name,
			validation.Required,
		),
		validation.Field(
			&m.Description,
			validation.Required,
		),
		validation.Field(
			&m.BeginsAt,
			validation.Required,
		),
		validation.Field(
			&m.EndsAt,
			validation.Required,
		),
		validation.Field(
			&m.RequiredPercentage,
			validation.Min(0),
		),
	)
}
