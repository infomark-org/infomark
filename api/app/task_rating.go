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

  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/render"
)

// TaskRatingResource specifies TaskRating management handler.
type TaskRatingResource struct {
  Stores *Stores
}

// NewTaskRatingResource create and returns a TaskRatingResource.
func NewTaskRatingResource(stores *Stores) *TaskRatingResource {
  return &TaskRatingResource{
    Stores: stores,
  }
}

// .............................................................................

// TaskRatingResponse is the response payload for TaskRating management.
type TaskRatingResponse struct {
  TaskID        int64   `json:"task_id"`
  AverageRating float32 `json:"average_rating"`
  OwnRating     int     `json:"own_rating"`
}

// newTaskRatingResponse creates a response from a TaskRating model.
func (rs *TaskRatingResource) newTaskRatingResponse(p *model.TaskRating, averageRating float32) *TaskRatingResponse {

  return &TaskRatingResponse{
    TaskID:        p.TaskID,
    OwnRating:     p.Rating,
    AverageRating: averageRating,
  }
}

// Render post-processes a TaskRatingResponse.
func (body *TaskRatingResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// GetHandler is the enpoint for the accumulated results.
// url:    /tasks/{task_id}/ratings
// method: GET
func (rs *TaskRatingResource) GetHandler(w http.ResponseWriter, r *http.Request) {
  // `TaskRating` is retrieved via middle-ware
  task := r.Context().Value("task").(*model.Task)

  // get rating
  averageRating, err := rs.Stores.Task.GetAverageRating(task.ID)
  if err != nil {
    // no entries so far
    averageRating = 0
  }

  // get own rating
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)
  taskRating, err := rs.Stores.Task.GetRatingOfTaskByUser(task.ID, accessClaims.LoginID)
  if err != nil {
    // record not found
    taskRating = &model.TaskRating{
      UserID: accessClaims.LoginID,
      TaskID: task.ID,
      Rating: 0,
    }
  }

  // render JSON reponse
  if err := render.Render(w, r, rs.newTaskRatingResponse(taskRating, averageRating)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// ChangeHandler is the enpoint for identities to rate a given task
// url:    /tasks/{task_id}/ratings
// method: POST
func (rs *TaskRatingResource) ChangeHandler(w http.ResponseWriter, r *http.Request) {
  // `TaskRating` is retrieved via middle-ware
  task := r.Context().Value("task").(*model.Task)
  accessClaims := r.Context().Value("access_claims").(*authenticate.AccessClaims)

  data := &TaskRatingRequest{TaskRating: &model.TaskRating{}}
  data.TaskRating.UserID = accessClaims.LoginID
  data.TaskRating.TaskID = task.ID
  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }
  // fmt.Println(data.TaskRating)
  // fmt.Println(data.TaskRating.Rating)
  givenRating, err := rs.Stores.Task.GetRatingOfTaskByUser(task.ID, accessClaims.LoginID)

  if err == nil {
    // there is a rating
    data.TaskRating.ID = givenRating.ID
    if err := rs.Stores.Task.UpdateRating(data.TaskRating); err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }
    render.Status(r, http.StatusNoContent)

  } else {
    // there is no rating so far from the user

    newTaskRating, err := rs.Stores.Task.CreateRating(data.TaskRating)
    if err != nil {
      render.Render(w, r, ErrRender(err))
      return
    }

    // get rating
    averageRating, err := rs.Stores.Task.GetAverageRating(task.ID)
    if err != nil {
      // no entries so far
      averageRating = 0
    }

    render.Status(r, http.StatusCreated)

    // return Task information of created entry
    if err := render.Render(w, r, rs.newTaskRatingResponse(newTaskRating, averageRating)); err != nil {
      render.Render(w, r, ErrRender(err))
      return
    }
  }

}
