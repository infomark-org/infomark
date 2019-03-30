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

type GradeStore struct {
  db *sqlx.DB
}

func NewGradeStore(db *sqlx.DB) *GradeStore {
  return &GradeStore{
    db: db,
  }
}

// func (s *GradeStore) Get(id int64) (*model.Grade, error) {
//   p := model.Grade{ID: id}
//   err := s.db.Get(&p, "SELECT * FROM grades WHERE id = $1 LIMIT 1;", p.ID)
//   return &p, err
// }

func (s *GradeStore) Get(id int64) (*model.Grade, error) {
  p := model.Grade{ID: id}
  err := s.db.Get(&p, `
SELECT g.*, s.user_id,
u.last_name user_last_name,
u.first_name user_first_name,
u.email user_email
FROM grades g
INNER JOIN submissions s ON g.submission_id = s.id
INNER JOIN users u ON s.user_id = u.id
WHERE g.id = $1 LIMIT 1;
`, p.ID)
  return &p, err
}

func (s *GradeStore) Create(p *model.Grade) (*model.Grade, error) {
  newID, err := Insert(s.db, "grades", p)
  if err != nil {
    return nil, err
  }
  return s.Get(newID)
}

func (s *GradeStore) UpdatePrivateTestInfo(gradeID int64, log string, status int) error {
  _, err := s.db.Exec(`
    UPDATE grades
    SET private_execution_state=2, private_test_log=$2, private_test_status=$3
    WHERE id = $1;
    `, gradeID, log, status)
  return err
}

func (s *GradeStore) UpdatePublicTestInfo(gradeID int64, log string, status int) error {
  _, err := s.db.Exec(`
    UPDATE grades
    SET public_execution_state=2, public_test_log=$2, public_test_status=$3
    WHERE id = $1;
    `, gradeID, log, status)
  return err
}

func (s *GradeStore) GetForSubmission(id int64) (*model.Grade, error) {
  p := model.Grade{}
  err := s.db.Get(&p, "SELECT * FROM grades WHERE submission_id = $1 LIMIT 1;", id)
  return &p, err
}

func (s *GradeStore) GetAllMissingGrades(tutorID int64) ([]model.MissingGrade, error) {
  p := []model.MissingGrade{}

  err := s.db.Select(&p,
    `
SELECT g.*, ts.task_id, ts.sheet_id, sg.course_id, s.user_id,
u.last_name user_last_name,
u.first_name user_first_name,
u.email user_email
from grades g
INNER JOIN submissions s ON s.id = g.submission_id
INNER JOIN task_sheet ts ON ts.task_id = s.task_id
INNER JOIN sheet_course sg ON sg.sheet_id = ts.sheet_id
INNER JOIN users u ON s.user_id = u.id
WHERE g.feedback like '' and tutor_id = $1;
  `, tutorID)
  return p, err
}

func (s *GradeStore) Update(p *model.Grade) error {
  return Update(s.db, "grades", p.ID, p)
}

func (s *GradeStore) GetFiltered(
  courseID int64,
  sheetID int64,
  taskID int64,
  groupID int64,
  userID int64,
  tutorID int64,
  feedback string,
  acquiredPoints int,
  publicTestStatus int,
  privateTestStatus int,
  publicExecutationState int,
  privateExecutationState int,
) ([]model.Grade, error) {

  p := []model.Grade{}
  err := s.db.Select(&p,
    `
SELECT
  g.*, s.user_id,
  u.last_name user_last_name,
  u.first_name user_first_name,
  u.email user_email
FROM
  grades g
INNER JOIN submissions s ON s.id = g.submission_id
INNER JOIN task_sheet ts ON ts.task_id = s.task_id
INNER JOIN sheet_course sc ON sc.sheet_id = ts.sheet_id
INNER JOIN user_group ug ON ug.user_id = s.user_id
INNER JOIN users u ON s.user_id = u.id
WHERE course_id = $1
AND ug.group_id = $4
AND ($2 = 0 OR ts.sheet_id = $2)
AND ($3 = 0 OR s.task_id = $3)
AND ($5 = 0 OR ug.user_id = $5)
AND ($6 = 0 OR tutor_id = $6)
AND feedback LIKE $7
AND ($8 = -1 OR g.acquired_points = $8)
AND ($9 = -1 OR g.public_test_status = $9)
AND ($10 = -1 OR g.private_test_status = $10)
AND ($11 = -1 OR g.public_execution_state = $11)
AND ($12 = -1 OR g.private_execution_state = $12)
  `,
    // AND ($4 = 0 OR ug.group_id = $4)
    courseID,                // $1
    sheetID,                 // $2
    taskID,                  // $3
    groupID,                 // $4
    userID,                  // $5
    tutorID,                 // $6
    feedback,                // $7
    acquiredPoints,          // $8
    publicTestStatus,        // $9
    privateTestStatus,       // $10
    publicExecutationState,  // $11
    privateExecutationState, // $12
  )
  return p, err
}

func (s *GradeStore) IdentifyCourseOfGrade(gradeID int64) (*model.Course, error) {

  course := &model.Course{}
  err := s.db.Get(course,
    `
SELECT c.*
FROM grades g
INNER JOIN submissions s ON s.id = g.submission_id
INNER JOIN task_sheet ts ON ts.task_id = s.task_id
INNER JOIN sheet_course sc ON sc.sheet_id = ts.sheet_id
INNER JOIN courses c ON sc.course_id = c.id
WHERE g.id = $1`,
    gradeID)
  if err != nil {
    return nil, err
  }

  return course, err
}

func (s *GradeStore) IdentifyTaskOfGrade(gradeID int64) (*model.Task, error) {

  task := &model.Task{}
  err := s.db.Get(task,
    `
SELECT t.*
FROM grades g
INNER JOIN submissions s ON s.id = g.submission_id
INNER JOIN task_sheet ts ON ts.task_id = s.task_id
INNER JOIN tasks t ON ts.task_id = t.id
WHERE g.id = $1`,
    gradeID)
  if err != nil {
    return nil, err
  }

  return task, err
}
