// InfoMark - a platform for managing exams with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  Infomark Authors
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

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/infomark-org/infomark-backend/model"
	"github.com/infomark-org/infomark-backend/symbol"
)

// ExamResource specifies exam management handler.
type ExamResource struct {
	Stores *Stores
}

// NewExamResource create and returns a ExamResource.
func NewExamResource(stores *Stores) *ExamResource {
	return &ExamResource{
		Stores: stores,
	}
}

// IndexHandler is public endpoint for
// URL: /courses/{course_id}/exams
// METHOD: get
// TAG: exams
// RESPONSE: 200,ExamResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  list all exams
func (rs *ExamResource) IndexHandler(w http.ResponseWriter, r *http.Request) {
	// fetch collection of exams from database
	exams, err := rs.Stores.Exam.GetAll()
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// render JSON reponse
	if err = render.RenderList(w, r, rs.newExamListResponse(exams)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// CreateHandler is public endpoint for
// URL: /courses/{course_id}/exams
// URLPARAM: course_id,integer
// METHOD: post
// TAG: exams
// REQUEST: examRequest
// RESPONSE: 204,ExamResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  create a new exam
func (rs *ExamResource) CreateHandler(w http.ResponseWriter, r *http.Request) {
	// start from empty Request
	data := &examRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

	exam := &model.Exam{}
	exam.Name = data.Name
	exam.Description = data.Description
	exam.ExamTime = data.ExamTime
	exam.CourseID = course.ID

	// create course entry in database
	newExam, err := rs.Stores.Exam.Create(exam)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusCreated)

	// return course information of created entry
	if err := render.Render(w, r, rs.newExamResponse(newExam)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

}

// GetHandler is public endpoint for
// URL: /courses/{course_id}/exams/{exam_id}
// URLPARAM: course_id,integer
// URLPARAM: exam_id,integer
// METHOD: get
// TAG: exams
// RESPONSE: 200,ExamResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get a specific exam
func (rs *ExamResource) GetHandler(w http.ResponseWriter, r *http.Request) {
	// `Task` is retrieved via middle-ware
	exam := r.Context().Value(symbol.CtxKeyExam).(*model.Exam)

	// render JSON reponse
	if err := render.Render(w, r, rs.newExamResponse(exam)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusOK)
}

// EditHandler is public endpoint for
// URL: /courses/{course_id}/exams/{exam_id}
// URLPARAM: course_id,integer
// URLPARAM: exam_id,integer
// METHOD: put
// TAG: exams
// REQUEST: ExamResponse
// RESPONSE: 204,NotContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  update a specific exam
func (rs *ExamResource) EditHandler(w http.ResponseWriter, r *http.Request) {
	// start from empty Request
	data := &examRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	exam := r.Context().Value(symbol.CtxKeyExam).(*model.Exam)
	exam.Name = data.Name
	exam.Description = data.Description
	exam.ExamTime = data.ExamTime

	// update database entry
	if err := rs.Stores.Exam.Update(exam); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// DeleteHandler is public endpoint for
// URL: /courses/{course_id}/exams/{exam_id}
// URLPARAM: course_id,integer
// URLPARAM: exam_id,integer
// METHOD: delete
// TAG: exams
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  delete a specific exam
func (rs *ExamResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	exam := r.Context().Value(symbol.CtxKeyExam).(*model.Exam)

	// update database entry
	if err := rs.Stores.Exam.Delete(exam.ID); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// Context middleware is used to load an Exam object from
// the URL parameter `examID` passed through as the request. In case
// the Exam could not be found, we stop here and return a 404.
// We do NOT check whether the exam is authorized to get this exam.
func (rs *ExamResource) Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: check permission if inquirer of request is allowed to access this exam
		// Should be done via another middleware
		var examID int64
		var err error

		// try to get id from URL
		if examID, err = strconv.ParseInt(chi.URLParam(r, "exam_id"), 10, 64); err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// find specific exam in database
		exam, err := rs.Stores.Exam.Get(examID)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// serve next
		ctx := context.WithValue(r.Context(), symbol.CtxKeyExam, exam)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
