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

type TaskStore struct {
  db *sqlx.DB
}

func NewTaskStore(db *sqlx.DB) *TaskStore {
  return &TaskStore{
    db: db,
  }
}

func (s *TaskStore) GetAllMissingTasksForUser(userID int64) ([]model.MissingTask, error) {
  p := []model.MissingTask{}
  err := s.db.Select(&p, `
SELECT t.*, ts.sheet_id, sc.course_id from tasks  t
INNER JOIN task_sheet ts ON ts.task_id = t.id
INNER JOIN sheet_course sc ON sc.sheet_id = ts.sheet_id
WHERE t.id NOT IN (SELECT task_id FROM submissions s WHERE s.user_id = $1);
    `, userID)
  return p, err
}

func (s *TaskStore) Get(taskID int64) (*model.Task, error) {
  p := model.Task{ID: taskID}
  err := s.db.Get(&p, "SELECT * FROM tasks WHERE id = $1 LIMIT 1;", p.ID)
  return &p, err
}

func (s *TaskStore) GetAll() ([]model.Task, error) {
  p := []model.Task{}
  err := s.db.Select(&p, "SELECT * FROM tasks;")
  return p, err
}

func (s *TaskStore) Create(p *model.Task, sheetID int64) (*model.Task, error) {
  // create Task
  newID, err := Insert(s.db, "tasks", p)
  if err != nil {
    return nil, err
  }

  // now associate sheet with course
  _, err = s.db.Exec(`INSERT INTO task_sheet
    (id,task_id,sheet_id)
    VALUES (DEFAULT, $1, $2);`, newID, sheetID)
  if err != nil {
    return nil, err
  }

  return s.Get(newID)
}

func (s *TaskStore) Update(p *model.Task) error {
  return Update(s.db, "tasks", p.ID, p)
}

func (s *TaskStore) Delete(taskID int64) error {
  return Delete(s.db, "tasks", taskID)
}

func (s *TaskStore) TasksOfSheet(sheetID int64, only_active bool) ([]model.Task, error) {
  p := []model.Task{}

  // t.public_test_path, t.private_test_path,
  err := s.db.Select(&p, `
    SELECT
      t.id, t.created_at, t.updated_at, t.max_points,
      t.public_docker_image, t.private_docker_image
    FROM task_sheet ts
    INNER JOIN
      tasks t ON ts.task_id = t.id
    INNER JOIN
      sheets s ON ts.sheet_id = s.id
    WHERE
      s.id = $1
    ORDER BY
      t.name ASC;`, sheetID)
  return p, err
}

func (s *TaskStore) IdentifyCourseOfTask(taskID int64) (*model.Course, error) {

  course := &model.Course{}
  err := s.db.Get(course,
    `
SELECT
  c.*
FROM
  task_sheet ts
INNER JOIN
  sheet_course sc ON sc.sheet_id = ts.sheet_id
INNER JOIN
  courses c ON c.id = sc.course_ID
WHERE ts.task_id = $1`,
    taskID)
  if err != nil {
    return nil, err
  }

  return course, err
}

func (s *TaskStore) IdentifySheetOfTask(taskID int64) (*model.Sheet, error) {

  sheet := &model.Sheet{}
  err := s.db.Get(sheet,
    `
SELECT
  s.*
FROM
  task_sheet ts
INNER JOIN
  sheets s ON s.id = ts.sheet_id
WHERE ts.task_id = $1`,
    taskID)
  if err != nil {
    return nil, err
  }

  return sheet, err
}

func (s *TaskStore) GetAverageRating(taskID int64) (float32, error) {
  var averageRating float32
  err := s.db.Get(&averageRating, `
SELECT
  AVG(rating) average_rating
FROM
  task_ratings tr
WHERE tr.task_id  = $1`, taskID)
  return averageRating, err
}

func (s *TaskStore) GetRatingOfTaskByUser(taskID int64, userID int64) (*model.TaskRating, error) {

  p := model.TaskRating{}
  err := s.db.Get(&p, `
    SELECT * from task_ratings where user_id = $1 and task_id = $2 LIMIT 1`, userID, taskID)
  return &p, err
}

func (s *TaskStore) GetRating(taskRatingID int64) (*model.TaskRating, error) {
  p := model.TaskRating{ID: taskRatingID}
  err := s.db.Get(&p, "SELECT * FROM task_ratings WHERE id = $1 LIMIT 1;", p.ID)
  return &p, err
}

func (s *TaskStore) CreateRating(p *model.TaskRating) (*model.TaskRating, error) {
  // create Task
  newID, err := Insert(s.db, "task_ratings", p)
  if err != nil {
    return nil, err
  }
  return s.GetRating(newID)
}

func (s *TaskStore) UpdateRating(p *model.TaskRating) error {
  return Update(s.db, "task_ratings", p.ID, p)
}
