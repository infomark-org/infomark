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
	"errors"
	"net/http"

	"github.com/cgtuebingen/infomark-backend/model"
)

// TaskRequest is the request payload for Task management.
type TaskRequest struct {
	*model.Task
	ProtectedID int64 `json:"id"`
}

// Bind preprocesses a TaskRequest.
func (body *TaskRequest) Bind(r *http.Request) error {

	if body.Task == nil {
		return errors.New("missing \"task\" data")
	}

	// Sending the id via request-body is invalid.
	// The id should be submitted in the url.
	body.ProtectedID = 0

	return body.Task.Validate()

}

// TaskRequest is the request payload for Task management.
type TaskRatingRequest struct {
	*model.TaskRating
	ProtectedID int64 `json:"id"`
}

// Bind preprocesses a TaskRequest.
func (body *TaskRatingRequest) Bind(r *http.Request) error {

	if body.TaskRating == nil {
		return errors.New("missing \"task_rating\" data")
	}

	// Sending the id via request-body is invalid.
	// The id should be submitted in the url.
	body.ProtectedID = 0
	return body.TaskRating.Validate()

}
