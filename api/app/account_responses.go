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
	null "gopkg.in/guregu/null.v3"
)

type userAccountCreatedResponse struct {
	ID            int64       `json:"id"`
	FirstName     string      `json:"first_name"`
	LastName      string      `json:"last_name"`
	AvatarURL     null.String `json:"avatar_url"`
	Email         string      `json:"email"`
	StudentNumber string      `json:"student_number"`
	Semester      int         `json:"semester"`
	Subject       string      `json:"subject"`
	Language      string      `json:"language"`
	Root          bool        `json:"root"`
}

// Render post-processes a userAccountCreatedResponse.
func (u *userAccountCreatedResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// nothing to hide
	return nil
}

// newUserAccountCreatedResponse creates a response from a user model.
func newUserAccountCreatedResponse(p *model.User) *userAccountCreatedResponse {
	return &userAccountCreatedResponse{
		ID:            p.ID,
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		AvatarURL:     p.AvatarURL,
		Email:         p.Email,
		StudentNumber: p.StudentNumber,
		Semester:      p.Semester,
		Subject:       p.Subject,
		Language:      p.Language,
		Root:          p.Root,
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
