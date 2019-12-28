// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019 ComputerGraphics Tuebingen
//               2020-present InfoMark.org
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
)

// TaskRatingRequest is the request payload when students give feedback rating
type TaskRatingRequest struct {
	Rating int `json:"rating" example:"2"`
}

// Bind preprocesses a TaskRatingRequest.
func (body *TaskRatingRequest) Bind(r *http.Request) error {

	if body == nil {
		return errors.New("missing \"task_rating\" data")
	}

	return body.Validate()

}

// Validate validates a TaskRatingRequest.
func (body *TaskRatingRequest) Validate() error {
	return nil
}
