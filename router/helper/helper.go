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
	"net/http"

	"github.com/cgtuebingen/infomark-backend/validation"
	"github.com/go-chi/render"
)

// to omit fields in json structs
// see: https://attilaolah.eu/2014/09/10/json-and-struct-composition-in-go/
type Omit *struct{}

// similar to gin.H as a neat wrapper
type H map[string]interface{}

func EmptyHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, H{"response": "empty"})
}

// func writeContentType(w http.ResponseWriter, value []string) {
// 	header := w.Header()
// 	if val := header["Content-Type"]; len(val) == 0 {
// 		header["Content-Type"] = value
// 	}
// }

func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	// writeContentType(w, []string{"application/json; charset=utf-8"})
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	w.Write(jsonBytes)
	return nil
}

// func WriteText(w http.ResponseWriter, format string) error {
// 	writeContentType(w, []string{"text/plain; charset=utf-8"})
// 	io.WriteString(w, format)
// 	return nil
// }

// func WriteTextf(w http.ResponseWriter, format string, a ...interface{}) error {
// 	writeContentType(w, []string{"text/plain; charset=utf-8"})
// 	io.WriteString(w, fmt.Sprintf(format, a...))
// 	return nil
// }

type ErrResponse struct {
	HTTPStatusCode int    `json:"-"`               // http response status code
	StatusText     string `json:"status"`          // user-level status message
	AppCode        int64  `json:"code,omitempty"`  // application-specific error code
	Err            error  `json:"-"`               // low-level runtime error
	ErrorText      string `json:"error,omitempty"` // application-level error message, for debugging
}

func NewErrResponse(status int, err error) *ErrResponse {
	if err != nil {
		return &ErrResponse{
			Err:            err,
			HTTPStatusCode: status,
			StatusText:     http.StatusText(status),
			ErrorText:      err.Error(),
		}
	} else {
		return &ErrResponse{
			HTTPStatusCode: status,
			StatusText:     http.StatusText(status),
		}
	}

}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func RenderValidation(hints *validation.CheckResponses, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	WriteJSON(w, hints)
}

var ErrNotFoundResponse = NewErrResponse(http.StatusNotFound, nil)

func ErrRenderResponse(err error) render.Renderer {
	return NewErrResponse(http.StatusUnprocessableEntity, err)
}

func ErrDatabaseResponse(err error) render.Renderer {
	return NewErrResponse(http.StatusServiceUnavailable, err)
}

// Example:
// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 	render.WriteJSON(w, render.H{"test": "hi"})
// 	render.WriteText(w, "hi")
// 	http.Redirect(w, r, "/", 301)
// })
//

// StatusContinue                      = 100 // RFC 7231, 6.2.1
// StatusSwitchingProtocols            = 101 // RFC 7231, 6.2.2
// StatusProcessing                    = 102 // RFC 2518, 10.1

// StatusOK                            = 200 // RFC 7231, 6.3.1
// StatusCreated                       = 201 // RFC 7231, 6.3.2
// StatusAccepted                      = 202 // RFC 7231, 6.3.3
// StatusNonAuthoritativeInfo          = 203 // RFC 7231, 6.3.4
// StatusNoContent                     = 204 // RFC 7231, 6.3.5
// StatusResetContent                  = 205 // RFC 7231, 6.3.6
// StatusPartialContent                = 206 // RFC 7233, 4.1
// StatusMultiStatus                   = 207 // RFC 4918, 11.1
// StatusAlreadyReported               = 208 // RFC 5842, 7.1
// StatusIMUsed                        = 226 // RFC 3229, 10.4.1

// StatusMultipleChoices               = 300 // RFC 7231, 6.4.1
// StatusMovedPermanently              = 301 // RFC 7231, 6.4.2
// StatusFound                         = 302 // RFC 7231, 6.4.3
// StatusSeeOther                      = 303 // RFC 7231, 6.4.4
// StatusNotModified                   = 304 // RFC 7232, 4.1
// StatusUseProxy                      = 305 // RFC 7231, 6.4.5
// _                                   = 306 // RFC 7231, 6.4.6 (Unused)
// StatusTemporaryRedirect             = 307 // RFC 7231, 6.4.7
// StatusPermanentRedirect             = 308 // RFC 7538, 3

// StatusBadRequest                    = 400 // RFC 7231, 6.5.1
// StatusUnauthorized                  = 401 // RFC 7235, 3.1
// StatusPaymentRequired               = 402 // RFC 7231, 6.5.2
// StatusForbidden                     = 403 // RFC 7231, 6.5.3
// StatusNotFound                      = 404 // RFC 7231, 6.5.4
// StatusMethodNotAllowed              = 405 // RFC 7231, 6.5.5
// StatusNotAcceptable                 = 406 // RFC 7231, 6.5.6
// StatusProxyAuthRequired             = 407 // RFC 7235, 3.2
// StatusRequestTimeout                = 408 // RFC 7231, 6.5.7
// StatusConflict                      = 409 // RFC 7231, 6.5.8
// StatusGone                          = 410 // RFC 7231, 6.5.9
// StatusLengthRequired                = 411 // RFC 7231, 6.5.10
// StatusPreconditionFailed            = 412 // RFC 7232, 4.2
// StatusRequestEntityTooLarge         = 413 // RFC 7231, 6.5.11
// StatusRequestURITooLong             = 414 // RFC 7231, 6.5.12
// StatusUnsupportedMediaType          = 415 // RFC 7231, 6.5.13
// StatusRequestedRangeNotSatisfiable  = 416 // RFC 7233, 4.4
// StatusExpectationFailed             = 417 // RFC 7231, 6.5.14
// StatusTeapot                        = 418 // RFC 7168, 2.3.3
// StatusMisdirectedRequest            = 421 // RFC 7540, 9.1.2
// StatusUnprocessableEntity           = 422 // RFC 4918, 11.2
// StatusLocked                        = 423 // RFC 4918, 11.3
// StatusFailedDependency              = 424 // RFC 4918, 11.4
// StatusUpgradeRequired               = 426 // RFC 7231, 6.5.15
// StatusPreconditionRequired          = 428 // RFC 6585, 3
// StatusTooManyRequests               = 429 // RFC 6585, 4
// StatusRequestHeaderFieldsTooLarge   = 431 // RFC 6585, 5
// StatusUnavailableForLegalReasons    = 451 // RFC 7725, 3

// StatusInternalServerError           = 500 // RFC 7231, 6.6.1
// StatusNotImplemented                = 501 // RFC 7231, 6.6.2
// StatusBadGateway                    = 502 // RFC 7231, 6.6.3
// StatusServiceUnavailable            = 503 // RFC 7231, 6.6.4
// StatusGatewayTimeout                = 504 // RFC 7231, 6.6.5
// StatusHTTPVersionNotSupported       = 505 // RFC 7231, 6.6.6
// StatusVariantAlsoNegotiates         = 506 // RFC 2295, 8.1
// StatusInsufficientStorage           = 507 // RFC 4918, 11.5
// StatusLoopDetected                  = 508 // RFC 5842, 7.2
// StatusNotExtended                   = 510 // RFC 2774, 7
// StatusNetworkAuthenticationRequired = 511 // RFC 6585, 6
