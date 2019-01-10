// Copyright 2019 ComputerGraphics Tuebingen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ==============================================================================
// Authors: Patrick Wieschollek

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cgtuebingen/infomark-backend/router/helper"
	. "github.com/franela/goblin"
)

func SimulateRequest(payload interface{}, api func(w http.ResponseWriter, r *http.Request)) *httptest.ResponseRecorder {

	payload_json, _ := json.Marshal(payload)
	r, _ := http.NewRequest("GET", "dummy", bytes.NewBuffer(payload_json))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api(w, r)

	return w
}

func TestUserCreate(t *testing.T) {

	g := Goblin(t)
	g.Describe("Router", func() {

		g.It("Should create new user", func() {
			w := SimulateRequest(helper.H{
				"ID":       33,
				"LastName": "awesome",
			},
				UserCreate,
			)
			g.Assert(w.Code).Equal(http.StatusOK)
			// g.Assert(resp.Body.String()).Equal(string(expectedBody))
		})

	})

}
