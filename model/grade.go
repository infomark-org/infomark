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

type Grade struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at,omitempty"`

	ExecutionState    int    `db:"execution_state"`
	PublicTestLog     string `db:"public_test_log"`
	PrivateTestLog    string `db:"private_test_log"`
	PublicTestStatus  int    `db:"public_test_status"`
	PrivateTestStatus int    `db:"private_test_status"`
	AcquiredPoints    int    `db:"acquired_points"`
	Feedback          string `db:"feedback"`
	TutorID           int64  `db:"tutor_id"`
	SubmissionID      int64  `db:"submission_id"`
}

type MissingGrade struct {
	ID                int64     `db:"id"`
	CreatedAt         time.Time `db:"created_at,omitempty"`
	UpdatedAt         time.Time `db:"updated_at,omitempty"`
	ExecutionState    int       `db:"execution_state"`
	PublicTestLog     string    `db:"public_test_log"`
	PrivateTestLog    string    `db:"private_test_log"`
	PublicTestStatus  int       `db:"public_test_status"`
	PrivateTestStatus int       `db:"private_test_status"`
	AcquiredPoints    int       `db:"acquired_points"`
	Feedback          string    `db:"feedback"`
	TutorID           int64     `db:"tutor_id"`
	SubmissionID      int64     `db:"submission_id"`

	CourseID int64 `db:"course_id"`
	SheetID  int64 `db:"sheet_id"`
	TaskID   int64 `db:"task_id"`
}
