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
)

// CommonResource specifies user management handler.
type CommonResource struct {
  Stores *Stores
}

// NewCommonResource create and returns a CommonResource.
func NewCommonResource(stores *Stores) *CommonResource {
  return &CommonResource{
    Stores: stores,
  }
}

// PingHandler is public endpoint for
// URL: /ping
// METHOD: post
// TAG: common
// RESPONSE: 200,pongResponse
// SUMMARY:  heartbeat of backend
func (rs *CommonResource) PingHandler(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("pong"))
}
