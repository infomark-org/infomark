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

package database

import (
	"github.com/infomark-org/infomark/model"
	"github.com/jmoiron/sqlx"
)

type UserStore struct {
	db *sqlx.DB
}

func NewUserStore(db *sqlx.DB) *UserStore {
	return &UserStore{
		db: db,
	}
}

func (s *UserStore) Get(userID int64) (*model.User, error) {
	p := model.User{ID: userID}
	err := s.db.Get(&p, "SELECT * FROM users WHERE id = $1 LIMIT 1;", p.ID)
	return &p, err
}

func (s *UserStore) FindByEmail(email string) (*model.User, error) {
	p := model.User{Email: email}
	err := s.db.Get(&p, "SELECT * FROM users WHERE email = $1 LIMIT 1;", p.Email)
	return &p, err
}

func (s *UserStore) Find(query string) ([]model.User, error) {
	p := []model.User{}
	err := s.db.Select(&p, `
SELECT
  *
FROM
  users
WHERE
 last_name LIKE $1
OR
 first_name LIKE $1
OR
 email LIKE $1`, query)
	return p, err
}

func (s *UserStore) GetAll() ([]model.User, error) {
	p := []model.User{}
	err := s.db.Select(&p, "SELECT * FROM users;")
	return p, err
}

func (s *UserStore) Create(p *model.User) (*model.User, error) {
	newID, err := Insert(s.db, "users", p)
	if err != nil {
		return nil, err
	}
	return s.Get(newID)
}

func (s *UserStore) Update(p *model.User) error {
	return Update(s.db, "users", p.ID, p)
}

func (s *UserStore) Delete(userID int64) error {
	return Delete(s.db, "users", userID)
}

func (s *UserStore) GetEnrollments(userID int64) ([]model.Enrollment, error) {
	p := []model.Enrollment{}
	err := s.db.Select(&p, `
SELECT
  course_id,
  role
FROM
  user_course
WHERE
  user_id = $1
`, userID)
	return p, err

}

func (s *UserStore) GetFromGrade(gradeID int64) (*model.User, error) {
	r := model.User{}
	err := s.db.Get(&r, `
SELECT u.*
FROM grades AS g,
	   submissions AS s,
		 users AS u
WHERE g.id = $1
AND g.submission_id = s.id
AND s.user_id = u.id
LIMIT 1;
`, gradeID)
	return &r, err
}
