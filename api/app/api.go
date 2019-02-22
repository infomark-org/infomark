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
  "github.com/jmoiron/sqlx"
  "github.com/sirupsen/logrus"
)

type ctxKey int

const (
  ctxAccount ctxKey = iota
  ctxProfile
)

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
}

func NewStores(courseStore CourseStore, userStore UserStore) *Stores {
  return &Stores{
    Course: courseStore,
    User:   userStore,
  }
}

// NewAPI configures and returns application API.
func NewAPI(db *sqlx.DB) (*API, error) {

  userStore := database.NewUserStore(db)
  courseStore := database.NewCourseStore(db)
  sheetStore := database.NewSheetStore(db)
  taskStore := database.NewTaskStore(db)

  stores := NewStores(courseStore, userStore)

  user := NewUserResource(userStore)
  account := NewAccountResource(userStore)
  auth := NewAuthResource(userStore)
  course := NewCourseResource(stores)
  sheet := NewSheetResource(sheetStore)
  task := NewTaskResource(taskStore)

  api := &API{
    Account: account,
    Auth:    auth,
    User:    user,
    Course:  course,
    Sheet:   sheet,
    Task:    task,
  }
  return api, nil
}

func log(r *http.Request) logrus.FieldLogger {
  return logging.GetLogEntry(r)
}
