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

package database

import (
	"github.com/infomark-org/infomark/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"errors"

	null "gopkg.in/guregu/null.v3"
)

type TeamStore struct {
	db *sqlx.DB
}

func NewTeamStore(db *sqlx.DB) *TeamStore {
	return &TeamStore{
		db: db,
	}
}

func (s *TeamStore) Get(teamID int64) (*model.Team, error) {
	p := model.Team{ID: teamID}
	err := s.db.Get(&p, "SELECT * FROM teams WHERE id = $1 LIMIT 1;", p.ID)
	return &p, err
}

func (s *TeamStore) GetTeamMembers(teamID int64) (*model.TeamRecord, error) {
	p := []model.TeamRecord{}
	err := s.db.Select(&p, `
SELECT CAST($1 AS INT) as id, 0 as user_id, array_agg(u.first_name || ' ' || u.last_name) as members
FROM users as u, user_course as e
WHERE u.id = e.user_id
AND e.team_id = $1
GROUP BY e.team_id
;`, teamID)
	if err != nil {
		return nil, err
	}
	if len(p) < 1 {
		// Team has no members
		team := model.TeamRecord{ID: null.NewInt(0, false), UserID: 0, Members: pq.StringArray{}}
		return &team, nil
	}

	return &p[0], err
}

func (s *TeamStore) GetTeamMembersOfUser(user_id int64, course_id int64) (*model.TeamRecord, error) {
	p := []model.TeamRecord{}
	err := s.db.Select(&p, `
SELECT e.team_id as id, 0 AS user_id, array_agg(uo.first_name || ' ' || uo.last_name) as members
FROM users as u, user_course as e, users as uo, user_course as eo
WHERE u.id = e.user_id
AND e.course_id = $2
AND uo.id = eo.user_id
AND eo.course_id = $2
AND u.id = $1
AND e.team_id IS DISTINCT FROM NULL
AND eo.team_id = e.team_id
GROUP BY e.team_id
`, user_id, course_id)
	if err != nil {
		return nil, err
	}
	if len(p) < 1 {
		// User has no team
		team := model.TeamRecord{ID: null.NewInt(0, false), UserID:user_id, Members: pq.StringArray{}}
		return &team, nil
	}

	return &p[0], err
}

func (s *TeamStore) TeamID(userID int64, courseID int64) (null.Int, error) {
	p := []null.Int{}
	err := s.db.Select(&p, `
	SELECT team_id
	FROM user_course
	WHERE user_id = $1 AND course_id = $2
	LIMIT 1;
	`, userID, courseID)
	if err != nil {
		return null.NewInt(0, false), err
	}
	if len(p) < 1 {
		return null.NewInt(0, false), nil
	}
	return p[0], err
}

func (s *TeamStore) GetAllInGroup(groupID int64) ([]model.TeamRecord, error) {
	p := []model.TeamRecord{}
	err := s.db.Select(&p, `
	SELECT e.team_id AS id, 0 as user_id, array_agg(DISTINCT uo.first_name || ' ' || uo.last_name) AS members
	FROM user_course as e, users as u, user_group as g, users as uo, user_course AS eo, user_group AS go
	WHERE u.id = e.user_id
	AND u.id = g.user_id
	AND uo.id = eo.user_id
	AND uo.id = go.user_id
	AND u.id != uo.id
	AND go.group_id = g.group_id
	AND g.group_id = $1
	AND eo.team_id = e.team_id
	AND NOT e.team_confirmed
	AND NOT eo.team_confirmed
	GROUP BY e.team_id
	ORDER BY e.team_id, members
	`, groupID)
	return p, err
}

func (s *TeamStore) GetOtherUnaryTeamsInGroup(userID int64, groupID int64) ([]model.TeamRecord, error) {
	p := []model.TeamRecord{}
	err := s.db.Select(&p, `
	SELECT e.team_id AS id, u.id as user_id, ARRAY[u.first_name || ' ' || u.last_name] AS members
	FROM user_course as e, users as u, user_group as g
	WHERE u.id = e.user_id
	AND u.id = g.user_id
	AND g.group_id = $1
	AND e.team_id IS NULL
	AND u.id != $2
	ORDER BY u.last_name
	`, groupID, userID)
	return p, err
}

func (s *TeamStore) Confirmed(teamID int64, courseID int64) (*model.BoolRecord, error) {
	p := []model.BoolRecord{}
	err := s.db.Select(&p, `
	SELECT BOOL_AND(e.team_confirmed) as bool
	FROM user_course as e
	WHERE e.team_id = $1
	AND e.course_id = $2
	GROUP BY e.team_id
	LIMIT 1;
	`, teamID, courseID)
	if err != nil {
		return nil, err
	}
	if len(p) < 1 {
		// This should never happen
		return nil, errors.New("Failed to aggregate team confirmed state")
	}

	return &p[0], err
}

func (s *TeamStore) UserConfirmed(userID int64, courseID int64) (*model.BoolRecord, error) {
	p := model.BoolRecord{Bool: false}
	err := s.db.Get(&p, `
	SELECT e.team_confirmed as bool
	FROM user_course as e
	WHERE e.user_id = $1
	AND e.course_id = $2
	LIMIT 1
	`, userID, courseID)
	return &p, err
}

func (s *TeamStore) UserConfirm(userID int64, courseID int64) (error) {
	_, err := s.db.Exec(`
	UPDATE user_course
	SET team_confirmed = CAST(1 as BOOLEAN)
	WHERE user_id = $1
	AND course_id = $2;`, userID, courseID)
	return err
}

func (s *TeamStore) UnconfirmMembers(teamID int64) (error) {
	_, err := s.db.Exec(`
	UPDATE user_course
	SET team_confirmed = CAST(0 as BOOLEAN)
	WHERE team_id = $1;
	`, teamID)
	return err
}

func (s *TeamStore) UpdateTeam(userID int64, courseID int64, teamID null.Int, confirmed bool) (error) {
	_, err := s.db.Exec(`
	UPDATE user_course
	SET team_id = $1,
	    team_confirmed = CAST($2 as BOOLEAN)
	WHERE user_id = $3
	AND course_id = $4;
	`, teamID, confirmed, userID, courseID)
	return err
}

func (s *TeamStore) Delete(teamID int64) (error) {
	_, err := s.db.Exec(`
	DELETE FROM teams
	WHERE id = $1;
	`, teamID)
	return err
}

func (s *TeamStore) Create() (*model.Team, error) {
	var newTeamID int64
	err := s.db.QueryRow("INSERT INTO teams VALUES (DEFAULT) RETURNING id;").Scan(&newTeamID)
	if err != nil {
		return nil, err
	}
	return s.Get(newTeamID)
}
