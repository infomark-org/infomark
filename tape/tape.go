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

// TODO consider to export as single library

package tape

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
)

// Tape can send requests to router endpoints. This is used by unit tests to
// check the endpoints
type Tape struct {
	Router *chi.Mux
}

// NewTape creates a new Tape
func NewTape() *Tape {
	return &Tape{}
}

type RequestModifier interface {
	Modify(r *http.Request)
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// ugly!! but see https://github.com/golang/go/issues/16425
func createFormFile(w *multipart.Writer, fieldname, filename string, contentType string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldname), escapeQuotes(filename)))
	h.Set("Content-Type", contentType)
	return w.CreatePart(h)
}

// CreateFileRequestBody create a multi-part form data. We assume all endpoints
// handling files are receicing a single file with form name "file_data"
func CreateFileRequestBody(path, contentType string, params map[string]string) (*bytes.Buffer, string, error) {
	// open file on disk
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	// create body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := createFormFile(writer, "file_data", filepath.Base(path), contentType)
	if err != nil {
		return nil, "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, "", err
	}

	for key, val := range params {
		err = writer.WriteField(key, val)
		if err != nil {
			return nil, "", err
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil

}

// BuildDataRequest creates a request
func BuildDataRequest(method, url string, data map[string]interface{}) *http.Request {

	var payloadJson *bytes.Buffer

	if data != nil {
		dat, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		payloadJson = bytes.NewBuffer(dat)
	} else {
		payloadJson = nil
	}

	r, err := http.NewRequest(method, url, payloadJson)
	if err != nil {
		panic(err)
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Add("X-Forwarded-For", "1.2.3.4")
	r.Header.Set("User-Agent", "Test-Agent")

	return r
}

// Play will send a request without any request body (like GET)
// func (t *Tape) Play(method, url string) *httptest.ResponseRecorder {
//   h := make(map[string]interface{})
//   r := BuildDataRequest(method, url, h)
//   return t.PlayRequest(r)
// }

// // PlayData will send a request with given data in body
// func (t *Tape) PlayData(method, url string, data map[string]interface{}) *httptest.ResponseRecorder {
//   r := BuildDataRequest(method, url, data)
//   return t.PlayRequest(r)
// }

// PlayRequest will send the request to the router and fetch the response
func (t *Tape) PlayRequest(r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	t.Router.ServeHTTP(w, r)
	return w
}

// ToH is a convenience wrapper create a json for any object
func ToH(z interface{}) map[string]interface{} {
	data, _ := json.Marshal(z)
	var msgMapTemplate interface{}
	_ = json.Unmarshal(data, &msgMapTemplate)
	return msgMapTemplate.(map[string]interface{})
}

// ToH is a convenience wrapper create a json for any object
func (t *Tape) ToH(z interface{}) map[string]interface{} {
	data, _ := json.Marshal(z)
	var msgMapTemplate interface{}
	_ = json.Unmarshal(data, &msgMapTemplate)
	return msgMapTemplate.(map[string]interface{})
}

// Get creates, sends a GET request and fetches the response
func (t *Tape) Get(url string, modifiers ...RequestModifier) *httptest.ResponseRecorder {
	h := make(map[string]interface{})
	r := BuildDataRequest("GET", url, h)

	for _, modifier := range modifiers {
		modifier.Modify(r)
	}

	return t.PlayRequest(r)
}

// Post creates, sends a POST request and fetches the response
func (t *Tape) Post(url string, data map[string]interface{}, modifiers ...RequestModifier) *httptest.ResponseRecorder {
	r := BuildDataRequest("POST", url, data)

	for _, modifier := range modifiers {
		modifier.Modify(r)
	}

	return t.PlayRequest(r)
}

// Put creates, sends a PUT request and fetches the response
func (t *Tape) Put(url string, data map[string]interface{}, modifiers ...RequestModifier) *httptest.ResponseRecorder {
	r := BuildDataRequest("PUT", url, data)

	for _, modifier := range modifiers {
		modifier.Modify(r)
	}

	return t.PlayRequest(r)
}

// Patch creates, sends a PATCH request and fetches the response
func (t *Tape) Patch(url string, data map[string]interface{}, modifiers ...RequestModifier) *httptest.ResponseRecorder {
	r := BuildDataRequest("PATCH", url, data)

	for _, modifier := range modifiers {
		modifier.Modify(r)
	}

	return t.PlayRequest(r)
}

// Delete creates, sends a DELETE request and fetches the response
func (t *Tape) Delete(url string, modifiers ...RequestModifier) *httptest.ResponseRecorder {
	h := make(map[string]interface{})
	r := BuildDataRequest("DELETE", url, h)

	for _, modifier := range modifiers {
		modifier.Modify(r)
	}

	return t.PlayRequest(r)
}

// FormatRequest pretty-formats a request and returns it as a string
func (t *Tape) FormatRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}

func (t *Tape) Upload(url string, filename string, contentType string, modifiers ...RequestModifier) (*httptest.ResponseRecorder, error) {
	return t.UploadWithParameters(url, filename, contentType, map[string]string{}, modifiers...)
}

func (t *Tape) UploadWithParameters(url string, filename string, contentType string, params map[string]string, modifiers ...RequestModifier) (*httptest.ResponseRecorder, error) {

	body, ct, err := CreateFileRequestBody(filename, contentType, params)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", ct)
	r.Header.Add("X-Forwarded-For", "1.2.3.4")
	r.Header.Set("User-Agent", "Test-Agent")

	for _, modifier := range modifiers {
		modifier.Modify(r)
	}

	return t.PlayRequest(r), nil
}
