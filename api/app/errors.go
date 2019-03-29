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

  "github.com/go-chi/render"
  validation "github.com/go-ozzo/ozzo-validation"
)

// ErrResponse renderer type for handling all sorts of errors.
type ErrResponse struct {
  Err            error `json:"-"` // low-level runtime error
  HTTPStatusCode int   `json:"-"` // http response status code

  StatusText       string            `json:"status"`           // user-level status message
  AppCode          int64             `json:"code,omitempty"`   // application-specific error code
  ErrorText        string            `json:"error,omitempty"`  // application-level error message, for debugging
  ValidationErrors validation.Errors `json:"errors,omitempty"` // user level model validation errors
}

// Render sets the application-specific error code in AppCode.
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
  render.Status(r, e.HTTPStatusCode)
  return nil
}

// ErrRender returns status 422 Unprocessable Entity rendering response error.
func ErrRender(err error) render.Renderer {
  return &ErrResponse{
    Err:            err,
    HTTPStatusCode: http.StatusUnprocessableEntity,
    StatusText:     http.StatusText(http.StatusUnprocessableEntity),
    ErrorText:      err.Error(),
  }
}

func ErrBadRequestWithDetails(err error) *ErrResponse {
  return &ErrResponse{
    Err:            err,
    HTTPStatusCode: http.StatusBadRequest,
    StatusText:     http.StatusText(http.StatusBadRequest),
    ErrorText:      err.Error(),
  }
}

func ErrInternalServerErrorWithDetails(err error) *ErrResponse {
  return &ErrResponse{
    Err:            err,
    HTTPStatusCode: http.StatusInternalServerError,
    StatusText:     http.StatusText(http.StatusInternalServerError),
    ErrorText:      err.Error(),
  }
}

func ErrTimeoutWithDetails(err error) *ErrResponse {
  return &ErrResponse{
    Err:            err,
    HTTPStatusCode: http.StatusGatewayTimeout,
    StatusText:     http.StatusText(http.StatusGatewayTimeout),
    ErrorText:      err.Error(),
  }
}

// see https://stackoverflow.com/a/50143519/7443104
var (
  // ErrBadRequest returns status 400 Bad Request for malformed request body.
  ErrBadRequest = &ErrResponse{HTTPStatusCode: http.StatusBadRequest, StatusText: http.StatusText(http.StatusBadRequest)}

  // ErrUnauthorized returns 401 Unauthorized.
  // e.g. "User has not logged-in"
  ErrUnauthenticated = &ErrResponse{HTTPStatusCode: http.StatusUnauthorized, StatusText: http.StatusText(http.StatusUnauthorized)}

  // ErrForbidden returns status 403 Forbidden for unauthorized request.
  // e.g. "User doesn't have enough privilege"
  ErrUnauthorized = &ErrResponse{HTTPStatusCode: http.StatusForbidden, StatusText: http.StatusText(http.StatusForbidden)}

  // ErrNotFound returns status 404 Not Found for invalid resource request.
  ErrNotFound = &ErrResponse{HTTPStatusCode: http.StatusNotFound, StatusText: http.StatusText(http.StatusNotFound)}

  // ErrInternalServerError returns status 500 Internal Server Error.
  ErrInternalServerError = &ErrResponse{HTTPStatusCode: http.StatusInternalServerError, StatusText: http.StatusText(http.StatusInternalServerError)}
)

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
