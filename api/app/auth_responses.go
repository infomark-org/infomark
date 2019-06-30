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
)

type AuthResponse struct {
	Access struct {
		Token string `json:"token" example:"eyJhbGciOiJIUzI1...rZikwLEI7XhY"`
	} `json:"access"`
	Refresh struct {
		Token string `json:"token" example:"eyJhbGciOiJIUzI1...EYCBjslOydswU"`
	} `json:"refresh"`
}

func (body *AuthResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// .............................................................................
type loginResponse struct {
	Root bool `json:"root" example:"false"`
}

func (body *loginResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// nothing to hide
	return nil
}
