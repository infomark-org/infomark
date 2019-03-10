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
	"fmt"
	"net/http"
	"strconv"

	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/go-chi/render"
)

// .............................................................................

// SubmissionResponse is the response payload for Submission management.
type SubmissionResponse struct {
	ID      int64  `json:"id" example:"61"`
	UserID  int64  `json:"user_id" example:"357"`
	TaskID  int64  `json:"task_id" example:"12"`
	FileURL string `json:"file_url" example:"/api/v1/submissions/61/file"`
}

// newSubmissionResponse creates a response from a Submission model.
func newSubmissionResponse(p *model.Submission) *SubmissionResponse {
	sr := &SubmissionResponse{
		ID:      p.ID,
		UserID:  p.UserID,
		TaskID:  p.TaskID,
		FileURL: fmt.Sprintf("/api/v1/submissions/%s/file", strconv.FormatInt(p.ID, 10)),
	}

	return sr
}

// newSubmissionListResponse creates a response from a list of Submission models.
func newSubmissionListResponse(Submissions []model.Submission) []render.Renderer {
	// https://stackoverflow.com/a/36463641/7443104
	list := []render.Renderer{}
	for k := range Submissions {
		list = append(list, newSubmissionResponse(&Submissions[k]))
	}
	return list
}

// Render post-processes a SubmissionResponse.
func (body *SubmissionResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
