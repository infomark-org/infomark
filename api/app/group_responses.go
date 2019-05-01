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

// GroupResponse is the response payload for Group management.
type GroupResponse struct {
	ID          int64  `json:"id" example:"9841"`
	CourseID    int64  `json:"course_id" example:"1"`
	Description string `json:"description" example:"Group every tuesday in room e43"`
	// TutorID     int64  `json:"tutor_id" example:"12"`

	// userResponse
	Tutor *struct {
		ID        int64       `json:"id" example:"1"`
		FirstName string      `json:"first_name" example:"Max"`
		LastName  string      `json:"last_name" example:"Mustermensch"`
		AvatarURL null.String `json:"avatar_url" example:"/url/to/file"`
		Email     string      `json:"email" example:"test@unit-tuebingen.de"`
		Language  string      `json:"language" example:"en" len:"2"`

		StudentNumber string `json:"student_number" example:"0815"`
		Semester      int    `json:"semester" example:"2" minval:"1"`
		Subject       string `json:"subject" example:"bio informatics"`
		Root          bool   `json:"root" example:"false"`
	} `json:"tutor"`
}

// newGroupResponse creates a response from a Group model.
func (rs *GroupResource) newGroupResponse(p *model.Group, t *model.User) *GroupResponse {

	tutor := &struct {
		ID        int64       `json:"id" example:"1"`
		FirstName string      `json:"first_name" example:"Max"`
		LastName  string      `json:"last_name" example:"Mustermensch"`
		AvatarURL null.String `json:"avatar_url" example:"/url/to/file"`
		Email     string      `json:"email" example:"test@unit-tuebingen.de"`
		Language  string      `json:"language" example:"en" len:"2"`

		// just for the frontend
		// TODO(???): change frontend accordingly
		StudentNumber string `json:"student_number" example:"0815"`
		Semester      int    `json:"semester" example:"2" minval:"1"`
		Subject       string `json:"subject" example:"bio informatics"`
		Root          bool   `json:"root" example:"false"`
	}{
		ID:        t.ID,
		FirstName: t.FirstName,
		LastName:  t.LastName,
		AvatarURL: t.AvatarURL,
		Email:     t.Email,
		Language:  t.Language,
	}

	return &GroupResponse{
		ID: p.ID,
		// TutorID:     p.TutorID,
		Tutor:       tutor,
		CourseID:    p.CourseID,
		Description: p.Description,
	}
}

// newGroupListResponse creates a response from a list of Group models.
func (rs *GroupResource) newGroupListResponse(Groups []model.GroupWithTutor) []render.Renderer {
	list := []render.Renderer{}
	for k := range Groups {
		// TODO(patwie): refactor this
		tutor := &model.User{
			ID:        Groups[k].TutorID,
			FirstName: Groups[k].TutorFirstName,
			LastName:  Groups[k].TutorLastName,
			AvatarURL: Groups[k].TutorAvatarURL,
			Email:     Groups[k].TutorEmail,
			Language:  Groups[k].TutorLanguage,
		}

		group := &model.Group{
			ID:          Groups[k].ID,
			CourseID:    Groups[k].CourseID,
			Description: Groups[k].Description,
		}
		list = append(list, rs.newGroupResponse(group, tutor))
	}
	return list
}

// Render post-processes a GroupResponse.
func (body *GroupResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// GroupBidResponse returns the preference of a user to a exercise group
type GroupBidResponse struct {
	Bid int `json:"bid" example:"4" minval:"0" maxval:"10"`
}

// Render post-processes a GroupResponse.
func (body *GroupBidResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
