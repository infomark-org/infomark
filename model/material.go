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
)

// Material are elements in a course like slides, additional code examples, ...
type Material struct {
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"-" db:"created_at,omitempty"`
	UpdatedAt time.Time `json:"-" db:"updated_at,omitempty"`

	Name      string    `json:"name" db:"name"`
	Filename  string    `json:"filename" db:"filename"` // we keep the original name, since it is meaningful
	Kind      int       `json:"kind" db:"kind"`
	PublishAt time.Time `json:"publish_at" db:"publish_at"`
	LectureAt time.Time `json:"lecture_at" db:"lecture_at"`
}

// Validate validates a `Material` object.
func (m *Material) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(
			&m.Name,
			validation.Required,
		),
		validation.Field(
			&m.Filename,
			validation.Required,
		),
		validation.Field(
			&m.PublishAt,
			validation.Required,
		),
		validation.Field(
			&m.LectureAt,
			validation.Required,
		),
		validation.Field(
			&m.Kind,
			// validation.Required,
			validation.Min(0),
			validation.Max(1),
		),
	)
}
