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

	null "gopkg.in/guregu/null.v3"
)

// Group is a database view for a group entity
type Group struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at,omitempty"`
	UpdatedAt time.Time `db:"updated_at,omitempty"`

	TutorID     int64  `db:"tutor_id"`
	CourseID    int64  `db:"course_id"`
	Description string `db:"description"`
}

// GroupEnrollment is a database view for an enrollment of a student into a group.
// Note, Tutors (a person who manage a group) is not enrolled in the group.
type GroupEnrollment struct {
	ID int64 `db:"id"`

	UserID  int64 `db:"user_id"`
	GroupID int64 `db:"group_id"`
}

// GroupWithTutor is a database view of a group including tutor information
type GroupWithTutor struct {
	Group

	TutorFirstName string      `db:"tutor_first_name"`
	TutorLastName  string      `db:"tutor_last_name"`
	TutorAvatarURL null.String `db:"tutor_avatar_url"`
	TutorEmail     string      `db:"tutor_email"`
	TutorLanguage  string      `db:"tutor_language"`
}
