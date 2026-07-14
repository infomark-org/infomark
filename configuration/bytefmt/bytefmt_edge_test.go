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
)

func TestByteSizeEdgeCases(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("FromString rejects invalid input", func() {

		// The empty string has no unit and must be rejected.
		g.It("Should reject the empty string", func() {
			_, err := FromString("")
			g.Assert(err == nil).IsFalse()
		})

		// A bare number without a unit must be rejected because the parser
		// requires a trailing unit letter.
		g.It("Should reject a number without a unit", func() {
			_, err := FromString("123")
			g.Assert(err == nil).IsFalse()
		})

		// An unknown unit suffix must be rejected.
		g.It("Should reject an unknown unit suffix", func() {
			_, err := FromString("1xb")
			g.Assert(err == nil).IsFalse()
		})

		// A negative quantity is not a valid byte count and must be rejected.
		g.It("Should reject a negative quantity", func() {
			_, err := FromString("-1kb")
			g.Assert(err == nil).IsFalse()
		})

		// A missing numeric part must be rejected.
		g.It("Should reject a missing numeric part", func() {
			_, err := FromString("kb")
			g.Assert(err == nil).IsFalse()
		})
	})

	g.Describe("FromString accepts well-formed input", func() {

		// Surrounding whitespace and upper case must be tolerated because the
		// parser trims and lower-cases the input.
		g.It("Should tolerate whitespace and upper case", func() {
			value, err := FromString("  1KB  ")
			g.Assert(err).Equal(nil)
			g.Assert(value).Equal(ByteSize(1 * Kilobyte))
		})

		// Fractional values must scale by the unit, using the binary basis.
		g.It("Should parse a fractional value", func() {
			value, err := FromString("1.5kb")
			g.Assert(err).Equal(nil)
			g.Assert(value).Equal(ByteSize(1536))
		})

		// Zero bytes is a valid quantity.
		g.It("Should parse a zero quantity", func() {
			value, err := FromString("0mb")
			g.Assert(err).Equal(nil)
			g.Assert(value).Equal(ByteSize(0))
		})
	})

	g.Describe("ToString", func() {

		// Zero must render as the canonical "0b".
		g.It("Should render zero as 0b", func() {
			g.Assert(ToString(ByteSize(0))).Equal("0b")
		})

		// A value that is a fractional multiple of a unit keeps one decimal.
		g.It("Should keep a single decimal for fractional multiples", func() {
			g.Assert(ToString(ByteSize(1536))).Equal("1.5kb")
		})

		// The largest unit that yields a value >= 1 must be chosen.
		g.It("Should choose the largest fitting unit", func() {
			g.Assert(ToString(ByteSize(2 * Megabyte))).Equal("2mb")
		})
	})

	g.Describe("ToString and FromString round-trip", func() {

		// For any exact unit multiple, formatting then parsing must return the
		// original byte count. This is the central serialization invariant.
		g.It("Should round-trip exact unit multiples", func() {
			exactMultiples := []ByteSize{
				ByteSize(0),
				ByteSize(1 * Byte),
				ByteSize(512 * Byte),
				ByteSize(1 * Kilobyte),
				ByteSize(4 * Kilobyte),
				ByteSize(2 * Megabyte),
				ByteSize(3 * Gigabyte),
				ByteSize(1 * Terabyte),
			}
			for _, original := range exactMultiples {
				rendered := ToString(original)
				parsed, err := FromString(rendered)
				g.Assert(err).Equal(nil)
				g.Assert(parsed).Equal(original)
			}
		})
	})
}
