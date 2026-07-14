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

package helper

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/franela/goblin"
)

// urlParameterDefault values below are arbitrary fallbacks used to confirm that
// the *FromURL helpers return the caller-supplied default when a parameter is
// absent or malformed.
const (
	stringParameterDefault = "fallback"
	intParameterDefault    = 7
	int64ParameterDefault  = int64(11)
)

func TestHelperBehavior(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("StringArrayToIntArray", func() {

		// A slice containing a non-integer element must fail and return no slice,
		// so callers cannot accidentally use a partially converted result.
		g.It("Should reject a non-integer element", func() {
			result, err := StringArrayToIntArray([]string{"1", "notanumber", "3"})
			g.Assert(err == nil).IsFalse()
			g.Assert(result == nil).IsTrue()
		})

		// An empty input must yield an empty, non-nil slice with no error.
		g.It("Should convert an empty slice", func() {
			result, err := StringArrayToIntArray([]string{})
			g.Assert(err).Equal(nil)
			g.Assert(len(result)).Equal(0)
		})

		// Negative numbers are valid integers and must be preserved.
		g.It("Should convert negative numbers", func() {
			result, err := StringArrayToIntArray([]string{"-5", "0", "5"})
			g.Assert(err).Equal(nil)
			g.Assert(result[0]).Equal(-5)
			g.Assert(result[1]).Equal(0)
			g.Assert(result[2]).Equal(5)
		})
	})

	g.Describe("StringArrayFromURL", func() {

		// A comma separated parameter must split into its elements.
		g.It("Should split a comma separated parameter", func() {
			request := httptest.NewRequest("GET", "/api/?tags=foo,bar,baz", nil)
			result := StringArrayFromURL(request, "tags", []string{"default"})
			g.Assert(len(result)).Equal(3)
			g.Assert(result[0]).Equal("foo")
			g.Assert(result[2]).Equal("baz")
		})

		// A missing parameter must fall back to the supplied default.
		g.It("Should return the default when the parameter is absent", func() {
			request := httptest.NewRequest("GET", "/api/", nil)
			result := StringArrayFromURL(request, "tags", []string{"default"})
			g.Assert(len(result)).Equal(1)
			g.Assert(result[0]).Equal("default")
		})
	})

	g.Describe("StringFromURL", func() {

		g.It("Should read a present parameter", func() {
			request := httptest.NewRequest("GET", "/api/?name=value", nil)
			g.Assert(StringFromURL(request, "name", stringParameterDefault)).Equal("value")
		})

		g.It("Should fall back for an absent parameter", func() {
			request := httptest.NewRequest("GET", "/api/", nil)
			g.Assert(StringFromURL(request, "name", stringParameterDefault)).Equal(stringParameterDefault)
		})
	})

	g.Describe("IntFromURL", func() {

		g.It("Should read a valid integer parameter", func() {
			request := httptest.NewRequest("GET", "/api/?count=42", nil)
			g.Assert(IntFromURL(request, "count", intParameterDefault)).Equal(42)
		})

		// A malformed integer must fall back to the default rather than error.
		g.It("Should fall back for a malformed integer", func() {
			request := httptest.NewRequest("GET", "/api/?count=notint", nil)
			g.Assert(IntFromURL(request, "count", intParameterDefault)).Equal(intParameterDefault)
		})

		g.It("Should fall back for an absent integer", func() {
			request := httptest.NewRequest("GET", "/api/", nil)
			g.Assert(IntFromURL(request, "count", intParameterDefault)).Equal(intParameterDefault)
		})
	})

	g.Describe("Int64FromURL", func() {

		g.It("Should read a valid int64 parameter", func() {
			request := httptest.NewRequest("GET", "/api/?count=42", nil)
			g.Assert(Int64FromURL(request, "count", int64ParameterDefault)).Equal(int64(42))
		})

		g.It("Should fall back for a malformed int64", func() {
			request := httptest.NewRequest("GET", "/api/?count=notint", nil)
			g.Assert(Int64FromURL(request, "count", int64ParameterDefault)).Equal(int64ParameterDefault)
		})
	})

	g.Describe("ToH", func() {

		// ToH must expose struct fields under their JSON keys.
		g.It("Should convert a struct into a keyed map", func() {
			type sample struct {
				Title string `json:"title"`
			}
			result := ToH(sample{Title: "hello"})
			g.Assert(result["title"]).Equal("hello")
		})
	})

	g.Describe("Time", func() {

		// Time must strip sub-second precision so database round-trips compare
		// equal. The returned value must carry no nanoseconds.
		g.It("Should drop the nanosecond component", func() {
			withNanoseconds := time.Date(2020, time.January, 2, 3, 4, 5, 123456789, time.UTC)
			truncated := Time(withNanoseconds)
			g.Assert(truncated.Nanosecond()).Equal(0)
		})
	})

	g.Describe("File magic-number detection", func() {

		// Each detector must accept its own signature and reject a foreign one.
		g.It("Should detect a zip signature", func() {
			g.Assert(IsZipFile([]byte{0x50, 0x4B, 0x03, 0x04})).IsTrue()
			g.Assert(IsZipFile([]byte{0x25, 0x50, 0x44, 0x46})).IsFalse()
		})

		g.It("Should detect a pdf signature", func() {
			g.Assert(IsPdfFile([]byte{0x25, 0x50, 0x44, 0x46})).IsTrue()
			g.Assert(IsPdfFile([]byte{0x50, 0x4B, 0x03, 0x04})).IsFalse()
		})

		g.It("Should detect a jpeg signature", func() {
			g.Assert(IsJpegFile([]byte{0xFF, 0xD8, 0xFF})).IsTrue()
			g.Assert(IsJpegFile([]byte{0x89, 0x50, 0x4E, 0x47})).IsFalse()
		})

		g.It("Should detect a png signature", func() {
			g.Assert(IsPngFile([]byte{0x89, 0x50, 0x4E, 0x47})).IsTrue()
			g.Assert(IsPngFile([]byte{0xFF, 0xD8, 0xFF})).IsFalse()
		})

		// A buffer shorter than the signature must be rejected without panicking
		// on an out-of-range index.
		g.It("Should reject buffers that are too short", func() {
			g.Assert(IsZipFile([]byte{0x50, 0x4B})).IsFalse()
			g.Assert(IsPdfFile([]byte{0x25})).IsFalse()
			g.Assert(IsJpegFile([]byte{0xFF})).IsFalse()
			g.Assert(IsPngFile([]byte{})).IsFalse()
		})
	})

	g.Describe("FileExists, FileTouch and FileDelete", func() {

		// A freshly created file must be reported as existing, and after deletion
		// it must be reported as absent, exercising the full lifecycle.
		g.It("Should track the lifecycle of a file", func() {
			temporaryDirectory, err := os.MkdirTemp("", "infomark-helper-test")
			g.Assert(err).Equal(nil)
			defer os.RemoveAll(temporaryDirectory)

			filePath := filepath.Join(temporaryDirectory, "created.bin")
			g.Assert(FileExists(filePath)).IsFalse()

			err = FileTouch(filePath)
			g.Assert(err).Equal(nil)
			g.Assert(FileExists(filePath)).IsTrue()

			err = FileDelete(filePath)
			g.Assert(err).Equal(nil)
			g.Assert(FileExists(filePath)).IsFalse()
		})
	})
}
