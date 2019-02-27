// InfoMark - a platform for managing Tasks with
//            distributing exercise Tasks and testing exercise submissions
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
  "errors"
  "net/http"
  "strconv"

  "github.com/cgtuebingen/infomark-backend/api/helper"
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

// TaskRequest is the request payload for Task management.
type TaskRequest struct {
  *model.Task
  ProtectedID int64 `json:"id"`
}

// TaskResponse is the response payload for Task management.
type TaskResponse struct {
  *model.Task
  Tasks []model.Task `json:"tasks"`
}

// newTaskResponse creates a response from a Task model.
func (rs *TaskResource) newTaskResponse(p *model.Task) *TaskResponse {

  return &TaskResponse{
    Task: p,
  }
}

// newTaskListResponse creates a response from a list of Task models.
func (rs *TaskResource) newTaskListResponse(Tasks []model.Task) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range Tasks {
    list = append(list, rs.newTaskResponse(&Tasks[k]))
  }

  return list
}

// Bind preprocesses a TaskRequest.
func (body *TaskRequest) Bind(r *http.Request) error {

  if body.Task == nil {
    return errors.New("Empty body")
  }

  // Sending the id via request-body is invalid.
  // The id should be submitted in the url.
  body.ProtectedID = 0

  return nil

}

// Render post-processes a TaskResponse.
func (body *TaskResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// IndexHandler is the enpoint for retrieving all Tasks if claim.root is true.
func (rs *TaskResource) IndexHandler(w http.ResponseWriter, r *http.Request) {

  var Tasks []model.Task
  var err error
  // we use middle to detect whether there is a sheet given
  sheet := r.Context().Value("sheet").(*model.Sheet)
  Tasks, err = rs.Stores.Task.TasksOfSheet(sheet.ID, false)

  // render JSON reponse
  if err = render.RenderList(w, r, rs.newTaskListResponse(Tasks)); err != nil {
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

  // validate final model
  if err := data.Task.Validate(); err != nil {
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
  if err := render.Render(w, r, rs.newTaskResponse(newTask)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  _ = sheet

}

// GetHandler is the enpoint for retrieving a specific Task.
func (rs *TaskResource) GetHandler(w http.ResponseWriter, r *http.Request) {
  // `Task` is retrieved via middle-ware
  Task := r.Context().Value("task").(*model.Task)

  // render JSON reponse
  if err := render.Render(w, r, rs.newTaskResponse(Task)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// PatchHandler is the endpoint fro updating a specific Task with given id.
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
  }
  render.Status(r, http.StatusOK)
}

func (rs *TaskResource) ChangePrivateTestFileHandler(w http.ResponseWriter, r *http.Request) {
  // will always be a POST
  task := r.Context().Value("task").(*model.Task)

  // the file will be located
  if err := helper.NewPrivateTestFileHandle(task.ID).WriteToDisk(r, "file_data"); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
  }
  render.Status(r, http.StatusOK)
}

// .............................................................................
// Context middleware is used to load an Task object from
// the URL parameter `TaskID` passed through as the request. In case
// the Task could not be found, we stop here and return a 404.
// We do NOT check whether the Task is authorized to get this Task.
func (d *TaskResource) Context(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // TODO: check permission if inquirer of request is allowed to access this Task
    // Should be done via another middleware
    var Task_id int64
    var err error

    // try to get id from URL
    if Task_id, err = strconv.ParseInt(chi.URLParam(r, "taskID"), 10, 64); err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // find specific Task in database
    Task, err := d.Stores.Task.Get(Task_id)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // serve next
    ctx := context.WithValue(r.Context(), "task", Task)
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
