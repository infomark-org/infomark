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

// UserEnrollmentResponse is the response payload for account management.
type UserEnrollmentResponse struct {
	ID       int64 `json:"id" example:"31"`
	CourseID int64 `json:"course_id" example:"1"`
	Role     int64 `json:"role" example:"1"`
}

// Render post-processes a userAccountCreatedResponse.
func (u *UserEnrollmentResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// nothing to hide
	return nil
}

// newCourseResponse creates a response from a course model.
func (rs *AccountResource) newUserEnrollmentResponse(p *model.Enrollment) *UserEnrollmentResponse {
	return &UserEnrollmentResponse{
		ID:       p.ID,
		CourseID: p.CourseID,
		Role:     p.Role,
	}
}

func (rs *AccountResource) newUserEnrollmentsResponse(enrollments []model.Enrollment) []render.Renderer {
	list := []render.Renderer{}
	for k := range enrollments {
		list = append(list, rs.newUserEnrollmentResponse(&enrollments[k]))
	}

	return list
}
