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
)

// -- 0: pending, 1: running, 2: finished
// -- 0 means ok, 1 failed (just like return codes)

type Grade struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at,omitempty"`

	PublicExecutionState  int    `db:"public_execution_state"`
	PrivateExecutionState int    `db:"private_execution_state"`
	PublicTestLog         string `db:"public_test_log"`
	PrivateTestLog        string `db:"private_test_log"`
	PublicTestStatus      int    `db:"public_test_status"`
	PrivateTestStatus     int    `db:"private_test_status"`
	AcquiredPoints        int    `db:"acquired_points"`
	Feedback              string `db:"feedback"`
	TutorID               int64  `db:"tutor_id"`
	SubmissionID          int64  `db:"submission_id"`
	UserID                int64  `db:"user_id,readonly"`
	UserFirstName         string `db:"user_first_name,readonly"`
	UserLastName          string `db:"user_last_name,readonly"`
	UserEmail             string `db:"user_email,readonly"`
}

type MissingGrade struct {
	Grade
	CourseID int64 `db:"course_id"`
	SheetID  int64 `db:"sheet_id"`
	TaskID   int64 `db:"task_id"`
}

type OverviewGrade struct {
	UserID            int64  `db:"user_id"`
	UserFirstName     string `db:"user_first_name"`
	UserLastName      string `db:"user_last_name"`
	UserStudentNumber string `db:"user_student_number"`
	UserEmail         string `db:"user_email"`

	SheetID int64  `db:"sheet_id"`
	Name    string `db:"name"`
	Points  int    `db:"points"`
}
