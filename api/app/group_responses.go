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

// GroupResponse is the response payload for Group management.
type GroupResponse struct {
	ID          int64  `json:"id"`
	TutorID     int64  `json:"tutor_id"`
	CourseID    int64  `json:"course_id"`
	Description string `json:"description"`
}

// newGroupResponse creates a response from a Group model.
func (rs *GroupResource) newGroupResponse(p *model.Group) *GroupResponse {
	return &GroupResponse{
		ID:          p.ID,
		TutorID:     p.TutorID,
		CourseID:    p.CourseID,
		Description: p.Description,
	}
}

// newGroupListResponse creates a response from a list of Group models.
func (rs *GroupResource) newGroupListResponse(Groups []model.Group) []render.Renderer {
	// https://stackoverflow.com/a/36463641/7443104
	list := []render.Renderer{}
	for k := range Groups {
		list = append(list, rs.newGroupResponse(&Groups[k]))
	}
	return list
}

// Render post-processes a GroupResponse.
func (body *GroupResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type GroupBidResponse struct {
	Bid int `json:"bid"`
}

// Render post-processes a GroupResponse.
func (body *GroupBidResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
