// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019 ComputerGraphics Tuebingen
//               2020-present InfoMark.org
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
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/franela/goblin"
	"github.com/infomark-org/infomark/configuration"
	otape "github.com/infomark-org/infomark/tape"
)

func TestMetrics(t *testing.T) {

	g := goblin.Goblin(t)

	tape := NewTape()

	g.Describe("Metrics", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
		})

		g.It("Should required credentials", func() {
			w := tape.Get("/metrics")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			h := make(map[string]interface{})
			r := otape.BuildDataRequest("GET", "/metrics", h)

			user := configuration.Configuration.Server.Services.Prometheus.User
			password := configuration.Configuration.Server.Services.Prometheus.Password

			authorization := fmt.Sprintf("%s:%s", user, password)
			authorization = base64.StdEncoding.EncodeToString([]byte(authorization))

			r.Header.Set("Authorization", "Basic "+authorization)
			w = tape.PlayRequest(r)
			g.Assert(w.Code).Equal(http.StatusOK)

		})

		g.AfterEach(func() {
			tape.AfterEach()
		})

	})

}
