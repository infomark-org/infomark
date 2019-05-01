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

// SheetRequest is the request payload for Sheet management.
type SheetRequest struct {
	Name      string    `json:"name" example:"Blatt 42"`
	PublishAt time.Time `json:"publish_at" example:"auto"`
	DueAt     time.Time `json:"due_at" example:"auto"`
}

// Bind preprocesses a SheetRequest.
func (body *SheetRequest) Bind(r *http.Request) error {

	if body == nil {
		return errors.New("missing \"sheet\" data")
	}

	return body.Validate()
}

// Validate validates a SheetRequest
func (body *SheetRequest) Validate() error {

	err := validation.ValidateStruct(body,
		validation.Field(
			&body.PublishAt,
			validation.Required,
		),
		validation.Field(
			&body.DueAt,
			validation.Required,
		),
		validation.Field(
			&body.Name,
			validation.Required,
		),
	)

	if err == nil {
		if body.DueAt.Sub(body.PublishAt).Seconds() < 0 {
			return errors.New("due_at should be later than publish_at")
		}
	}

	return err
}
