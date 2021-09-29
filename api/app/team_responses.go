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
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/infomark-org/infomark/model"
	null "gopkg.in/guregu/null.v3"
)

// TeamResponse is the response payload for team management.
type TeamResponse struct {
	ID          null.Int `json:"id" example:"1"`
	UserID      int64    `json:"user_id" example:"1"`
	Members     []string `json:"members" example:"['Jon Doe', 'Jane Doe'}]"`
	MemberMails []string `json:"member_mails" example:"['jon.joe@mail.com', 'jane.doe@mail.com'}]"`
}

// Render post-processes a TeamResponse.
func (body *TeamResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newTeamResponse creates a response from a course model.
func (rs *TeamResource) newTeamResponse(teamID null.Int, userID int64, members []string) *TeamResponse {
	return &TeamResponse{
		ID:          teamID,
		UserID:      userID,
		Members:     members,
		MemberMails: nil,
	}
}

// newTeamResponse creates a response from a course model.
func (rs *TeamResource) newTeamResponseWithMails(teamID null.Int, userID int64, members []string, mails []string) *TeamResponse {
	return &TeamResponse{
		ID:          teamID,
		UserID:      userID,
		Members:     members,
		MemberMails: mails,
	}
}

// BoolResponse is the response payload for booleans.
type BoolResponse struct {
	Bool bool `json:"bool" example:"true"`
}

// Render post-processes a TeamResponse.
func (body *BoolResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newBoolResponse creates a response from a course model.
func (rs *TeamResource) newBoolResponse(boolRecord *model.BoolRecord) *BoolResponse {
	return &BoolResponse{
		Bool: boolRecord.Bool,
	}
}

// TeamJoinRequest is the request payload for joining an existing team.
type TeamJoinRequest struct {
	TeamID int64 `json:"team_id" example:"23"`
}

// TeamJoinRequest is the request payload for forming a new team.
type TeamFormRequest struct {
	UserID int64 `json:"user_id" example:"12"`
}

// Bind preprocesses a TeamJoinRequest.
func (body *TeamJoinRequest) Bind(r *http.Request) error {

	if body == nil {
		return errors.New("missing \"team_id\" in data")
	}

	return body.Validate()
}

func (body *TeamJoinRequest) Validate() error {
	return validation.ValidateStruct(body,
		validation.Field(
			&body.TeamID,
			validation.Required,
		))
}

// Bind preprocesses a TeamFormRequest.
func (body *TeamFormRequest) Bind(r *http.Request) error {

	if body == nil {
		return errors.New("missing \"user_id\" in data")
	}

	return body.Validate()
}

func (body *TeamFormRequest) Validate() error {
	return validation.ValidateStruct(body,
		validation.Field(
			&body.UserID,
			validation.Required,
		))
}
