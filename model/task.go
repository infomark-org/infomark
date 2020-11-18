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

package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	null "gopkg.in/guregu/null.v3"
)

// Task is part of an exercise sheet
type Task struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at,omitempty"`

	Name               string      `db:"name"`
	MaxPoints          int         `db:"max_points"`
	PublicDockerImage  null.String `db:"public_docker_image"`
	PrivateDockerImage null.String `db:"private_docker_image"`
}

// TaskRating contains the feedback of students to a task.
type TaskRating struct {
	ID int64 `db:"id"`

	UserID int64 `db:"user_id"`
	TaskID int64 `db:"task_id"`
	Rating int   `db:"rating"`
}

// Validate validates a TaskRating before interacting with a database
// TODO(): should be part of the request
func (body *TaskRating) Validate() error {
	return validation.ValidateStruct(body,
		validation.Field(
			&body.UserID,
			validation.Required,
		),
		validation.Field(
			&body.TaskID,
			validation.Required,
		),
		validation.Field(
			&body.Rating,
			validation.Required,
			validation.Min(1),
			validation.Max(5),
		),
	)
}

// TaskPoints is a performance summary of a student for a given task
type TaskPoints struct {
	AchievablePoints int `db:"achievable_points"`
	AquiredPoints    int `db:"acquired_points"`
	MaxPoints        int `db:"max_points"`
	TaskID           int `db:"task_id"`
}

// Validate validates TaskPoints
func (body *TaskPoints) Validate() error {
	// just a join and read only
	return nil
}

// MissingTask is an item in the list of tasks a user has not uploaded any solution
// (this is used by the dashboard)
type MissingTask struct {
	*Task

	SheetID  int64 `db:"sheet_id"`
	CourseID int64 `db:"course_id"`
}
