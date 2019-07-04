// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  Infomark Authors
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

// Course holds specific application settings linked to an entity, which
// represents a course
type Exam struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at,omitempty"`

	Name        string    `db:"name"`
	Description string    `db:"description"`
	ExamTime    time.Time `db:"exam_time"`
	CourseID    int64     `db:"course_id"`
}

// Enrollment represents a an enrollment-type of a given user
type UserExam struct {
	ID       int64 `db:"id"`
	UserID   int64 `db:"user_id"`
	ExamID   int64 `db:"exam_id"`
	CourseID int64 `db:"course_id,readonly"`

	Status int    `db:"status"`
	Mark   string `db:"mark"`
}
