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
`, courseID)
// TODO: users without team missing
	return p, err
}

func (s *TeamStore) GetTeamMembers(teamID int64) ([]model.User, error) {
	p:= []model.User{}
	err := s.db.Select(&p, `
SELECT u.*
FROM users as u, user_course as e
WHERE u.id = e.user_id
AND e.team_id = $1;
`, teamID)
	return p, err
}

func (s *TeamStore) GetTeamMembersOfUser(user_id int64, course_id int64) ([]model.User, error) {
	p:= []model.User{}
	err := s.db.Select(&p, `
SELECT u.*
FROM users as u, user_course as e
WHERE u.id = e.user_id
AND u.id = $1
AND e.course_id = $2
AND e.team_id IS DISTINCT FROM NULL
AND e.team_id = (SELECT team_id FROM user_course WHERE user_id = $1 AND course_id = $2);
`, user_id, course_id)
	return p, err
}

func (s *TeamStore) HasTeam(userID int64, courseID int64) (bool, error) {
	p := null.Int{}
	err := s.db.Select(&p, `
	SELECT team_id
	FROM user_course
	WHERE user_id = $1 AND course_id = $2
	LIMIT 1;
	`, userID, courseID)
	return p.Valid, err
}

