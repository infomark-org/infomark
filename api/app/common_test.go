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
  "testing"

  "github.com/cgtuebingen/infomark-backend/tape"
  "github.com/franela/goblin"
)

func TestCommon(t *testing.T) {
  g := goblin.Goblin(t)

  tape := &tape.Tape{}

  var (
    r *http.Request
  )

  g.Describe("Common", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      tape.Router, _ = New(tape.DB, false)
    })

    g.It("Should pong", func() {
      r, _ = http.NewRequest("GET", "/api/v1/ping", nil)
      w := tape.PlayRequest(r)
      g.Assert(w.Code).Equal(http.StatusOK)
      g.Assert(w.Body.String()).Equal("pong")

    })

    g.AfterEach(func() {
      tape.AfterEach()
    })
  })

}
