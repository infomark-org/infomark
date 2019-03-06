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

package app

import (
	"net/http"

	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/go-chi/render"
)

type (
	// userAccountCreatedResponse is the response payload for account management.
	// description: Account has been created.
	// content:
	//   application/json:
	//     schema:
	//       type: object
	//       required:
	//         - id
	//         - first_name
	//         - last_name
	//         - email
	//         - student_number
	//         - semester
	//         - subject
	//         - language
	//       properties:
	//         id:
	//           type: integer
	//           format: int64
	//         first_name:
	//           type: string
	//         last_name:
	//           type: string
	//         avatar_url:
	//           type: string
	//           format: uri
	//         email:
	//           type: string
	//           format: email
	//         student_number:
	//           type: string
	//         semester:
	//           type: integer
	//           minimum: 1
	//         subject:
	//           type: string
	//         language:
	//           type: string
	//           length: 2
	//           properties:
	userAccountCreatedResponse struct {
		*model.User
	}
)

// Render post-processes a userAccountCreatedResponse.
func (u *userAccountCreatedResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// nothing to hide
	return nil
}

// newUserAccountCreatedResponse creates a response from a user model.
func newUserAccountCreatedResponse(p *model.User) *userAccountCreatedResponse {
	return &userAccountCreatedResponse{
		User: p,
	}
}

// userAccountCreatedResponse is the response payload for account management.
type userEnrollmentResponse struct {
	CourseID int64 `json:"course_id"`
	Role     int64 `json:"role"`
}

// Render post-processes a userAccountCreatedResponse.
func (u *userEnrollmentResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// nothing to hide
	return nil
}

// newCourseResponse creates a response from a course model.
func (rs *AccountResource) newUserEnrollmentResponse(p *model.Enrollment) *userEnrollmentResponse {
	return &userEnrollmentResponse{
		CourseID: p.CourseID,
		Role:     p.Role,
	}
}

func (rs *AccountResource) newUserEnrollmentsResponse(enrollments []model.Enrollment) []render.Renderer {
	// https://stackoverflow.com/a/36463641/7443104
	list := []render.Renderer{}
	for k := range enrollments {
		list = append(list, rs.newUserEnrollmentResponse(&enrollments[k]))
	}

	return list
}
