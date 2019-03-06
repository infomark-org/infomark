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

package fixture

import (
	"net/http"
	"time"
)

type Grade struct {
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"-" db:"created_at,omitempty"`
	UpdatedAt time.Time `json:"-" db:"updated_at,omitempty"`

	ExecutionState    int    `json:"execution_state" db:"execution_state" required:"true"`
	PublicTestLog     string `json:"public_test_log" db:"public_test_log"`
	PrivateTestLog    string `json:"private_test_log" db:"private_test_log"`
	PublicTestStatus  int    `json:"public_test_status" db:"public_test_status"`
	PrivateTestStatus int    `json:"private_test_status" db:"private_test_status"`
	AcquiredPoints    int    `json:"acquired_points" db:"acquired_points"`
	Feedback          string `json:"feedback" db:"feedback"`
	TutorID           int64  `json:"tutor_id" db:"tutor_id"`
	SubmissionID      int64  `json:"submission_id" db:"submission_id"`
}

type (
	// createUserAccountRequest is the request payload when registering a new user.
	// User:
	//   type: object
	//   required:
	//     - id
	//     - first_name
	//     - last_name
	//     - email
	//     - student_number
	//     - semester
	//     - subject
	//     - language
	//   properties:
	//     id:
	//       type: integer
	//       format: int64
	//     first_name:
	//       type: string
	//     last_name:
	//       type: string
	//     avatar_url:
	//       type: string
	//       format: uri
	//     email:
	//       type: string
	//       format: email
	//     student_number:
	//       type: string
	//     semester:
	//       type: integer
	//       minimum: 1
	//     subject:
	//       type: string
	//     language:
	//       type: string
	//       length: 2
	//       properties:
	// Account:
	//   type: object
	//   email:
	//     type: string
	//     format: email
	//   plain_password:
	//     type: string
	//     format: password
	//   required:
	//     - email
	//     - plain_password
	createUserAccountRequest struct {
		User    *model.User  `json:"user"`
		Account *accountInfo `json:"account"`
	}
)

type MissingGrade struct {
	*Grade
	CourseID int64 `json:"course_id" db:"course_id"`
	SheetID  int64 `json:"sheet_id" db:"sheet_id"`
	TaskID   int64 `json:"task_id" db:"task_id"`
}

// LoginHandler is public endpoint for
// URL: /auth/token
// METHOD: post
// SECTION: login
// REQUEST: loginRequest
// RESPONSE: 200,TokenResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 400,Unauthenticated
// RESPONSE: 400,Unauthorized
// SUMMARY:  Receive a access and refresh JSON Web Token
// DESCRIPTION:
// Login with an email and password to get the generated JWT refresh
// and access tokens. Alternatively, if the refresh token is already present
// n the header a new access token is returned.
func LoginHandler(w http.ResponseWriter, r *http.Request) {}

// LogoutHandler is public endpoint for
// URL: /auth/token
// METHOD: delete
// SECTION: login
// REQUEST: loginRequest
// RESPONSE: 200,TokenResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 400,Unauthenticated
// RESPONSE: 400,Unauthorized
// SUMMARY:  Receive a access and refresh JSON Web Token
// DESCRIPTION:
// Login with an email and password to get the generated JWT refresh
// and access tokens. Alternatively, if the refresh token is already present
// n the header a new access token is returned.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {}
