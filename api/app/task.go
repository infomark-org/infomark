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
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/infomark-org/infomark/api/helper"
	"github.com/infomark-org/infomark/auth/authenticate"
	"github.com/infomark-org/infomark/auth/authorize"
	"github.com/infomark-org/infomark/model"
	"github.com/infomark-org/infomark/symbol"
	null "gopkg.in/guregu/null.v3"
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

// IndexHandler is public endpoint for
// URL: /courses/{course_id}/sheets/{sheet_id}/tasks
// URLPARAM: course_id,integer
// URLPARAM: sheet_id,integer
// METHOD: get
// TAG: tasks
// RESPONSE: 200,TaskResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  Get all tasks of a given sheet
func (rs *TaskResource) IndexHandler(w http.ResponseWriter, r *http.Request) {

	var tasks []model.Task
	var err error
	// we use middle to detect whether there is a sheet given
	sheet := r.Context().Value(symbol.CtxKeySheet).(*model.Sheet)
	tasks, err = rs.Stores.Task.TasksOfSheet(sheet.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// render JSON response
	if err = render.RenderList(w, r, newTaskListResponse(tasks)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// MissingIndexHandler is public endpoint for
// URL: /courses/{course_id}/tasks/missing
// URLPARAM: course_id,integer
// METHOD: get
// TAG: tasks
// RESPONSE: 200,MissingTaskResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// SUMMARY:  Get all tasks which are not solved by the request identity
func (rs *TaskResource) MissingIndexHandler(w http.ResponseWriter, r *http.Request) {
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	// check if there is a team
	teamID, err := rs.Stores.Team.TeamID(accessClaims.LoginID, course.ID)
	if err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	var tasks []model.MissingTask
	if teamID.Valid {
		isConfirmed, errConfirm := rs.Stores.Team.Confirmed(teamID.Int64, course.ID)
		if errConfirm != nil || !isConfirmed.Bool {
			// team is not confirmed
			tasks, err = rs.Stores.Task.GetAllMissingTasksForUser(accessClaims.LoginID)
		} else {
			// check if team has missing tasks
			tasks, err = rs.Stores.Task.GetAllMissingTasksForTeam(teamID.Int64)
		}
	} else {
		tasks, err = rs.Stores.Task.GetAllMissingTasksForUser(accessClaims.LoginID)
	}

	// TODO empty list
	if err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	givenRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)

	// render JSON response
	if err = render.RenderList(w, r, newMissingTaskListResponse(tasks, givenRole)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// CreateHandler is public endpoint for
// URL: /courses/{course_id}/sheets/{sheet_id}/tasks
// URLPARAM: course_id,integer
// URLPARAM: sheet_id,integer
// METHOD: post
// TAG: tasks
// REQUEST: TaskRequest
// RESPONSE: 204,TaskResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  create a new task
func (rs *TaskResource) CreateHandler(w http.ResponseWriter, r *http.Request) {

	sheet := r.Context().Value(symbol.CtxKeySheet).(*model.Sheet)

	// start from empty Request
	data := &TaskRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	task := &model.Task{
		Name:               data.Name,
		MaxPoints:          data.MaxPoints,
		PublicDockerImage:  null.StringFrom(data.PublicDockerImage),
		PrivateDockerImage: null.StringFrom(data.PrivateDockerImage),
	}

	// create Task entry in database
	newTask, err := rs.Stores.Task.Create(task, sheet.ID)
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

// GetHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: get
// TAG: tasks
// RESPONSE: 200,TaskResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get a specific task
func (rs *TaskResource) GetHandler(w http.ResponseWriter, r *http.Request) {
	// `Task` is retrieved via middle-ware
	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)

	// render JSON response
	if err := render.Render(w, r, newTaskResponse(task)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusOK)
}

// EditHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: put
// TAG: tasks
// REQUEST: TaskRequest
// RESPONSE: 204,NotContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  edit a specific task
func (rs *TaskResource) EditHandler(w http.ResponseWriter, r *http.Request) {
	// start from empty Request
	data := &TaskRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)
	task.Name = data.Name
	task.MaxPoints = data.MaxPoints
	task.PublicDockerImage = null.StringFrom(data.PublicDockerImage)
	task.PrivateDockerImage = null.StringFrom(data.PrivateDockerImage)

	// update database entry
	if err := rs.Stores.Task.Update(task); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// DeleteHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: delete
// TAG: tasks
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  delete a specific task
func (rs *TaskResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	Task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)

	// update database entry
	if err := rs.Stores.Task.Delete(Task.ID); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// GetPublicTestFileHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}/public_file
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: get
// TAG: tasks
// RESPONSE: 200,ZipFile
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get the zip with the testing framework for the public tests
func (rs *TaskResource) GetPublicTestFileHandler(w http.ResponseWriter, r *http.Request) {

	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)
	hnd := helper.NewPublicTestFileHandle(task.ID)

	if !hnd.Exists() {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := hnd.WriteToBody(w); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
	}

}

// GetPrivateTestFileHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}/private_file
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: get
// TAG: tasks
// RESPONSE: 200,ZipFile
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get the zip with the testing framework for the private tests
func (rs *TaskResource) GetPrivateTestFileHandler(w http.ResponseWriter, r *http.Request) {

	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)
	hnd := helper.NewPrivateTestFileHandle(task.ID)

	if !hnd.Exists() {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := hnd.WriteToBody(w); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
	}

}

// ChangePublicTestFileHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}/public_file
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: post
// TAG: tasks
// REQUEST: Zipfile
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  change the zip with the testing framework for the public tests
func (rs *TaskResource) ChangePublicTestFileHandler(w http.ResponseWriter, r *http.Request) {
	// will always be a POST
	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)

	// the file will be located
	if _, err := helper.NewPublicTestFileHandle(task.ID).WriteToDisk(r, "file_data"); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	render.Status(r, http.StatusOK)
}

// ChangePrivateTestFileHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}/private_file
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: post
// TAG: tasks
// REQUEST: Zipfile
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  change the zip with the testing framework for the private tests
func (rs *TaskResource) ChangePrivateTestFileHandler(w http.ResponseWriter, r *http.Request) {
	// will always be a POST
	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)

	// the file will be located
	if _, err := helper.NewPrivateTestFileHandle(task.ID).WriteToDisk(r, "file_data"); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	render.Status(r, http.StatusOK)
}

// GetSubmissionResultHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}/result
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: get
// TAG: tasks
// RESPONSE: 200,GradeResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  the the public results (grades) for a test and the request identity
func (rs *TaskResource) GetSubmissionResultHandler(w http.ResponseWriter, r *http.Request) {
	givenRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	if givenRole != authorize.STUDENT {
		render.Render(w, r, ErrBadRequestWithDetails(errors.New("This endpoint is ment for students only")))
		return
	}

	// `Task` is retrieved via middle-ware
	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)

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

	// render JSON response
	if err := render.Render(w, r, newGradeResponse(grade, course.ID)); err != nil {
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
		courseFromURL := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

		var taskID int64
		var err error

		// try to get id from URL
		if taskID, err = strconv.ParseInt(chi.URLParam(r, "task_id"), 10, 64); err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// find specific Task in database
		task, err := rs.Stores.Task.Get(taskID)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), symbol.CtxKeyTask, task)

		// find sheet
		sheet, err := rs.Stores.Task.IdentifySheetOfTask(task.ID)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

		if sheetID, err := strconv.ParseInt(chi.URLParam(r, "sheet_id"), 10, 64); err == nil {
			if sheetID != sheet.ID {
				render.Render(w, r, ErrNotFound)
				return
			}
		} else {
			ctx = context.WithValue(ctx, symbol.CtxKeySheet, sheet)
		}

		// public yet?
		if r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole) == authorize.STUDENT && !PublicYet(sheet.PublishAt) {
			render.Render(w, r, ErrBadRequestWithDetails(fmt.Errorf("sheet not published yet")))
			return
		}

		// when there is a taskID in the url, there is NOT a courseID in the url,
		// BUT: when there is a task, there is a course

		course, err := rs.Stores.Task.IdentifyCourseOfTask(task.ID)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

		if courseFromURL.ID != course.ID {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx = context.WithValue(ctx, symbol.CtxKeyCourse, course)

		// serve next
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
