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

package app

import (
  "net/http"

  "github.com/cgtuebingen/infomark-backend/database"
  "github.com/cgtuebingen/infomark-backend/logging"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/jmoiron/sqlx"
  "github.com/sirupsen/logrus"
)

type ctxKey int

const (
  ctxAccount ctxKey = iota
  ctxProfile
)

type UserStore interface {
  Get(userID int64) (*model.User, error)
  Update(p *model.User) error
  GetAll() ([]model.User, error)
  Create(p *model.User) (*model.User, error)
  FindByEmail(email string) (*model.User, error)
  GetEnrollments(userID int64) ([]model.Enrollment, error)
}

// CourseStore specifies required database queries for course management.
type CourseStore interface {
  Get(courseID int64) (*model.Course, error)
  Update(p *model.Course) error
  GetAll() ([]model.Course, error)
  Create(p *model.Course) (*model.Course, error)
  Delete(courseID int64) error
  Enroll(courseID int64, userID int64) error
  Disenroll(courseID int64, userID int64) error
  EnrolledUsers(course *model.Course) ([]model.UserCourse, error)
}

// SheetStore specifies required database queries for Sheet management.
type SheetStore interface {
  Get(SheetID int64) (*model.Sheet, error)
  Update(p *model.Sheet) error
  GetAll() ([]model.Sheet, error)
  Create(p *model.Sheet, c *model.Course) (*model.Sheet, error)
  Delete(SheetID int64) error
  SheetsOfCourse(course *model.Course, only_active bool) ([]model.Sheet, error)
}

// TaskStore specifies required database queries for Task management.
type TaskStore interface {
  Get(TaskID int64) (*model.Task, error)
  Update(p *model.Task) error
  GetAll() ([]model.Task, error)
  Create(p *model.Task, s *model.Sheet) (*model.Task, error)
  Delete(TaskID int64) error
  TasksOfSheet(Sheet *model.Sheet, only_active bool) ([]model.Task, error)
}

// API provides application resources and handlers.
type API struct {
  User    *UserResource
  Account *AccountResource
  Auth    *AuthResource
  Course  *CourseResource
  Sheet   *SheetResource
  Task    *TaskResource
}

type Stores struct {
  Course CourseStore
  User   UserStore
  Sheet  SheetStore
  Task   TaskStore
}

func NewStores(db *sqlx.DB) *Stores {

  return &Stores{
    Course: database.NewCourseStore(db),
    User:   database.NewUserStore(db),
    Sheet:  database.NewSheetStore(db),
    Task:   database.NewTaskStore(db),
  }
}

// NewAPI configures and returns application API.
func NewAPI(db *sqlx.DB) (*API, error) {

  stores := NewStores(db)

  api := &API{
    Account: NewAccountResource(stores),
    Auth:    NewAuthResource(stores),
    User:    NewUserResource(stores),
    Course:  NewCourseResource(stores),
    Sheet:   NewSheetResource(stores),
    Task:    NewTaskResource(stores),
  }
  return api, nil
}

func log(r *http.Request) logrus.FieldLogger {
  return logging.GetLogEntry(r)
}
