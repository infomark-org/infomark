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

// .............................................................................

// UserResponse is the response payload for user management.
type UserResponse struct {
	ID            int64       `json:"id" example:"1"`
	FirstName     string      `json:"first_name" example:"Max"`
	LastName      string      `json:"last_name" example:"Mustermensch"`
	AvatarURL     null.String `json:"avatar_url" example:"/url/to/file"`
	Email         string      `json:"email" example:"test@unit-tuebingen.de"`
	StudentNumber string      `json:"student_number" example:"0815"`
	Semester      int         `json:"semester" example:"2" minval:"1"`
	Subject       string      `json:"subject" example:"bio informatics"`
	Language      string      `json:"language" example:"en" len:"2"`
	Root          bool        `json:"root" example:"false"`
}

// newUserResponse creates a response from a user model.
func newUserResponse(p *model.User) *UserResponse {
	return &UserResponse{
		ID:            p.ID,
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		AvatarURL:     p.AvatarURL,
		Email:         p.Email,
		StudentNumber: p.StudentNumber,
		Semester:      p.Semester,
		Subject:       p.Subject,
		Language:      p.Language,
	}
}

// newUserListResponse creates a response from a list of user models.
func newUserListResponse(users []model.User) []render.Renderer {
	list := []render.Renderer{}
	for k := range users {
		list = append(list, newUserResponse(&users[k]))
	}

	return list
}

// Render post-processes a UserResponse.
func (u *UserResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// nothing to hide
	return nil
}
