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
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/infomark-org/infomark/auth/authenticate"
	"github.com/infomark-org/infomark/model"
	"github.com/infomark-org/infomark/symbol"
	null "gopkg.in/guregu/null.v3"
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
		return
	}
	team, err := rs.Stores.Team.GetTeamMembersOfUser(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	// add mail addresses if team is confirmed
	if teamID.Valid {
		isConfirmed, err := rs.Stores.Team.Confirmed(teamID.Int64, course.ID)
		if err != nil {
			render.Render(w, r, ErrRender(err))
			return
		}
		if isConfirmed.Bool {
			// render JSON response
			if err = render.Render(w, r, rs.newTeamResponseWithMails(teamID, accessClaims.LoginID, team.Members, team.Mails)); err != nil {
				render.Render(w, r, ErrRender(err))
				return
			}
			return
		}
	}
	// otherwise just return team members without emails
	// render JSON response
	if err = render.Render(w, r, rs.newTeamResponse(teamID, accessClaims.LoginID, team.Members)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
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
	if err == sql.ErrNoRows {
		render.RenderList(w, r, []render.Renderer{})
		return
	}
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	// Get other Users in group without a team
	noTeams, err := rs.Stores.Team.GetOtherUnaryTeamsInGroup(accessClaims.LoginID, groupEnrollment.GroupID)
	if err == sql.ErrNoRows {
		render.RenderList(w, r, []render.Renderer{})
		return
	}
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	// get all teams in group
	teams, err := rs.Stores.Team.GetAllInGroup(groupEnrollment.GroupID)
	if err == sql.ErrNoRows {
		render.RenderList(w, r, []render.Renderer{})
		return
	}
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	// combine unary teams with not complete teams
	teams = append(noTeams, teams...)

	// render the groups that are not maxed out already
	list := []render.Renderer{}
	for k := range teams {
		teamRecord := teams[k]
		if len(teamRecord.Members) < course.MaxTeamSize {
			// incomplete team
			var teamResponse = rs.newTeamResponse(teamRecord.ID, teamRecord.UserID, teamRecord.Members)
			list = append(list, teamResponse)
		}
	}

	// render JSON response
	if err = render.RenderList(w, r, list); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

}

// TeamConfirmedHandler is public endpoint for
// URL: /courses/{course_id}/team/{team_id}/confirmed
// METHOD: get
// TAG: team
// RESPONSE: 200,BoolResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  Did all users in the team confirm the team
func (rs *TeamResource) TeamConfirmedHandler(w http.ResponseWriter, r *http.Request) {
	// get current course
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	team := r.Context().Value(symbol.CtxKeyTeam).(*model.Team)
	isConfirmed, err := rs.Stores.Team.Confirmed(team.ID, course.ID)

	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	// render JSON response
	if err = render.Render(w, r, rs.newBoolResponse(isConfirmed)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// UserConfirmedHandler is public endpoint for
// URL: /courses/{course_id}/team/userconfirmed
// METHOD: get
// TAG: team
// RESPONSE: 200,BoolResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  Did user confirm the team
func (rs *TeamResource) UserConfirmedHandler(w http.ResponseWriter, r *http.Request) {
	// get current course
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)

	isConfirmed, err := rs.Stores.Team.UserConfirmed(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	// render JSON response
	if err = render.Render(w, r, rs.newBoolResponse(isConfirmed)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// UserConfirmedHandlerPut is public endpoint for
// URL: /courses/{course_id}/team/{team_id}/userconfirmed
// METHOD: put
// TAG: team
// RESPONSE: 200,BoolResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  Confirm choice of team for user.
func (rs *TeamResource) ConfirmTeamForUserHandler(w http.ResponseWriter, r *http.Request) {
	// get current course
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)

	// check if user has a team
	teamID, err := rs.Stores.Team.TeamID(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
	if !teamID.Valid {
		render.Render(w, r, ErrBadRequestWithDetails(errors.New("No team to confirm")))
		return
	}

	err = rs.Stores.Team.UserConfirm(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	isConfirmed, err := rs.Stores.Team.UserConfirmed(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	// render JSON response
	if err = render.Render(w, r, rs.newBoolResponse(isConfirmed)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// TeamJoinHandler is public endpoint for
// URL: /courses/{course_id}/team/{team_id}/join
// METHOD: put
// TAG: team
// REQUEST:  TeamJoinRequest
// RESPONSE: 200,TeamResponse
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  updating the enrollment record to join the team
func (rs *TeamResource) TeamJoinHandler(w http.ResponseWriter, r *http.Request) {
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

	data := &TeamJoinRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	// Get current team of user
	teamID, err := rs.Stores.Team.TeamID(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// make sure the team with the id actually exists in the database
	team, err := rs.Stores.Team.Get(data.TeamID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	// check if team is already confirmed
	isConfirmed, err := rs.Stores.Team.Confirmed(team.ID, course.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	if isConfirmed.Bool {
		render.Render(w, r, ErrBadRequestWithDetails(errors.New("The team is already confirmed.")))
		return
	}

	// check whether team is full
	teamRecord, err := rs.Stores.Team.GetTeamMembers(team.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	if len(teamRecord.Members) >= course.MaxTeamSize {
		render.Render(w, r, ErrBadRequestWithDetails(errors.New("Team is already full")))
		return
	}
	// Set confirmed to false for all existing members in team
	rs.Stores.Team.UnconfirmMembers(team.ID)
	// join the team (confirmed choice)
	err = rs.Stores.Team.UpdateTeam(accessClaims.LoginID, course.ID, null.IntFrom(team.ID), true)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// collect all team members for response
	teamMembers, err := rs.Stores.Team.GetTeamMembers(team.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	if teamID.Valid {
		// delete team if only one person is remaining
		teamMembersOld, err := rs.Stores.Team.GetTeamMembers(teamID.Int64)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}
		if len(teamMembersOld.Members) < 2 {
			rs.Stores.Team.Delete(teamID.Int64)
			if err != nil {
				render.Render(w, r, ErrInternalServerErrorWithDetails(err))
				return
			}
		}
	}

	// render JSON response
	if err = render.Render(w, r, rs.newTeamResponse(null.NewInt(team.ID, true), accessClaims.LoginID, teamMembers.Members)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// TeamFormHandler is public endpoint for
// URL: /courses/{course_id}/team/{user_id}/form
// METHOD: post
// TAG: team
// REQUEST:  TeamFormRequest
// RESPONSE: 200,TeamResponse
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  Create new team and updating the enrollment records to join the team
func (rs *TeamResource) TeamFormHandler(w http.ResponseWriter, r *http.Request) {
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

	data := &TeamFormRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	// Make sure the other user exists
	user, err := rs.Stores.User.Get(data.UserID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	// Make sure the other user is not already in a team
	maybeTeamID, err := rs.Stores.Team.TeamID(user.ID, course.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	if maybeTeamID.Valid {
		render.Render(w, r, ErrBadRequestWithDetails(errors.New("User is already in another Team.")))
		return
	}

	// Get current team of user
	teamID, err := rs.Stores.Team.TeamID(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// Create Team
	team, err := rs.Stores.Team.Create()
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	// Add the two users to the team
	// join the team and confirmed choice
	err = rs.Stores.Team.UpdateTeam(accessClaims.LoginID, course.ID, null.IntFrom(team.ID), true)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	// Add other user to the team but without confirming the choice
	err = rs.Stores.Team.UpdateTeam(user.ID, course.ID, null.IntFrom(team.ID), false)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// collect all team members for response
	teamMembers, err := rs.Stores.Team.GetTeamMembers(team.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	if teamID.Valid {
		// delete team if only one person is remaining
		teamMembersOld, err := rs.Stores.Team.GetTeamMembers(teamID.Int64)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}
		if len(teamMembersOld.Members) < 2 {
			rs.Stores.Team.Delete(teamID.Int64)
			if err != nil {
				render.Render(w, r, ErrInternalServerErrorWithDetails(err))
				return
			}
		}
	}

	// render JSON response
	if err = render.Render(w, r, rs.newTeamResponse(null.NewInt(team.ID, true), accessClaims.LoginID, teamMembers.Members)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// TeamLeaveHandler is public endpoint for
// URL: /courses/{course_id}/team/leave
// METHOD: put
// TAG: team
// REQUEST:  -
// RESPONSE: 200,TeamResponse
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  Removes user from its current team
func (rs *TeamResource) TeamLeaveHandler(w http.ResponseWriter, r *http.Request) {
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

	// Get current team of user
	teamID, err := rs.Stores.Team.TeamID(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	if !teamID.Valid {
		render.Render(w, r, ErrBadRequestWithDetails(errors.New("User has no team to leave")))
		return
	}

	// Set confirmed to false for all members in team
	rs.Stores.Team.UnconfirmMembers(teamID.Int64)

	// leave the team
	err = rs.Stores.Team.UpdateTeam(accessClaims.LoginID, course.ID, null.NewInt(0, false), false)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// delete team if only one person is remaining
	teamMembers, err := rs.Stores.Team.GetTeamMembers(teamID.Int64)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	if len(teamMembers.Members) < 2 {
		rs.Stores.Team.Delete(teamID.Int64)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}
	}

	// render JSON response
	if err = render.Render(w, r,
		rs.newTeamResponse(null.NewInt(0, false), accessClaims.LoginID, []string{})); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// .............................................................................
// Context middleware is used to load a Team object from
// the URL parameter `teamID` passed through as the request. In case
// the Team could not be found, we stop here and return a 404.
// We do NOT check whether the user is authorized to get this team.
func (rs *TeamResource) Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var teamID int64
		var err error

		// try to get id from URL
		if teamID, err = strconv.ParseInt(chi.URLParam(r, "team_id"), 10, 64); err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// find specific team in database
		team, err := rs.Stores.Team.Get(teamID)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// serve next
		ctx := context.WithValue(r.Context(), symbol.CtxKeyTeam, team)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
