// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2020-present InfoMark.org
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

package bytefmt

import (
	"testing"

	"github.com/franela/goblin"
	"gopkg.in/yaml.v2"
)

func TestConfiguration(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Configuration", func() {

		g.It("Should produce correct strings", func() {

			g.Assert(ToString(0 * BYTE)).Equal("0b")
			g.Assert(ToString(1 * BYTE)).Equal("1b")
			g.Assert(ToString(1 * KILOBYTE)).Equal("1kb")
			g.Assert(ToString(1 * MEGABYTE)).Equal("1mb")
			g.Assert(ToString(1 * GIGABYTE)).Equal("1gb")
			g.Assert(ToString(1 * TERABYTE)).Equal("1tb")
			g.Assert(ToString(1 * PETABYTE)).Equal("1pb")
			g.Assert(ToString(1 * EXABYTE)).Equal("1eb")

		})

		g.It("Should produce read from strings", func() {
			var value ByteSize
			var err error
			value, err = FromString("0b")
			g.Assert(value).Equal(ByteSize(0 * BYTE))
			g.Assert(err).Equal(nil)
			value, err = FromString("1b")
			g.Assert(value).Equal(ByteSize(1 * BYTE))
			g.Assert(err).Equal(nil)
			value, err = FromString("1kb")
			g.Assert(value).Equal(ByteSize(1 * KILOBYTE))
			g.Assert(err).Equal(nil)
			value, err = FromString("1mb")
			g.Assert(value).Equal(ByteSize(1 * MEGABYTE))
			g.Assert(err).Equal(nil)
			value, err = FromString("1gb")
			g.Assert(value).Equal(ByteSize(1 * GIGABYTE))
			g.Assert(err).Equal(nil)
			value, err = FromString("1tb")
			g.Assert(value).Equal(ByteSize(1 * TERABYTE))
			g.Assert(err).Equal(nil)
			value, err = FromString("1pb")
			g.Assert(value).Equal(ByteSize(1 * PETABYTE))
			g.Assert(err).Equal(nil)
			value, err = FromString("1eb")
			g.Assert(value).Equal(ByteSize(1 * EXABYTE))
			g.Assert(err).Equal(nil)

		})

		type SimpleSizeStruct struct {
			Sizes []ByteSize
		}

		g.It("Should unmarshal", func() {

			s := &SimpleSizeStruct{}

			y := `sizes:
    - 0b
    - 1b
    - 1kb
    - 1mb
    - 1gb
    - 1tb
    - 1pb
    - 1eb
    - 152mb
`
			err := yaml.Unmarshal([]byte(y), &s)
			g.Assert(err).Equal(nil)

			g.Assert(s.Sizes[0]).Equal((ByteSize(0 * BYTE)))
			g.Assert(s.Sizes[1]).Equal((ByteSize(1 * BYTE)))
			g.Assert(s.Sizes[2]).Equal((ByteSize(1 * KILOBYTE)))
			g.Assert(s.Sizes[3]).Equal((ByteSize(1 * MEGABYTE)))
			g.Assert(s.Sizes[4]).Equal((ByteSize(1 * GIGABYTE)))
			g.Assert(s.Sizes[5]).Equal((ByteSize(1 * TERABYTE)))
			g.Assert(s.Sizes[6]).Equal((ByteSize(1 * PETABYTE)))
			g.Assert(s.Sizes[7]).Equal((ByteSize(1 * EXABYTE)))
			g.Assert(s.Sizes[8]).Equal((ByteSize(152 * MEGABYTE)))
		})

	})

}
