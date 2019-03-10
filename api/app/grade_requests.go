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

	validation "github.com/go-ozzo/ozzo-validation"
)

// GradeRequest is the request payload for submission management.
// This will mostly handle the Feedback from tutors. Other changes like
// execution state will be handle internally and is not user-facing.
type GradeRequest struct {
	// SubmissionID   int64  `json:"submission_id"`
	AcquiredPoints int    `json:"acquired_points" example:"13"`
	Feedback       string `json:"feedback" example:"Das war gut"`
}

// Bind preprocesses a GradeRequest.
func (body *GradeRequest) Bind(r *http.Request) error {
	return body.Validate()
}

func (m *GradeRequest) Validate() error {
	return validation.ValidateStruct(m,
		// validation.Field(
		// 	&m.SubmissionID,
		// 	validation.Required,
		// ),
		validation.Field(
			&m.AcquiredPoints,
			validation.Required,
		),
		validation.Field(
			&m.Feedback,
			validation.Required,
		),
	)
}

type GradeFromWorkerRequest struct {
	Log    string `json:"log" example:"failed in line ..."`
	Status int    `json:"status" example:"1"`
}

// Bind preprocesses a GradeRequest.
func (body *GradeFromWorkerRequest) Bind(r *http.Request) error {
	fmt.Println(body)
	fmt.Println(body.Validate())

	return body.Validate()
}

func (m *GradeFromWorkerRequest) Validate() error {
	return validation.ValidateStruct(m,
		validation.Field(
			&m.Log,
			validation.Required,
		),
		// validation.Field(
		// 	&m.Status,
		// 	validation.Required,
		// ),
	)
}
