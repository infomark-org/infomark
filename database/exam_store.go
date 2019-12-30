// InfoMark - a platform for managing exams with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  Infomark Authors
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

type ExamStore struct {
	db *sqlx.DB
}

func NewExamStore(db *sqlx.DB) *ExamStore {
	return &ExamStore{
		db: db,
	}
}

func (s *ExamStore) Get(examID int64) (*model.Exam, error) {
	p := model.Exam{ID: examID}
	err := s.db.Get(&p, "SELECT * FROM exams WHERE id = $1 LIMIT 1;", p.ID)
	return &p, err
}

func (s *ExamStore) ExamsOfCourse(courseID int64) ([]model.Exam, error) {
	p := []model.Exam{}
	err := s.db.Select(&p, "SELECT * FROM exams WHERE course_id = $1;", courseID)
	return p, err
}

func (s *ExamStore) GetAll() ([]model.Exam, error) {
	p := []model.Exam{}
	err := s.db.Select(&p, "SELECT * FROM exams;")
	return p, err
}

func (s *ExamStore) Create(p *model.Exam) (*model.Exam, error) {
	newID, err := Insert(s.db, "exams", p)
	if err != nil {
		return nil, err
	}
	return s.Get(newID)
}

func (s *ExamStore) Update(p *model.Exam) error {
	return Update(s.db, "exams", p.ID, p)
}

func (s *ExamStore) Delete(examID int64) error {
	return Delete(s.db, "exams", examID)
}

func (s *ExamStore) Enroll(examID int64, userID int64) error {
	err := s.Disenroll(examID, userID)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
INSERT INTO
  user_exam (id, user_id, exam_id, status, mark)
VALUES (DEFAULT, $1, $2, 0, '');
`, userID, examID)
	return err
}

func (s *ExamStore) Disenroll(examID int64, userID int64) error {
	_, err := s.db.Exec(`
DELETE FROM
  user_exam
WHERE
  user_id = $1
AND
  exam_id = $2; `, userID, examID)
	return err
}

func (s *ExamStore) GetEnrollmentsOfUser(userID int64) ([]model.UserExam, error) {
	p := []model.UserExam{}

	// , u.avatar_path
	err := s.db.Select(&p, `
SELECT
  ue.status,
  ue.mark,
  ue.user_id,
  ue.exam_id,
  e.course_id,
  ue.id
FROM
  user_exam ue
INNER JOIN exams e ON ue.exam_id = e.id
WHERE
  ue.user_id = $1`, userID,
	)
	return p, err
}

func (s *ExamStore) GetEnrollmentOfUser(examID int64, userID int64) (*model.UserExam, error) {
	p := model.UserExam{}

	// , u.avatar_path
	err := s.db.Get(&p, `
SELECT
  ue.status,
  ue.mark,
  ue.user_id,
  ue.exam_id,
  e.course_id,
  ue.id
FROM
  user_exam ue
INNER JOIN exams e ON ue.exam_id = e.id
WHERE
  ue.user_id = $1
AND
  ue.exam_id = $2
LIMIT 1`, userID, examID,
	)
	return &p, err
}

func (s *ExamStore) UpdateUserExam(p *model.UserExam) error {

	return Update(s.db, "user_exam", p.ID, p)
}

func (s *ExamStore) GetEnrollmentsInCourseOfExam(courseID int64, examID int64) ([]model.UserExam, error) {
	p := []model.UserExam{}

	// , u.avatar_path
	err := s.db.Select(&p, `
SELECT
  ue.status,
  ue.mark,
  ue.user_id,
  ue.exam_id,
  e.course_id,
  ue.id
FROM
  user_exam ue
INNER JOIN exams e ON ue.exam_id = e.id
WHERE
  ue.exam_id = $1
AND
  e.course_id = $2`, examID, courseID,
	)
	return p, err
}
