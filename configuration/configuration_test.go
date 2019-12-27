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

package configuration

import (
	"testing"

	"github.com/franela/goblin"
)

func TestConfiguration(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Configuration", func() {

		g.It("Should read configuration", func() {

			config, err := ParseConfiguration("example.yml")
			g.Assert(err).Equal(nil)
			g.Assert(config.Server.HTTP.Port).Equal(3000)
			g.Assert(config.Server.HTTP.Domain).Equal("sub.domain.com")

			g.Assert(config.Server.Debugging.Enabled).Equal(false)
			g.Assert(config.Server.Debugging.UserID).Equal(int64(1))
			g.Assert(config.Server.Debugging.UserIsRoot).Equal(false)
			g.Assert(config.Server.Debugging.LogLevel).Equal("debug")

		})

	})

}
