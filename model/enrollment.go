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

import null "gopkg.in/guregu/null.v3"

// "time"
// validation "github.com/go-ozzo/ozzo-validation"

type Enrollment struct {
	ID       int64 `json:"id" db:"id"`
	CourseID int64 `json:"course_id" db:"course_id"`
	Role     int64 `json:"role" db:"role"`
}

type UserCourseEnrollment struct {
	UserID int64 `json:"id" db:"id"`
	// EnrollmentID int64 `json:"id" db:"id"`

	Role int64 `json:"role" db:"role"`

	FirstName     string      `json:"first_name" db:"first_name"`
	LastName      string      `json:"last_name" db:"last_name"`
	AvatarPath    null.String `json:"avatar_url" db:"avatar_path"`
	Email         string      `json:"email" db:"email"`
	StudentNumber string      `json:"student_number" db:"student_number"`
	Semester      int         `json:"semester" db:"semester"`
	Subject       string      `json:"subject" db:"subject"`
	Language      string      `json:"language" db:"language"`
}

func (d *Enrollment) Validate() error {

	return nil

}
