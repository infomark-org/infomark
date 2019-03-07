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
  "context"
  "net/http"
  "strconv"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/auth/authorize"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
)

// TaskResource specifies Task management handler.
type TaskResource struct {
  Stores *Stores
}

// NewTaskResource create and returns a TaskResource.
func NewTaskResource(stores *Stores) *TaskResource {
  return &TaskResource{
    Stores: stores,
  }
}

// .............................................................................

// TaskResponse is the response payload for Task management.
type TaskResponse struct {
  *model.Task
}

// newTaskResponse creates a response from a Task model.
func newTaskResponse(p *model.Task) *TaskResponse {
  return &TaskResponse{
    Task: p,
  }
}

// Render post-processes a TaskResponse.
func (body *TaskResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// newTaskListResponse creates a response from a list of Task models.
func newTaskListResponse(Tasks []model.Task) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range Tasks {
    list = append(list, newTaskResponse(&Tasks[k]))
  }
  return list
}

// TaskResponse is the response payload for Task management.
type MissingTaskResponse struct {
  Task     *model.Task `json:"task"`
  CourseID int64       `json:"course_id"`
  SheetID  int64       `json:"sheet_id"`
}

// newTaskResponse creates a response from a Task model.
func newMissingTaskResponse(p *model.MissingTask) *MissingTaskResponse {
  return &MissingTaskResponse{
    Task:     p.Task,
    CourseID: p.CourseID,
    SheetID:  p.SheetID,
  }
}

// Render post-processes a TaskResponse.
func (body *MissingTaskResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// newTaskListResponse creates a response from a list of Task models.
func newMissingTaskListResponse(Tasks []model.MissingTask) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range Tasks {
    list = append(list, newMissingTaskResponse(&Tasks[k]))
  }
  return list
}

// .............................................................................
//
// IndexHandler is the enpoint for retrieving all Tasks if claim.root is true.
func (rs *TaskResource) IndexHandler(w http.ResponseWriter, r *http.Request) {

  var tasks []model.Task
  var err error
  // we use middle to detect whether there is a sheet given
  sheet := r.Context().Value("sheet").(*model.Sheet)
  tasks, err = rs.Stores.Task.TasksOfSheet(sheet.ID, false)

  // render JSON reponse
  if err = render.RenderList(w, r, newTaskListResponse(tasks)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// MissingIndexHandler is the enpoint for retrieving all task without a submission form the request identity
// URL : /tasks/missing
// METHOD: GET
func (rs *TaskResource) MissingIndexHandler(w http.ResponseWriter, r *http.Request) {

  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  tasks, err := rs.Stores.Task.GetAllMissingTasksForUser(accessClaims.LoginID)

  // TODO empty list
  if err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // render JSON reponse
  if err = render.RenderList(w, r, newMissingTaskListResponse(tasks)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// CreateHandler is the enpoint for retrieving all Tasks if claim.root is true.
func (rs *TaskResource) CreateHandler(w http.ResponseWriter, r *http.Request) {

  sheet := r.Context().Value("sheet").(*model.Sheet)

  // start from empty Request
  data := &TaskRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // create Task entry in database
  newTask, err := rs.Stores.Task.Create(data.Task, sheet.ID)
  if err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusCreated)

  // return Task information of created entry
  if err := render.Render(w, r, newTaskResponse(newTask)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

}

// GetHandler is the enpoint for retrieving a specific Task.
func (rs *TaskResource) GetHandler(w http.ResponseWriter, r *http.Request) {
  // `Task` is retrieved via middle-ware
  task := r.Context().Value("task").(*model.Task)

  // render JSON reponse
  if err := render.Render(w, r, newTaskResponse(task)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// EditHandler is the endpoint fro updating a specific Task with given id.
func (rs *TaskResource) EditHandler(w http.ResponseWriter, r *http.Request) {
  // start from empty Request
  data := &TaskRequest{
    Task: r.Context().Value("task").(*model.Task),
  }

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // update database entry
  if err := rs.Stores.Task.Update(data.Task); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

func (rs *TaskResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
  Task := r.Context().Value("task").(*model.Task)

  // update database entry
  if err := rs.Stores.Task.Delete(Task.ID); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

func (rs *TaskResource) GetPublicTestFileHandler(w http.ResponseWriter, r *http.Request) {

  task := r.Context().Value("task").(*model.Task)
  hnd := helper.NewPublicTestFileHandle(task.ID)

  if !hnd.Exists() {
    render.Render(w, r, ErrNotFound)
    return
  } else {
    if err := hnd.WriteToBody(w); err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    }
  }
}

func (rs *TaskResource) GetPrivateTestFileHandler(w http.ResponseWriter, r *http.Request) {

  task := r.Context().Value("task").(*model.Task)
  hnd := helper.NewPrivateTestFileHandle(task.ID)

  if !hnd.Exists() {
    render.Render(w, r, ErrNotFound)
    return
  } else {
    if err := hnd.WriteToBody(w); err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    }
  }
}

func (rs *TaskResource) ChangePublicTestFileHandler(w http.ResponseWriter, r *http.Request) {
  // will always be a POST
  task := r.Context().Value("task").(*model.Task)

  // the file will be located
  if err := helper.NewPublicTestFileHandle(task.ID).WriteToDisk(r, "file_data"); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }
  render.Status(r, http.StatusOK)
}

func (rs *TaskResource) ChangePrivateTestFileHandler(w http.ResponseWriter, r *http.Request) {
  // will always be a POST
  task := r.Context().Value("task").(*model.Task)

  // the file will be located
  if err := helper.NewPrivateTestFileHandle(task.ID).WriteToDisk(r, "file_data"); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }
  render.Status(r, http.StatusOK)
}

// GetSubmissionResultHandler returns the public submission result information
// URL: /task/{task_id}/result
// METHOD: GET
func (rs *TaskResource) GetSubmissionResultHandler(w http.ResponseWriter, r *http.Request) {
  givenRole := r.Context().Value("course_role").(authorize.CourseRole)

  if givenRole != authorize.STUDENT {
    render.Render(w, r, ErrBadRequest)
    return
  }

  // `Task` is retrieved via middle-ware
  task := r.Context().Value("task").(*model.Task)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  submission, err := rs.Stores.Submission.GetByUserAndTask(accessClaims.LoginID, task.ID)
  if err != nil {
    render.Render(w, r, ErrNotFound)
    return
  }

  grade, err := rs.Stores.Grade.GetForSubmission(submission.ID)
  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // TODO (patwie): does not make sense for TUTOR, ADMIN anyway

  grade.PrivateTestStatus = -1
  grade.PrivateTestLog = ""

  // render JSON reponse
  if err := render.Render(w, r, newGradeResponse(grade)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// .............................................................................
// Context middleware is used to load an Task object from
// the URL parameter `TaskID` passed through as the request. In case
// the Task could not be found, we stop here and return a 404.
// We do NOT check whether the identity is authorized to get this Task.
func (rs *TaskResource) Context(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // TODO: check permission if inquirer of request is allowed to access this Task
    // Should be done via another middleware
    var taskID int64
    var err error

    // try to get id from URL
    if taskID, err = strconv.ParseInt(chi.URLParam(r, "taskID"), 10, 64); err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // find specific Task in database
    task, err := rs.Stores.Task.Get(taskID)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    ctx := context.WithValue(r.Context(), "task", task)

    // when there is a taskID in the url, there is NOT a courseID in the url,
    // BUT: when there is a task, there is a course

    course, err := rs.Stores.Task.IdentifyCourseOfTask(task.ID)
    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }

    ctx = context.WithValue(ctx, "course", course)

    // serve next
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
