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

func (s *TeamStore) GetAll(courseID int64) ([][]model.User, error) {
	p := [][]model.User{}
	err := s.db.Select(p, `
SELECT array_agg(u.*)
FROM users as u, user_course as e
WHERE e.course_id = $1
AND e.user_id = u.id
GROUP BY e.team_id
;`, courseID)
	// TODO: users without team missing
	return p, err
}

func (s *TeamStore) GetTeamMembers(teamID int64) (*model.TeamRecord, error) {
	p := model.TeamRecord{ID: null.NewInt(0, false), UserID: 0, Members: []string{}}
	err := s.db.Select(&p, `
SELECT $1 as id, 0 as user_id, array_agg(u.first_name || ' ' || u.last_name)
FROM users as u, user_course as e
WHERE u.id = e.user_id
AND e.team_id = $1
GROUP BY e.team_id
;`, teamID)
	return &p, err
}

func (s *TeamStore) GetTeamMembersOfUser(user_id int64, course_id int64) (*model.TeamRecord, error) {
	p := model.TeamRecord{ID: null.NewInt(0, false), UserID:user_id, Members: []string{}}
	err := s.db.Select(&p, `
SELECT e.team_id as id, e.user_id, array_agg(uo.first_name || ' ' || uo.last_name)
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

	return &p, err
}

func (s *TeamStore) TeamID(userID int64, courseID int64) (null.Int, error) {
	p := null.Int{}
	err := s.db.Select(&p, `
	SELECT team_id
	FROM user_course
	WHERE user_id = $1 AND course_id = $2
	LIMIT 1;
	`, userID, courseID)
	return p, err
}

func (s *TeamStore) GetAllInGroup(groupID int64) ([]model.TeamRecord, error) {
	p := []model.TeamRecord{}
	err := s.db.Select(&p, `
	SELECT e.team_id AS id, 0 as user_id, array_agg(uo.first_name || ' ' || uo.last_name) AS members
	FROM user_course as e, users as u, user_group as g, user as uo, user_course AS eo, user_group AS go
	WHERE u.id = e.user_id
	AND u.id = g.user_id
	AND uo.id = eo.user_id
	AND uo.id = go.user_id
	AND go.group_id = g.group_id
	AND g.group_id = $1
	AND eo.team_id = e.team_id
	AND e.team_confirmed = 'false'
	AND eo.team_confirmed = 'false'
	GROUP BY e.team_id
	ORDER BY u.last_name
	`, groupID)
	return p, err
}

func (s *TeamStore) GetUnaryTeamsInGroup(groupID int64) ([]model.TeamRecord, error) {
	p := []model.TeamRecord{}
	err := s.db.Select(&p, `
	SELECT e.team_id AS id, u.id as user_id, ARRAY[u.first_name || ' ' || u.last_name] AS members
	FROM user_course as e, users as u, user_group as g
	WHERE u.id = e.user_id
	AND u.id = g.user_id
	AND g.group_id = $1
	AND e.team_id IS NULL
	AND e.team_confirmed = 'false'
	ORDER BY u.last_name
	`, groupID)
	return p, err
}
