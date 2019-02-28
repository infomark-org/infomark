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

package database

import (
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/jmoiron/sqlx"
)

type GroupStore struct {
  db *sqlx.DB
}

func NewGroupStore(db *sqlx.DB) *GroupStore {
  return &GroupStore{
    db: db,
  }
}

func (s *GroupStore) Get(groupID int64) (*model.Group, error) {
  p := model.Group{ID: groupID}
  err := s.db.Get(&p, "SELECT * FROM groups WHERE id = $1 LIMIT 1;", p.ID)
  return &p, err
}

func (s *GroupStore) GetAll() ([]model.Group, error) {
  p := []model.Group{}
  err := s.db.Select(&p, "SELECT * FROM groups;")
  return p, err
}

func (s *GroupStore) Create(p *model.Group) (*model.Group, error) {
  // create Group
  newID, err := Insert(s.db, "groups", p)
  if err != nil {
    return nil, err
  }

  return s.Get(newID)
}

func (s *GroupStore) Update(p *model.Group) error {
  return Update(s.db, "groups", p.ID, p)
}

func (s *GroupStore) Delete(taskID int64) error {
  return Delete(s.db, "groups", taskID)
}

func (s *GroupStore) GroupsOfCourse(courseID int64) ([]model.Group, error) {
  p := []model.Group{}

  err := s.db.Select(&p, `
    SELECT
      *
    FROM
      groups
    WHERE
      course_id = $1`, courseID)
  return p, err
}
