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

type SubmissionStore struct {
	db *sqlx.DB
}

func NewSubmissionStore(db *sqlx.DB) *SubmissionStore {
	return &SubmissionStore{
		db: db,
	}
}

func (s *SubmissionStore) Get(submissionID int64) (*model.Submission, error) {
	p := model.Submission{ID: submissionID}
	err := s.db.Get(&p, `SELECT * FROM submissions WHERE id = $1 LIMIT 1;`, p.ID)
	return &p, err
}

func (s *SubmissionStore) GetByUserAndTask(userID int64, taskID int64) (*model.Submission, error) {
	p := model.Submission{}
	err := s.db.Get(&p, `
SELECT
  s.*, g.user_id AS "user_id"
FROM
  submissions s, grades g
WHERE
  s.id = g.submission_id
AND
  g.user_id = $1
AND
  s.task_id = $2
LIMIT 1;`,
		userID, taskID)
	return &p, err
}

func (s *SubmissionStore) GetByTeamID(teamID int64, taskID int64) (*model.Submission, error) {
	p := model.Submission{}
	err := s.db.Get(&p, `
SELECT
  s.*
FROM
  submissions s, teams t
WHERE
  s.team_id = t.id
AND
  s.task_id = $2
AND
  t.id = $1
		`, teamID, taskID)
	return &p, err
}

func (s *SubmissionStore) Create(p *model.Submission) (*model.Submission, error) {
	newID, err := Insert(s.db, "submissions", p)
	if err != nil {
		return nil, err
	}
	return s.Get(newID)
}

func (s *SubmissionStore) GetFiltered(filterCourseID, filterGroupID, filterUserID, filterSheetID, filterTaskID int64) ([]model.Submission, error) {

	p := []model.Submission{}
	err := s.db.Select(&p,
		`
SELECT
  s.*
FROM
  grades g
INNER JOIN submissions s ON g.submission_id = s.id
INNER JOIN user_group ug ON ug.user_id = g.user_id
INNER JOIN groups gr ON gr.id = ug.group_id
INNEr JOIN task_sheet ts ON ts.task_id = s.task_id
WHERE
  ($1 = 0 or g.user_id = $1)
AND
  ($2 = 0 or s.task_id = $2)
AND
  ($3 = 0 or ug.group_id = $3)
AND
  ($4 = 0 or ts.sheet_id = $4)
AND
  ($5 = 0 or gr.course_id = $5)
`,
		filterUserID, filterTaskID, filterGroupID, filterSheetID, filterCourseID)
	return p, err
}

func (s *SubmissionStore) Update(p *model.Submission) error {
	return Update(s.db, "submissions", p.ID, p)
}
