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

type Task struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at,omitempty"`

	Name               string `db:"name"`
	MaxPoints          int    `db:"max_points"`
	PublicDockerImage  string `db:"public_docker_image"`
	PrivateDockerImage string `db:"private_docker_image"`
}

type TaskRating struct {
	ID int64 `db:"id"`

	UserID int64 `db:"user_id"`
	TaskID int64 `db:"task_id"`
	Rating int   `db:"rating"`
}

func (m *TaskRating) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(
			&m.UserID,
			validation.Required,
		),
		validation.Field(
			&m.TaskID,
			validation.Required,
		),
		validation.Field(
			&m.Rating,
			validation.Required,
			validation.Min(1),
			validation.Max(5),
		),
	)
}

type TaskPoints struct {
	AquiredPoints int `db:"acquired_points"`
	MaxPoints     int `db:"max_points"`
	TaskID        int `db:"task_id"`
}

func (d *TaskPoints) Validate() error {
	// just a join and read only
	return nil
}

type MissingTask struct {
	*Task

	SheetID  int64 `db:"sheet_id"`
	CourseID int64 `db:"course_id"`
}
