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
	"time"

	"github.com/franela/goblin"
)

func TestCommon(t *testing.T) {
	PrepareTests()
	g := goblin.Goblin(t)

	tape := NewTape()

	g.Describe("Common", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
		})

		g.It("Should pong", func() {
			w := tape.Get("/api/v1/ping")
			g.Assert(w.Code).Equal(http.StatusOK)
			g.Assert(w.Body.String()).Equal("pong")

		})

		g.It("Too late is too late", func() {

			now := NowUTC()
			before := now.Add(-time.Hour)
			after := now.Add(time.Hour)

			g.Assert(OverTime(before)).Equal(true) // is over time
			g.Assert(OverTime(after)).Equal(false) // is ok

			// is public
			g.Assert(PublicYet(after)).Equal(false)
			g.Assert(PublicYet(before)).Equal(true)

		})

		g.AfterEach(func() {
			tape.AfterEach()
		})
	})

}
