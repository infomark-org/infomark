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

	"github.com/infomark-org/infomark-backend/model"
)

// TaskRatingResponse is the response payload for TaskRating management.
type TaskRatingResponse struct {
	TaskID        int64   `json:"task_id" example:"143"`
	AverageRating float32 `json:"average_rating" example:"3.15"`
	OwnRating     int     `json:"own_rating" example:"4"`
}

// newTaskRatingResponse creates a response from a TaskRating model.
func (rs *TaskRatingResource) newTaskRatingResponse(p *model.TaskRating, averageRating float32) *TaskRatingResponse {

	return &TaskRatingResponse{
		TaskID:        p.TaskID,
		OwnRating:     p.Rating,
		AverageRating: averageRating,
	}
}

// Render post-processes a TaskRatingResponse.
func (body *TaskRatingResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
