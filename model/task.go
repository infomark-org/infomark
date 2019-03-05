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
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"-" db:"created_at,omitempty"`
	UpdatedAt time.Time `json:"-" db:"updated_at,omitempty"`

	MaxPoints          int    `json:"max_points" db:"max_points"`
	PublicDockerImage  string `json:"public_docker_image" db:"public_docker_image"`
	PrivateDockerImage string `json:"private_docker_image" db:"private_docker_image"`
}

func (m *Task) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(
			&m.MaxPoints,
			validation.Min(0),
		),
	)
}

type TaskRating struct {
	ID int64 `json:"id" db:"id"`

	UserID int64 `json:"user_id" db:"user_id"`
	TaskID int64 `json:"task_id" db:"task_id"`
	Rating int   `json:"rating" db:"rating"`
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
	AquiredPoints int `json:"acquired_points" db:"acquired_points"`
	MaxPoints     int `json:"max_points" db:"max_points"`
	TaskID        int `json:"task_id" db:"task_id"`
}

func (d *TaskPoints) Validate() error {
	// just a join and read only
	return nil
}

type MissingTask struct {
	*Task

	SheetID  int64 `json:"sheet_id" db:"sheet_id"`
	CourseID int64 `json:"course_id" db:"course_id"`
}
