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

type Task struct {
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"-" db:"created_at,omitempty"`
	UpdatedAt time.Time `json:"-" db:"updated_at,omitempty"`

	MaxPoints int `json:"max_points" db:"max_points"`
	// PublicTestPath     string `json:"-" db:"public_test_path"`
	// PrivateTestPath    string `json:"-" db:"private_test_path"`
	PublicDockerImage  string `json:"public_docker_image" db:"public_docker_image"`
	PrivateDockerImage string `json:"private_docker_image" db:"private_docker_image"`
}

func (d *Task) Validate() error {

	return nil

}
