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

package app

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/infomark-org/infomark-backend/auth/authorize"
	"github.com/infomark-org/infomark-backend/configuration"
	"github.com/infomark-org/infomark-backend/model"
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
// METHOD: get
// TAG: common
// RESPONSE: 200,PongResponse
// SUMMARY:  heartbeat of backend
func (rs *CommonResource) PingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

// VersionHandler is public endpoint for
// URL: /version
// METHOD: get
// TAG: common
// RESPONSE: 200,VersionResponse
// SUMMARY:  all version information
func (rs *CommonResource) VersionHandler(w http.ResponseWriter, r *http.Request) {
	// render JSON reponse
	if err := render.Render(w, r, newVersionResponse()); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusOK)
}

// PrivacyStatementHandler is public endpoint for
// URL: /privacy_statement
// METHOD: get
// TAG: common
// RESPONSE: 200,RawResponse
// SUMMARY:  the privacy statement
func (rs *CommonResource) PrivacyStatementHandler(w http.ResponseWriter, r *http.Request) {

	buf, err := ioutil.ReadFile(fmt.Sprintf("%s/privacy_statement.md", configuration.Configuration.Server.Paths.Common)) // just pass the file name
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	text := string(buf) // convert content to a 'string'

	if err := render.Render(w, r, newRawResponse(text)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

}

// EnsurePrivacyInEnrollments removes some data from the request to ensure that not everyone has access to personal data
func EnsurePrivacyInEnrollments(enrolledUsers []model.UserCourse, givenRole authorize.CourseRole) []model.UserCourse {
	if givenRole == authorize.STUDENT {
		for k := range enrolledUsers {

			if enrolledUsers[k].Role == 0 {
				enrolledUsers[k].Email = ""

			}
		}
	}

	if givenRole != authorize.ADMIN {
		for k := range enrolledUsers {

			enrolledUsers[k].StudentNumber = ""
			enrolledUsers[k].Semester = 0
			enrolledUsers[k].Subject = ""

		}
	}
	return enrolledUsers
}

// PublicYet tests if a given time is now or in the past
func PublicYet(t time.Time) bool {
	return NowUTC().Sub(t) > 0
}

// OverTime tests if the deadline is missed (alias for publicyet)
func OverTime(t time.Time) bool {
	return NowUTC().Sub(t) > 0
}

// NowUTC returns the current server time
func NowUTC() time.Time {
	loc, _ := time.LoadLocation("UTC")
	return time.Now().In(loc)
}
