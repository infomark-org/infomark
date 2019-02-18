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

package helper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// similar to gin.H as a neat wrapper
type H map[string]interface{}

func SimulateRequest(payload interface{}, api func(w http.ResponseWriter, r *http.Request)) *httptest.ResponseRecorder {

	payload_json, _ := json.Marshal(payload)
	r, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(payload_json))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api(w, r)
	return w
}
