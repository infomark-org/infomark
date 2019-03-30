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
  "github.com/lib/pq"
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

func (s *GroupStore) GroupsOfCourse(courseID int64) ([]model.GroupWithTutor, error) {
  p := []model.GroupWithTutor{}

  err := s.db.Select(&p, `
    SELECT
      g.*,
u.first_name as tutor_first_name,
u.last_name as tutor_last_name,
u.avatar_url as tutor_avatar_url,
u.email as tutor_email,
u.language as tutor_language
    FROM
      groups g
    INNER JOIN users u ON g.tutor_id = u.id
    WHERE
      course_id = $1
      ORDER BY
      g.id ASC`, courseID)
  return p, err
}

func (s *GroupStore) GetMembers(groupID int64) ([]model.User, error) {
  p := []model.User{}

  err := s.db.Select(&p, `
SELECT
  u.*
FROM
  users u
INNER JOIN
  user_group ug ON ug.user_id = u.id
WHERE
  ug.group_id = $1`, groupID)
  return p, err
}

func (s *GroupStore) GetInCourseWithUser(userID int64, courseID int64) ([]model.GroupWithTutor, error) {
  p := []model.GroupWithTutor{}

  err := s.db.Select(&p, `
    SELECT
      g.*,
u.first_name as tutor_first_name,
u.last_name as tutor_last_name,
u.avatar_url as tutor_avatar_url,
u.email as tutor_email,
u.language as tutor_language
    FROM
      groups g
    INNER JOIN users u ON g.tutor_id = u.id
    INNER JOIN
      user_group ug on g.id = ug.group_id
    WHERE
      course_id = $2
    AND ug.user_id = $1
      ORDER BY
      g.id ASC`, userID, courseID)
  return p, err
}

func (s *GroupStore) EnrolledUsers(
  courseID int64,
  groupID int64,
  roleFilter []string,
  filterFirstName string,
  filterLastName string,
  filterEmail string,
  filterSubject string,
  filterLanguage string) ([]model.UserCourse, error) {
  p := []model.UserCourse{}

  // , u.avatar_path
  err := s.db.Select(&p, `
   SELECT
      uc.role, u.id, u.first_name, u.last_name, u.email,
      u.student_number, u.semester, u.subject, u.language, u.avatar_url FROM user_course uc
    INNER JOIN
      users u ON uc.user_id = u.id
    INNER JOIN
      user_group ug ON ug.user_id = u.id
    WHERE
      uc.course_id = $1
    AND
      ug.group_id = $2
    AND
      uc.role = ANY($3)
    AND
      LOWER(u.first_name) LIKE $4
    AND
      LOWER(u.last_name) LIKE $5
    AND
      LOWER(u.email) LIKE $6
    AND
      LOWER(u.subject) LIKE $7
    AND
      LOWER(u.language) LIKE $8
    `, courseID, groupID, pq.Array(roleFilter),
    filterFirstName, filterLastName, filterEmail,
    filterSubject, filterLanguage,
  )
  return p, err
}

func (s *GroupStore) GetGroupEnrollmentOfUserInCourse(userID int64, courseID int64) (*model.GroupEnrollment, error) {
  p := &model.GroupEnrollment{}
  err := s.db.Get(p, `
    SELECT ug.*
FROM user_group ug
INNER JOIN groups g ON g.id = ug.group_id
WHERE ug.user_id = $1
AND g.course_id = $2`, userID, courseID)
  return p, err
}

func (s *GroupStore) CreateGroupEnrollmentOfUserInCourse(p *model.GroupEnrollment) (*model.GroupEnrollment, error) {
  newID, err := Insert(s.db, "user_group", p)
  if err != nil {
    return nil, err
  }

  res := &model.GroupEnrollment{}

  err = s.db.Get(res, `SELECT * FROM user_group WHERE id= $1`, newID)
  if err != nil {
    return nil, err
  }

  return res, nil
}

func (s *GroupStore) ChangeGroupEnrollmentOfUserInCourse(p *model.GroupEnrollment) error {
  return Update(s.db, "user_group", p.ID, p)
}

func (s *GroupStore) GetOfTutor(tutorID int64, courseID int64) ([]model.GroupWithTutor, error) {
  p := []model.GroupWithTutor{}

  err := s.db.Select(&p, `
    SELECT
      g.*,
u.first_name as tutor_first_name,
u.last_name as tutor_last_name,
u.avatar_url as tutor_avatar_url,
u.email as tutor_email,
u.language as tutor_language
    FROM
      groups g
    INNER JOIN users u ON g.tutor_id = u.id
    WHERE
      course_id = $2
    AND g.tutor_id = $1
      ORDER BY
      g.id ASC`, tutorID, courseID)
  return p, err
}

func (s *GroupStore) IdentifyCourseOfGroup(groupID int64) (*model.Course, error) {

  course := &model.Course{}
  err := s.db.Get(course,
    `
SELECT
  c.*
FROM
  groups g
INNER JOIN
  courses c ON c.id = g.course_ID
WHERE g.id = $1`,
    groupID)
  if err != nil {
    return nil, err
  }

  return course, err
}

func (s *GroupStore) GetBidOfUserForGroup(userID int64, groupID int64) (bid int, err error) {
  err = s.db.Get(&bid, `
SELECT
  bid
FROM
  group_bids
WHERE
  user_id = $1 and group_id = $2
LIMIT 1`, userID, groupID)
  return bid, err
}

func (s *GroupStore) InsertBidOfUserForGroup(userID int64, groupID int64, bid int) (int, error) {

  // insert
  _, err := Insert(s.db, "group_bids", &model.GroupBid{UserID: userID, GroupID: groupID, Bid: bid})
  if err != nil {
    return 0, err
  }

  return s.GetBidOfUserForGroup(userID, groupID)
}

func (s *GroupStore) UpdateBidOfUserForGroup(userID int64, groupID int64, bid int) (int, error) {

  // update
  _, err := s.db.Exec(`
UPDATE
  group_bids
SET bid = $3
WHERE
  user_id = $1 and group_id = $2`, userID, groupID, bid)
  if err != nil {
    return 0, err
  }

  return s.GetBidOfUserForGroup(userID, groupID)
}

func (s *GroupStore) GetBidsForCourseForUser(courseID int64, userID int64) ([]model.GroupBid, error) {

  p := []model.GroupBid{}

  err := s.db.Select(&p, `
SELECT gb.*
FROM
group_bids gb
INNER JOIN
  groups g ON gb.group_id = g.id
WHERE gb.user_id = $2
AND g.course_id = $1`, courseID, userID)
  return p, err

}

func (s *GroupStore) GetBidsForCourse(courseID int64) ([]model.GroupBid, error) {

  p := []model.GroupBid{}

  err := s.db.Select(&p, `
SELECT gb.*
FROM
group_bids gb
INNER JOIN
  groups g ON gb.group_id = g.id
AND g.course_id = $1`, courseID)
  return p, err

}
