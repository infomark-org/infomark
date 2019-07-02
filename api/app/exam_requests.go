// InfoMark - a platform for managing exams with
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
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

// examRequest is the request payload for exam management.
type examRequest struct {
	Name        string    `json:"name" example:"Info 2"`
	Description string    `json:"description" example:"An example exam."`
	ExamTime    time.Time `json:"exam_time" example:"auto"`
}

// Bind preprocesses a examRequest.
func (body *examRequest) Bind(r *http.Request) error {
	if body == nil {
		return errors.New("missing \"exam\" data")
	}
	return body.Validate()
}

func (body *examRequest) Validate() error {
	return validation.ValidateStruct(body,
		validation.Field(
			&body.Name,
			validation.Required,
		),
		validation.Field(
			&body.Description,
			validation.Required,
		),
		validation.Field(
			&body.ExamTime,
			validation.Required,
		),
	)
}

// // examRequest is the request payload for exam management.
// type examEnrollmentRequest struct {
// 	UserID int64 `json:"user_id" example:"15"`
// }

// // Bind preprocesses a examRequest.
// func (body *examEnrollmentRequest) Bind(r *http.Request) error {
// 	if body == nil {
// 		return errors.New("missing \"exam\" data")
// 	}
// 	return body.Validate()
// }

// func (body *examEnrollmentRequest) Validate() error {
// 	return validation.ValidateStruct(body,
// 		validation.Field(
// 			&body.UserID,
// 			validation.Required,
// 		),
// 	)
// }
