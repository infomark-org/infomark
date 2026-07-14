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

import null "github.com/guregu/null/v6"

// Enrollment represents a an enrollment-type of a given user
type Enrollment struct {
	ID       int64 `db:"id"`
	CourseID int64 `db:"course_id"`
	Role     int64 `db:"role"`
}

// UserCourse gives enrollment information for multiple users
type UserCourse struct {
	// Both a user id and an enrollment id used to be mapped onto the same
	// db:"id" column via two struct fields. sqlx v1.3+ resolves duplicate
	// db tags to the first declared field (previously the last), which
	// silently redirected the scan target. The queries only select the
	// user id (u.id), so a single field is correct.
	ID int64 `db:"id"`

	Role int64 `db:"role"`

	FirstName     string      `db:"first_name"`
	LastName      string      `db:"last_name"`
	AvatarURL     null.String `db:"avatar_url"`
	Email         string      `db:"email"`
	StudentNumber string      `db:"student_number"`
	Semester      int         `db:"semester"`
	Subject       string      `db:"subject"`
	Language      string      `db:"language"`
}
