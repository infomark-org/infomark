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

package app

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/infomark-org/infomark-backend/auth/authenticate"
	"github.com/infomark-org/infomark-backend/model"
	"github.com/infomark-org/infomark-backend/symbol"
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

// GetHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}/ratings
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: get
// TAG: tasks
// RESPONSE: 200,TaskRatingResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get all stats (average rating, own rating, ..) for a task
func (rs *TaskRatingResource) GetHandler(w http.ResponseWriter, r *http.Request) {
	// `TaskRating` is retrieved via middle-ware
	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)

	// get rating
	averageRating, err := rs.Stores.Task.GetAverageRating(task.ID)
	if err != nil {
		// no entries so far
		averageRating = 0
	}

	// get own rating
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
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

// ChangeHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}/ratings
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: post
// TAG: tasks
// REQUEST: TaskRatingRequest
// RESPONSE: 200,TaskRatingResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  updates and gets all stats (average rating, own rating, ..) for a task
func (rs *TaskRatingResource) ChangeHandler(w http.ResponseWriter, r *http.Request) {
	// `TaskRating` is retrieved via middle-ware
	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)

	data := &TaskRatingRequest{}
	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	givenRating, err := rs.Stores.Task.GetRatingOfTaskByUser(task.ID, accessClaims.LoginID)

	if err == nil {
		// there is a rating
		givenRating.Rating = data.Rating
		if err := rs.Stores.Task.UpdateRating(givenRating); err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}
		render.Status(r, http.StatusNoContent)

	} else {
		// there is no rating so far from the user

		rating := &model.TaskRating{
			UserID: accessClaims.LoginID,
			TaskID: task.ID,
			Rating: data.Rating,
		}

		newTaskRating, err := rs.Stores.Task.CreateRating(rating)
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
