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
	// validation "github.com/go-ozzo/ozzo-validation"
)

type Sheet struct {
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"-" db:"created_at,omitempty"`
	UpdatedAt time.Time `json:"-" db:"updated_at,omitempty"`

	Name           string    `json:"name" db:"name"`
	Ordering       int       `json:"ordering" db:"ordering"`
	PublishedAt    time.Time `json:"publish_at" db:"publish_at"`
	DueAt          time.Time `json:"due_at" db:"due_at"`
	RequiredPoints int       `json:"required_points" db:"required_points"`
}

func (d *Sheet) Validate() error {

	return nil

}
