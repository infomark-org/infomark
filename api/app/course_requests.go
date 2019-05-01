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
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

// courseRequest is the request payload for course management.
type courseRequest struct {
	Name               string    `json:"name" example:"Info 2"`
	Description        string    `json:"description" example:"An example course."`
	BeginsAt           time.Time `json:"begins_at" example:"auto"`
	EndsAt             time.Time `json:"ends_at" example:"auto"`
	RequiredPercentage int       `json:"required_percentage" example:"80"`
}

// Bind preprocesses a courseRequest.
func (body *courseRequest) Bind(r *http.Request) error {

	if body == nil {
		return errors.New("missing \"course\" data")
	}

	return body.Validate()

}

func (body *courseRequest) Validate() error {
	if body.EndsAt.Sub(body.BeginsAt).Seconds() < 0 {
		return errors.New("ends_at should be later than begins_at")
	}

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
			&body.BeginsAt,
			validation.Required,
		),
		validation.Field(
			&body.EndsAt,
			validation.Required,
		),
		validation.Field(
			&body.RequiredPercentage,
			validation.Min(0),
		),
	)
}

type changeRoleInCourseRequest struct {
	Role int `json:"role" example:"0"`
}

func (body *changeRoleInCourseRequest) Bind(r *http.Request) error {
	return nil
}
