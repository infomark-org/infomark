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

	"github.com/cgtuebingen/infomark-backend/symbol"
)

// RawResponse is the response payload for course management.
type RawResponse struct {
	Text string `json:"text" example:"some text"`
}

// Render post-processes a RawResponse.
func (body *RawResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newCourseResponse creates a response from a course model.
func newRawResponse(text string) *RawResponse {
	return &RawResponse{
		Text: text,
	}
}

// versionResponse is the response payload for course management.
type versionResponse struct {
	Commit  string `json:"commit" example:"d725269a8a7498aae1dbb07786bed4c88b002661"`
	Version string `json:"version" example:"1"`
}

// newVersionResponse creates a response from a course model.
func newVersionResponse() *versionResponse {
	return &versionResponse{
		Commit:  symbol.GitCommit,
		Version: symbol.Version.String(),
	}
}

// Render post-processes a versionResponse.
func (body *versionResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
