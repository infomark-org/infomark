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

// ExamRequest is the request payload for exam management.
type ExamRequest struct {
	Name        string    `json:"name" example:"Info 2"`
	Description string    `json:"description" example:"An example exam."`
	ExamTime    time.Time `json:"exam_time" example:"auto"`
}

// Bind preprocesses a ExamRequest.
func (body *ExamRequest) Bind(r *http.Request) error {
	if body == nil {
		return errors.New("missing \"exam\" data")
	}
	return body.Validate()
}

func (body *ExamRequest) Validate() error {
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

// ExamRequest is the request payload for exam management.
type UserExamRequest struct {
	Status int    `json:"status" example:"1"`
	Mark   string `json:"mark" example:"1"`
	UserID int64  `json:"user_id" example:"42"`
}

// Bind preprocesses a ExamRequest.
func (body *UserExamRequest) Bind(r *http.Request) error {
	return nil
}

func (body *UserExamRequest) Validate() error {
	return validation.ValidateStruct(body,
		validation.Field(
			&body.UserID,
			validation.Required,
		),
	)
}
