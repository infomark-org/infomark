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

import null "gopkg.in/guregu/null.v3"

// Enrollment represents a an enrollment-type of a given user
type Enrollment struct {
	ID       int64 `db:"id"`
	CourseID int64 `db:"course_id"`
	Role     int64 `db:"role"`
}

// UserCourse gives enrollment information for multiple users
type UserCourse struct {
	UserID int64 `db:"id"`
	ID     int64 `db:"id"`

	Role int64 `db:"role"`

	FirstName     string      `db:"first_name"`
	LastName      string      `db:"last_name"`
	AvatarURL     null.String `db:"avatar_url"`
	Email         string      `db:"email"`
	StudentNumber string      `db:"student_number"`
	Semester      int         `db:"semester"`
	Subject       string      `db:"subject"`
	Language      string      `db:"language"`

	TeamID        null.Int `db:"team_id"`
	TeamConfirmed bool     `db:"confirmed_team"`
}
