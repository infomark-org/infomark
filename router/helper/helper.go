// Copyright 2019 ComputerGraphics Tuebingen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ==============================================================================
// Authors: Patrick Wieschollek

package helper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"
)

type Omit *struct{}

// similar to gin.H as a neat wrapper
type H map[string]interface{}

func EmptyHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, H{"response": "empty"})
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}

func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	writeContentType(w, []string{"application/json; charset=utf-8"})
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	w.Write(jsonBytes)
	return nil
}

func WriteText(w http.ResponseWriter, format string) error {
	writeContentType(w, []string{"text/plain; charset=utf-8"})
	io.WriteString(w, format)
	return nil
}

func WriteTextf(w http.ResponseWriter, format string, a ...interface{}) error {
	writeContentType(w, []string{"text/plain; charset=utf-8"})
	io.WriteString(w, fmt.Sprintf(format, a...))
	return nil
}

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

// Example:
// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 	render.WriteJSON(w, render.H{"test": "hi"})
// 	render.WriteText(w, "hi")
// 	http.Redirect(w, r, "/", 301)
// })
//

// StatusContinue           = 100
// StatusSwitchingProtocols = 101

// StatusOK                   = 200
// StatusCreated              = 201
// StatusAccepted             = 202
// StatusNonAuthoritativeInfo = 203
// StatusNoContent            = 204
// StatusResetContent         = 205
// StatusPartialContent       = 206

// StatusMultipleChoices   = 300
// StatusMovedPermanently  = 301
// StatusFound             = 302
// StatusSeeOther          = 303
// StatusNotModified       = 304
// StatusUseProxy          = 305
// StatusTemporaryRedirect = 307

// StatusBadRequest                   = 400
// StatusUnauthorized                 = 401
// StatusPaymentRequired              = 402
// StatusForbidden                    = 403
// StatusNotFound                     = 404
// StatusMethodNotAllowed             = 405
// StatusNotAcceptable                = 406
// StatusProxyAuthRequired            = 407
// StatusRequestTimeout               = 408
// StatusConflict                     = 409
// StatusGone                         = 410
// StatusLengthRequired               = 411
// StatusPreconditionFailed           = 412
// StatusRequestEntityTooLarge        = 413
// StatusRequestURITooLong            = 414
// StatusUnsupportedMediaType         = 415
// StatusRequestedRangeNotSatisfiable = 416
// StatusExpectationFailed            = 417
// StatusTeapot                       = 418

// StatusInternalServerError     = 500
// StatusNotImplemented          = 501
// StatusBadGateway              = 502
// StatusServiceUnavailable      = 503
// StatusGatewayTimeout          = 504
// StatusHTTPVersionNotSupported = 505
