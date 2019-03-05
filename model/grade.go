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

type Grade struct {
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"-" db:"created_at,omitempty"`
	UpdatedAt time.Time `json:"-" db:"updated_at,omitempty"`

	ExecutionState    int    `json:"execution_state" db:"execution_state"`
	PublicTestLog     string `json:"public_test_log" db:"public_test_log"`
	PrivateTestLog    string `json:"private_test_log" db:"private_test_log"`
	PublicTestStatus  int    `json:"public_test_status" db:"public_test_status"`
	PrivateTestStatus int    `json:"private_test_status" db:"private_test_status"`
	AcquiredPoints    int    `json:"acquired_points" db:"acquired_points"`
	Feedback          string `json:"feedback" db:"feedback"`
	TutorID           int64  `json:"tutor_id" db:"tutor_id"`
	SubmissionID      int64  `json:"submission_id" db:"submission_id"`
}

func (m *Grade) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(
			&m.TutorID,
			validation.Required,
		),
		validation.Field(
			&m.SubmissionID,
			validation.Required,
		),
	)
}
