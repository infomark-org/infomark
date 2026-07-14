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

package tape

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/franela/goblin"
)

func TestTapeHelpers(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("BuildDataRequest", func() {

		// A data request must carry the JSON content type and the fixed headers
		// the test harness relies on, and it must target the requested method
		// and URL.
		g.It("Should set method, url and the standard headers", func() {
			request := BuildDataRequest("POST", "http://example.org/api/v1/ping", map[string]interface{}{"key": "value"})
			g.Assert(request.Method).Equal("POST")
			g.Assert(request.URL.Path).Equal("/api/v1/ping")
			g.Assert(request.Header.Get("Content-Type")).Equal("application/json")
			g.Assert(request.Header.Get("X-Forwarded-For")).Equal("1.2.3.4")
			g.Assert(request.Header.Get("User-Agent")).Equal("Test-Agent")
			// A non-nil data map must produce a request body.
			g.Assert(request.Body == nil).IsFalse()
		})

		// A nil data map means "no body". The builder now leaves the io.Reader
		// as a nil interface value, so http.NewRequest yields a request without
		// a body instead of panicking on a typed-nil *bytes.Buffer.
		g.It("Should yield a bodyless request when given a nil data map", func() {
			request := BuildDataRequest("GET", "http://example.org/api/v1/ping", nil)
			g.Assert(request.Method).Equal("GET")
			g.Assert(request.URL.Path).Equal("/api/v1/ping")
			// A nil data map must not produce a request body.
			g.Assert(request.Body == nil).IsTrue()
			// The standard headers are still set regardless of the body.
			g.Assert(request.Header.Get("Content-Type")).Equal("application/json")
		})

		// An explicitly empty (non-nil) map is the supported bodyless form and
		// must produce a request without panicking.
		g.It("Should accept an empty data map", func() {
			request := BuildDataRequest("GET", "http://example.org/api/v1/ping", map[string]interface{}{})
			g.Assert(request.Method).Equal("GET")
			g.Assert(request.Body == nil).IsFalse()
		})
	})

	g.Describe("ToH", func() {

		// ToH must round-trip a struct into its JSON object representation so
		// that field values are reachable by their JSON keys.
		g.It("Should convert a struct into a keyed map", func() {
			type sample struct {
				Name  string `json:"name"`
				Count int    `json:"count"`
			}
			result := ToH(sample{Name: "widget", Count: 3})
			g.Assert(result["name"]).Equal("widget")
			// JSON numbers decode into float64, matching encoding/json semantics.
			g.Assert(result["count"]).Equal(float64(3))
		})
	})

	g.Describe("escapeQuotes", func() {

		// Backslashes and double quotes must be escaped so they can be embedded
		// safely inside a quoted multipart header value.
		g.It("Should escape backslashes and double quotes", func() {
			g.Assert(escapeQuotes(`a"b\c`)).Equal(`a\"b\\c`)
			// A string without special characters must be returned unchanged.
			g.Assert(escapeQuotes("plain")).Equal("plain")
		})
	})

	g.Describe("FormatRequest", func() {

		// The formatted request must begin with the request line and expose the
		// host and lower-cased header names, which is what the debug output relies on.
		g.It("Should render the request line, host and headers", func() {
			tape := NewTape()
			request := BuildDataRequest("GET", "http://example.org/api/v1/ping", map[string]interface{}{})
			formatted := tape.FormatRequest(request)
			g.Assert(strings.Contains(formatted, "GET")).IsTrue()
			g.Assert(strings.Contains(formatted, "Host: example.org")).IsTrue()
			// Header names are lower-cased by FormatRequest.
			g.Assert(strings.Contains(formatted, "content-type: application/json")).IsTrue()
			g.Assert(strings.Contains(formatted, "user-agent: Test-Agent")).IsTrue()
		})
	})

	g.Describe("CreateFileRequestBody", func() {

		// Building a multipart body from a real file on disk must succeed and
		// return a multipart content type carrying a boundary, plus a body that
		// embeds the file field and an extra form parameter.
		g.It("Should build a multipart body from a file", func() {
			temporaryDirectory, err := os.MkdirTemp("", "infomark-tape-test")
			g.Assert(err).Equal(nil)
			defer os.RemoveAll(temporaryDirectory)

			filePath := filepath.Join(temporaryDirectory, "payload.txt")
			err = os.WriteFile(filePath, []byte("hello world"), 0644)
			g.Assert(err).Equal(nil)

			body, contentType, err := CreateFileRequestBody(filePath, "text/plain", map[string]string{"extra": "field-value"})
			g.Assert(err).Equal(nil)
			g.Assert(strings.HasPrefix(contentType, "multipart/form-data; boundary=")).IsTrue()

			rendered := body.String()
			// The default upload field name is file_data.
			g.Assert(strings.Contains(rendered, `name="file_data"`)).IsTrue()
			g.Assert(strings.Contains(rendered, "payload.txt")).IsTrue()
			g.Assert(strings.Contains(rendered, "hello world")).IsTrue()
			// The additional parameter must appear as its own form field.
			g.Assert(strings.Contains(rendered, `name="extra"`)).IsTrue()
			g.Assert(strings.Contains(rendered, "field-value")).IsTrue()
		})

		// A missing source file must surface as an error rather than a panic.
		g.It("Should error when the source file is missing", func() {
			_, _, err := CreateFileRequestBody(filepath.Join(os.TempDir(), "no-such-file-infomark.bin"), "text/plain", nil)
			g.Assert(err == nil).IsFalse()
		})
	})
}
