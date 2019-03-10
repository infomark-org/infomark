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

type Tape struct {
  Router *chi.Mux
}

func NewTape() *Tape {
  return &Tape{}
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
func CreateFileRequestBody(path, contentType string) (*bytes.Buffer, string, error) {
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

  err = writer.Close()
  if err != nil {
    return nil, "", err
  }

  return body, writer.FormDataContentType(), nil

}

func BuildDataRequest(method, url string, data map[string]interface{}) *http.Request {

  var payload_json *bytes.Buffer

  if data != nil {
    dat, err := json.Marshal(data)
    if err != nil {
      panic(err)
    }
    payload_json = bytes.NewBuffer(dat)
  } else {
    payload_json = nil
  }

  r, err := http.NewRequest(method, url, payload_json)
  if err != nil {
    panic(err)
  }

  r.Header.Set("Content-Type", "application/json")

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

func (t *Tape) PlayRequest(r *http.Request) *httptest.ResponseRecorder {
  w := httptest.NewRecorder()
  t.Router.ServeHTTP(w, r)
  return w
}

func ToH(z interface{}) map[string]interface{} {
  data, _ := json.Marshal(z)
  var msgMapTemplate interface{}
  _ = json.Unmarshal(data, &msgMapTemplate)
  return msgMapTemplate.(map[string]interface{})
}

func (t *Tape) Get(url string) *httptest.ResponseRecorder {
  h := make(map[string]interface{})
  r := BuildDataRequest("GET", url, h)
  return t.PlayRequest(r)
}

func (t *Tape) Post(url string, data map[string]interface{}) *httptest.ResponseRecorder {
  r := BuildDataRequest("POST", url, data)
  return t.PlayRequest(r)
}

func (t *Tape) Put(url string, data map[string]interface{}) *httptest.ResponseRecorder {
  r := BuildDataRequest("PUT", url, data)
  return t.PlayRequest(r)
}

func (t *Tape) Patch(url string, data map[string]interface{}) *httptest.ResponseRecorder {
  r := BuildDataRequest("PATCH", url, data)
  return t.PlayRequest(r)
}

func (t *Tape) Delete(url string) *httptest.ResponseRecorder {
  h := make(map[string]interface{})
  r := BuildDataRequest("DELETE", url, h)
  return t.PlayRequest(r)
}

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
