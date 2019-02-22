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

type CourseStore struct {
  db *sqlx.DB
}

func NewCourseStore(db *sqlx.DB) *CourseStore {
  return &CourseStore{
    db: db,
  }
}

func (s *CourseStore) Get(courseID int64) (*model.Course, error) {
  p := model.Course{ID: courseID}
  err := s.db.Get(&p, "SELECT * FROM courses WHERE id = $1 LIMIT 1;", p.ID)
  return &p, err
}

func (s *CourseStore) GetAll() ([]model.Course, error) {
  p := []model.Course{}
  err := s.db.Select(&p, "SELECT * FROM courses;")
  return p, err
}

func (s *CourseStore) Create(p *model.Course) (*model.Course, error) {
  newID, err := Insert(s.db, "courses", p)
  if err != nil {
    return nil, err
  }
  return s.Get(newID)
}

func (s *CourseStore) Update(p *model.Course) error {
  return Update(s.db, "courses", p.ID, p)
}

func (s *CourseStore) Delete(courseID int64) error {
  return Delete(s.db, "courses", courseID)
}

func (s *CourseStore) Enroll(courseID int64, userID int64) error {
  stmt := "INSERT INTO user_course (id, user_id, course_id, role) VALUES (DEFAULT, $1, $2, $3);"
  _, err := s.db.Exec(stmt, userID, courseID, 0)
  return err
}

func (s *CourseStore) Disenroll(courseID int64, userID int64) error {
  stmt := "DELETE FROM user_course WHERE user_id = $1 AND course_id = $2;"
  _, err := s.db.Exec(stmt, userID, courseID)
  return err
}

func (s *CourseStore) EnrolledUsers(course *model.Course) ([]model.UserCourseEnrollment, error) {
  p := []model.UserCourseEnrollment{}

  err := s.db.Select(&p, `
    SELECT
      uc.role, u.id, u.first_name, u.last_name, u.avatar_path, u.email,
      u.student_number, u.semester, u.subject, u.language FROM user_course uc
    INNER JOIN
      users u ON uc.user_id = u.id
    WHERE
      uc.course_id = $1;`, course.ID)
  return p, err
}
