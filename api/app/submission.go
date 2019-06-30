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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/api/shared"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/auth/authorize"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/cgtuebingen/infomark-backend/symbol"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/spf13/viper"
)

// SubmissionResource specifies Submission management handler.
type SubmissionResource struct {
	Stores *Stores
}

// NewSubmissionResource create and returns a SubmissionResource.
func NewSubmissionResource(stores *Stores) *SubmissionResource {
	return &SubmissionResource{
		Stores: stores,
	}
}

// GetFileHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}/submission
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: get
// TAG: submissions
// RESPONSE: 200,ZipFile
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get the zip file containing the submission of the request identity for a given task
func (rs *SubmissionResource) GetFileHandler(w http.ResponseWriter, r *http.Request) {
	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)
	// submission := r.Context().Value(symbol.CtxKeySubmission).(*model.Submission)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	givenRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)

	submission, err := rs.Stores.Submission.GetByUserAndTask(accessClaims.LoginID, task.ID)
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	// students can only access their own files
	if submission.UserID != accessClaims.LoginID {
		if givenRole == authorize.STUDENT {
			render.Render(w, r, ErrUnauthorized)
			return
		}
	}

	hnd := helper.NewSubmissionFileHandle(submission.ID)

	if !hnd.Exists() {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := hnd.WriteToBody(w); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

}

// GetCollectionHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}/groups/{group_id}
// URLPARAM: course_id,integer
// URLPARAM: sheet_id,integer
// URLPARAM: task_id,integer
// URLPARAM: group_id,integer
// METHOD: get
// TAG: submissions
// RESPONSE: 200,ZipFile
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get the path to the zip file containing all submissions for a given task and a given group if exists
func (rs *SubmissionResource) GetCollectionHandler(w http.ResponseWriter, r *http.Request) {
	givenRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)

	if givenRole == authorize.STUDENT {
		render.Render(w, r, ErrUnauthorized)
		return
	}

	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)

	var groupID int64
	var err error

	// try to get id from URL
	if groupID, err = strconv.ParseInt(chi.URLParam(r, "group_id"), 10, 64); err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	sheet, err := rs.Stores.Task.IdentifySheetOfTask(task.ID)
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	// find specific group in database
	group, err := rs.Stores.Group.Get(groupID)
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	hnd := helper.NewSubmissionsCollectionFileHandle(course.ID, sheet.ID, task.ID, group.ID)

	text := ""
	if hnd.Exists() {
		text = fmt.Sprintf("%s/courses/%d/tasks/%d/groups/%d/file",
			viper.GetString("url"),
			course.ID, task.ID, group.ID,
		)
	}
	if err := render.Render(w, r, newRawResponse(text)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// GetCollectionFileHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}/groups/{group_id}/file
// URLPARAM: course_id,integer
// URLPARAM: sheet_id,integer
// URLPARAM: task_id,integer
// URLPARAM: group_id,integer
// METHOD: get
// TAG: submissions
// RESPONSE: 200,ZipFile
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get the zip file containing all submissions for a given task and a given group
func (rs *SubmissionResource) GetCollectionFileHandler(w http.ResponseWriter, r *http.Request) {
	givenRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)

	if givenRole == authorize.STUDENT {
		render.Render(w, r, ErrUnauthorized)
		return
	}

	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)

	var groupID int64
	var err error

	// try to get id from URL
	if groupID, err = strconv.ParseInt(chi.URLParam(r, "group_id"), 10, 64); err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	sheet, err := rs.Stores.Task.IdentifySheetOfTask(task.ID)
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	// find specific group in database
	group, err := rs.Stores.Group.Get(groupID)
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	hnd := helper.NewSubmissionsCollectionFileHandle(course.ID, sheet.ID, task.ID, group.ID)

	if !hnd.Exists() {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := hnd.WriteToBody(w); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
}

// GetFileByIDHandler is public endpoint for
// URL: /courses/{course_id}/submissions/{submission_id}/file
// URLPARAM: course_id,integer
// URLPARAM: submission_id,integer
// METHOD: get
// TAG: submissions
// RESPONSE: 200,ZipFile
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get the zip file of a specific submission
func (rs *SubmissionResource) GetFileByIDHandler(w http.ResponseWriter, r *http.Request) {

	submission := r.Context().Value(symbol.CtxKeySubmission).(*model.Submission)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)
	givenRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)

	submission, err := rs.Stores.Submission.Get(submission.ID)
	if err != nil {
		render.Render(w, r, ErrNotFound)
		return
	}

	// students can only access their own files
	if submission.UserID != accessClaims.LoginID {
		if givenRole == authorize.STUDENT {
			render.Render(w, r, ErrUnauthorized)
			return
		}
	}

	hnd := helper.NewSubmissionFileHandle(submission.ID)

	if !hnd.Exists() {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := hnd.WriteToBody(w); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

}

// UploadFileHandler is public endpoint for
// URL: /courses/{course_id}/tasks/{task_id}/submission
// URLPARAM: course_id,integer
// URLPARAM: task_id,integer
// METHOD: post
// TAG: submissions
// REQUEST: zipfile
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  changes the zip file of a submission belonging to the request identity
func (rs *SubmissionResource) UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	task := r.Context().Value(symbol.CtxKeyTask).(*model.Task)
	sheet := r.Context().Value(symbol.CtxKeySheet).(*model.Sheet)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)

	course_role := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)

	if course_role == authorize.STUDENT && !PublicYet(sheet.PublishAt) {
		render.Render(w, r, ErrBadRequestWithDetails(fmt.Errorf("sheet not published yet")))
		return
	}

	if course_role == authorize.STUDENT && OverTime(sheet.DueAt) {
		render.Render(w, r, ErrBadRequestWithDetails(fmt.Errorf("too late deadline was %v but now it is %v", sheet.DueAt, NowUTC())))
		return
	}

	usedUserID := accessClaims.LoginID
	if r.FormValue("user_id") != "" && course_role == authorize.ADMIN {
		// admins cannot upload solutions for students even after the deadline
		requested_user_id, _ := strconv.Atoi(r.FormValue("user_id"))
		usedUserID = int64(requested_user_id)
	}

	var grade *model.Grade

	defaultPublicTestLog := "submission received and will be tested"
	if !task.PublicDockerImage.Valid || !helper.NewPublicTestFileHandle(task.ID).Exists() {
		//  .Valid == true iff not null
		defaultPublicTestLog = "no unit tests for this task are available"
	}

	defaultPrivateTestLog := "submission received and will be tested"
	if !task.PrivateDockerImage.Valid || !helper.NewPrivateTestFileHandle(task.ID).Exists() {
		//  .Valid == true iff not null
		defaultPrivateTestLog = "no unit tests for this task are available"
	}

	// create submisison if not exists
	submission, err := rs.Stores.Submission.GetByUserAndTask(usedUserID, task.ID)
	if err != nil {
		// no such submission
		submission, err = rs.Stores.Submission.Create(&model.Submission{UserID: usedUserID, TaskID: task.ID})
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

		// create also empty grade, which will be filled in later
		grade = &model.Grade{
			PublicExecutionState:  0,
			PrivateExecutionState: 0,
			PublicTestLog:         defaultPublicTestLog,
			PrivateTestLog:        defaultPrivateTestLog,
			PublicTestStatus:      0,
			PrivateTestStatus:     0,
			AcquiredPoints:        0,
			Feedback:              "",
			TutorID:               1,
			SubmissionID:          submission.ID,
		}

		// fetch id from grade as we need it
		grade, err = rs.Stores.Grade.Create(grade)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

	} else {
		// submission exists, we only need to get the grade
		grade, err = rs.Stores.Grade.GetForSubmission(submission.ID)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

		// and update the grade
		grade.PublicExecutionState = 0
		grade.PrivateExecutionState = 0
		grade.PublicTestLog = defaultPublicTestLog
		grade.PrivateTestLog = defaultPrivateTestLog

		err = rs.Stores.Grade.Update(grade)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

	}

	// the file will be located
	if _, err := helper.NewSubmissionFileHandle(submission.ID).WriteToDisk(r, "file_data"); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	sha256, err := helper.NewSubmissionFileHandle(submission.ID).Sha256()
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// enqueue file into testing queue
	// By definition user with id 1 is the system itself with root access
	tokenManager, err := authenticate.NewTokenAuth()
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}
	accessToken, err := tokenManager.CreateAccessJWT(
		authenticate.NewAccessClaims(1, true))
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	if task.PublicDockerImage.Valid && helper.NewPublicTestFileHandle(task.ID).Exists() {
		// enqueue public test

		request := shared.NewSubmissionAMQPWorkerRequest(
			course.ID, task.ID, submission.ID, grade.ID,
			accessToken, viper.GetString("url"), task.PublicDockerImage.String, sha256, "public")

		body, err := json.Marshal(request)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

		err = DefaultSubmissionProducer.Publish(body)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}
	} else {
		grade.PublicTestLog = "No public dockerimage was specified --> will not run any public test"
		err = rs.Stores.Grade.Update(grade)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}
	}

	if task.PrivateDockerImage.Valid && helper.NewPrivateTestFileHandle(task.ID).Exists() {
		// enqueue private test

		request := shared.NewSubmissionAMQPWorkerRequest(
			course.ID, task.ID, submission.ID, grade.ID,
			accessToken, viper.GetString("url"), task.PrivateDockerImage.String, sha256, "private")

		body, err := json.Marshal(request)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

		err = DefaultSubmissionProducer.Publish(body)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}
	} else {
		grade.PrivateTestLog = "No private dockerimage was specified --> will not run any private test"
		err = rs.Stores.Grade.Update(grade)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}
	}

	totalSubmissionCounterVec.WithLabelValues(fmt.Sprintf("%d", task.ID)).Inc()

	render.Status(r, http.StatusOK)
}

// IndexHandler is public endpoint for
// URL: /courses/{course_id}/submissions
// URLPARAM: course_id,integer
// QUERYPARAM: sheet_id,integer
// QUERYPARAM: task_id,integer
// QUERYPARAM: group_id,integer
// QUERYPARAM: user_id,integer
// METHOD: get
// TAG: submissions
// RESPONSE: 200,SubmissionResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  Query submissions in a course
func (rs *SubmissionResource) IndexHandler(w http.ResponseWriter, r *http.Request) {
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

	filterGroupID := helper.Int64FromURL(r, "group_id", 0)
	filterUserID := helper.Int64FromURL(r, "user_id", 0)
	filterSheetID := helper.Int64FromURL(r, "sheet_id", 0)
	filterTaskID := helper.Int64FromURL(r, "task_id", 0)

	submissions, err := rs.Stores.Submission.GetFiltered(course.ID, filterGroupID, filterUserID, filterSheetID, filterTaskID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// render JSON reponse
	if err = render.RenderList(w, r, newSubmissionListResponse(submissions, course.ID)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusOK)

}

// .............................................................................

// Context middleware is used to load an Submission object from
// the URL parameter `TaskID` passed through as the request. In case
// the Submission could not be found, we stop here and return a 404.
// We do NOT check whether the identity is authorized to get this Submission.
func (rs *SubmissionResource) Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// course_from_url := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

		var submissionID int64
		var err error

		// try to get id from URL
		if submissionID, err = strconv.ParseInt(chi.URLParam(r, "submission_id"), 10, 64); err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// find specific Submission in database
		submission, err := rs.Stores.Submission.Get(submissionID)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), symbol.CtxKeySubmission, submission)

		// when there is a submissionID in the url, there is NOT a taskID in the url,
		// BUT: when there is a Submission, there is a task
		// BUT: when there is a task, there is a course (other middlewarwe)

		// find specific Task in database
		var task *model.Task
		task, err = rs.Stores.Task.Get(submission.TaskID)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx = context.WithValue(ctx, symbol.CtxKeyTask, task)

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

		// find course
		course, err := rs.Stores.Task.IdentifyCourseOfTask(task.ID)
		if err != nil {
			render.Render(w, r, ErrInternalServerErrorWithDetails(err))
			return
		}

		ctx = context.WithValue(ctx, symbol.CtxKeyCourse, course)

		// serve next
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
