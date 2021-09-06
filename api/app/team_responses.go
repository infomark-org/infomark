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
	null "gopkg.in/guregu/null.v3"
	"net/http"
)

// TeamResponse is the response payload for team management.
type TeamResponse struct {
	ID      null.Int `json:"id" example:"1"`
	UserID  int64    `json:"user_id" example:"1"`
	Members []string `json:"members" example:"['Jon Doe', 'Jane Doe'}]"`
}

// Render post-processes a TeamResponse.
func (body *TeamResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newTeamResponse creates a response from a course model.
func (rs *TeamResource) newTeamResponse(teamID null.Int, userID int64, members []string) *TeamResponse {
	return &TeamResponse{
		ID:      teamID,
		UserID:  userID,
		Members: members,
	}
}
