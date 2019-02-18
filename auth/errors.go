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

package auth

import (
  "errors"
  "net/http"

  "github.com/go-chi/render"
)

// The list of jwt token errors presented to the end user.
var (
  ErrTokenUnauthorized   = errors.New("token unauthorized")
  ErrTokenExpired        = errors.New("token expired")
  ErrInvalidAccessToken  = errors.New("invalid access token")
  ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// ErrResponse renderer type for handling all sorts of errors.
type ErrResponse struct {
  Err            error `json:"-"` // low-level runtime error
  HTTPStatusCode int   `json:"-"` // http response status code

  StatusText string `json:"status"`          // user-level status message
  AppCode    int64  `json:"code,omitempty"`  // application-specific error code
  ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

// Render sets the application-specific error code in AppCode.
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
  render.Status(r, e.HTTPStatusCode)
  return nil
}

// ErrUnauthenticated renders status 401 Unauthorized with custom error message.
// The request has no credentials at all.
func ErrUnauthenticatedWithDetails(err error) render.Renderer {
  // StatusUnauthorized                  = 401 // RFC 7235, 3.1
  return &ErrResponse{
    Err:            err,
    HTTPStatusCode: http.StatusUnauthorized,
    StatusText:     http.StatusText(http.StatusUnauthorized),
    ErrorText:      err.Error(),
  }
}

// ErrUnauthorized renders status 403 Unauthorized with custom error message.
// The request is issued with credential, but these are invalid or not sufficient
// to gain access to a ressource.
func ErrUnauthorizedWithDetails(err error) render.Renderer {
  // StatusForbidden                     = 403 // RFC 7231, 6.5.3
  return &ErrResponse{
    Err:            err,
    HTTPStatusCode: http.StatusUnauthorized,
    StatusText:     http.StatusText(http.StatusUnauthorized),
    ErrorText:      err.Error(),
  }
}

var (
  ErrUnauthenticated = &ErrResponse{HTTPStatusCode: http.StatusUnauthorized, StatusText: http.StatusText(http.StatusUnauthorized)}

  // ErrForbidden returns status 403 Forbidden for unauthorized request.
  // e.g. "User doesn't have enough privilege"
  ErrUnauthorized = &ErrResponse{HTTPStatusCode: http.StatusForbidden, StatusText: http.StatusText(http.StatusForbidden)}
)
