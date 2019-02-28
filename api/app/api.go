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
  "github.com/cgtuebingen/infomark-backend/database"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/jmoiron/sqlx"
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
  Delete(userID int64) error
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
  EnrolledUsers(
    course *model.Course,
    roleFilter []string,
    filterFirstName string,
    filterLastName string,
    filterEmail string,
    filterSubject string,
    filterLanguage string) ([]model.UserCourse, error)
  PointsForUser(userID int64, courseID int64) ([]model.SheetPoints, error)
  RoleInCourse(userID int64, courseID int64) (database.CourseRole, error)
}

// SheetStore specifies required database queries for Sheet management.
type SheetStore interface {
  Get(SheetID int64) (*model.Sheet, error)
  Update(p *model.Sheet) error
  GetAll() ([]model.Sheet, error)
  Create(p *model.Sheet, courseID int64) (*model.Sheet, error)
  Delete(SheetID int64) error
  SheetsOfCourse(courseID int64, only_active bool) ([]model.Sheet, error)
}

// TaskStore specifies required database queries for Task management.
type TaskStore interface {
  Get(TaskID int64) (*model.Task, error)
  Update(p *model.Task) error
  GetAll() ([]model.Task, error)
  Create(p *model.Task, sheetID int64) (*model.Task, error)
  Delete(TaskID int64) error
  TasksOfSheet(sheetID int64, only_active bool) ([]model.Task, error)
}

// GroupStore specifies required database queries for Task management.
type GroupStore interface {
  Get(groupID int64) (*model.Group, error)
  GetAll() ([]model.Group, error)
  Create(p *model.Group) (*model.Group, error)
  Update(p *model.Group) error
  Delete(taskID int64) error
  GroupsOfCourse(courseID int64) ([]model.Group, error)
  GetInCourseWithUser(userID int64, courseID int64) (*model.Group, error)
  GetOfTutor(tutorID int64, courseID int64) (*model.Group, error)
}

// API provides application resources and handlers.
type API struct {
  User    *UserResource
  Account *AccountResource
  Auth    *AuthResource
  Course  *CourseResource
  Sheet   *SheetResource
  Task    *TaskResource
  Group   *GroupResource
}

type Stores struct {
  Course CourseStore
  User   UserStore
  Sheet  SheetStore
  Task   TaskStore
  Group  GroupStore
}

func NewStores(db *sqlx.DB) *Stores {

  return &Stores{
    Course: database.NewCourseStore(db),
    User:   database.NewUserStore(db),
    Sheet:  database.NewSheetStore(db),
    Task:   database.NewTaskStore(db),
    Group:  database.NewGroupStore(db),
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
    Group:   NewGroupResource(stores),
  }
  return api, nil
}

// func log(r *http.Request) logrus.FieldLogger {
//   return logging.GetLogEntry(r)
// }
