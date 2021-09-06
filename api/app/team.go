// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019 ComputerGraphics Tuebingen
//               2020-present InfoMark.org
// Authors: Raphael Braun
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
	"github.com/infomark-org/infomark/auth/authenticate"
	"github.com/infomark-org/infomark/model"
	"github.com/infomark-org/infomark/symbol"
)

// TeamResource specifies team management handler.
type TeamResource struct {
	Stores *Stores
}

// NewTeamResource create and returns a TeamResource.
func NewTeamResource(stores *Stores) *TeamResource {
	return &TeamResource{
		Stores: stores,
	}
}

// IndexHandler is public endpoint for
// URL: /courses/{course_id}/team
// METHOD: get
// TAG: team
// RESPONSE: 200,TeamResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  Current exercise team
func (rs *TeamResource) IndexTeamHandler(w http.ResponseWriter, r *http.Request) {
	// get current course
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	// get userID
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	// get team of this user
	teamID, err := rs.Stores.Team.TeamID(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
	}
	team, err := rs.Stores.Team.GetTeamMembersOfUser(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
	}

	// render JSON response
	if err = render.Render(w, r, rs.newTeamResponse(teamID, accessClaims.LoginID, team.Members)); err != nil {
		render.Render(w, r, ErrRender(err))
	}
}

// IncompleteTeamHandler is public endpoint for
// URL: /courses/{course_id}/teams
// METHOD: get
// TAG: teams
// RESPONSE: 200,TeamListResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  List of incomplete teams in same exercise-group
func (rs *TeamResource) IncompleteTeamsHandler(w http.ResponseWriter, r *http.Request) {
	// get current course
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	// get userID
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	// get group enrollment
	groupEnrollment, err := rs.Stores.Group.GetGroupEnrollmentOfUserInCourse(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
	}

	// Get Users in group without a team
	noTeams, err := rs.Stores.Team.GetUnaryTeamsInGroup(groupEnrollment.GroupID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
	}

	// get all teams in group
	teams, err := rs.Stores.Team.GetAllInGroup(groupEnrollment.GroupID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
	}

	// combine unary teams with not complete teams
	teams = append(noTeams, teams...)

	// render the groups that are not maxed out already
	list := []render.Renderer{}
	for k := range teams {
		teamRecord:= teams[k]
		if (len(teamRecord.Members) < course.MaxTeamSize) {
			// incomplete team
			var teamResponse = rs.newTeamResponse(teamRecord.ID, teamRecord.UserID, teamRecord.Members)
			list = append(list, teamResponse)
		}
	}

	// render JSON response
	if err = render.RenderList(w, r, list); err != nil {
		render.Render(w, r, ErrRender(err))
	}

}
