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

	validation "github.com/go-ozzo/ozzo-validation"
)

// groupRequest is the request payload for course management.
type groupRequest struct {
	TutorID int64 `json:"tutor_id" example:"15"`
	// CourseID    int64  `json:"course_id"`
	Description string `json:"description" example:"Gruppe fuer ersties am Montag im Raum C25435"`
}

// Bind preprocesses a groupRequest.
func (body *groupRequest) Bind(r *http.Request) error {

	if body == nil {
		return errors.New("missing \"group\" data")
	}

	return body.Validate()

}

func (m *groupRequest) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(
			&m.TutorID,
			validation.Required,
		),
		validation.Field(
			&m.Description,
			validation.Required,
		),
	)
}

type groupBidRequest struct {
	Bid int `json:"bid" example:"5" minval:"0" maxval:"10"`
}

// Bind preprocesses a groupRequest.
func (body *groupBidRequest) Bind(r *http.Request) error {
	return body.Validate()
}

func (m *groupBidRequest) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(
			&m.Bid,
			validation.Required,
			validation.Min(0),
			validation.Max(10),
		),
	)
}
