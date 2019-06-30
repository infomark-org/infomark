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

	"github.com/go-chi/render"
	"github.com/infomark-org/infomark-backend/model"
	"github.com/spf13/viper"
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
func newSubmissionResponse(p *model.Submission, courseID int64) *SubmissionResponse {
	// this does always exists
	fileURL := fmt.Sprintf("%s/api/v1/courses/%d/submissions/%d/file",
		viper.GetString("url"),
		courseID,
		p.ID,
	)

	sr := &SubmissionResponse{
		ID:      p.ID,
		UserID:  p.UserID,
		TaskID:  p.TaskID,
		FileURL: fileURL,
	}

	return sr
}

// newSubmissionListResponse creates a response from a list of Submission models.
func newSubmissionListResponse(Submissions []model.Submission, courseID int64) []render.Renderer {
	list := []render.Renderer{}
	for k := range Submissions {
		list = append(list, newSubmissionResponse(&Submissions[k], courseID))
	}
	return list
}

// Render post-processes a SubmissionResponse.
func (body *SubmissionResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
