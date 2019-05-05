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
	"fmt"
	"net/http"
	"strconv"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/auth/authorize"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/cgtuebingen/infomark-backend/symbol"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// SheetResource specifies Sheet management handler.
type SheetResource struct {
	Stores *Stores
}

// NewSheetResource create and returns a SheetResource.
func NewSheetResource(stores *Stores) *SheetResource {
	return &SheetResource{
		Stores: stores,
	}
}

// IndexHandler is public endpoint for
// URL: /courses/{course_id}/sheets
// URLPARAM: course_id,integer
// METHOD: get
// TAG: sheets
// RESPONSE: 200,SheetResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get all sheets in course
// DESCRIPTION:
// The sheets are ordered by their names
func (rs *SheetResource) IndexHandler(w http.ResponseWriter, r *http.Request) {

	var sheets []model.Sheet
	var err error
	// we use middle to detect whether there is a course given

	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	sheets, err = rs.Stores.Sheet.SheetsOfCourse(course.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	givenRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)

	// render JSON reponse
	if err = render.RenderList(w, r, rs.newSheetListResponse(givenRole, sheets)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// CreateHandler is public endpoint for
// URL: /courses/{course_id}/sheets
// URLPARAM: course_id,integer
// METHOD: post
// TAG: sheets
// REQUEST: SheetRequest
// RESPONSE: 204,SheetResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  create a new sheet
func (rs *SheetResource) CreateHandler(w http.ResponseWriter, r *http.Request) {

	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

	// start from empty Request
	data := &SheetRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	sheet := &model.Sheet{
		Name:      data.Name,
		PublishAt: data.PublishAt,
		DueAt:     data.DueAt,
	}

	// create Sheet entry in database
	newSheet, err := rs.Stores.Sheet.Create(sheet, course.ID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusCreated)

	// return Sheet information of created entry
	if err := render.Render(w, r, rs.newSheetResponse(newSheet)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

}

// GetHandler is public endpoint for
// URL: /courses/{course_id}/sheets/{sheet_id}
// URLPARAM: course_id,integer
// URLPARAM: sheet_id,integer
// METHOD: get
// TAG: sheets
// RESPONSE: 200,SheetResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get a specific sheet
func (rs *SheetResource) GetHandler(w http.ResponseWriter, r *http.Request) {
	// `Sheet` is retrieved via middle-ware
	Sheet := r.Context().Value(symbol.CtxKeySheet).(*model.Sheet)

	// render JSON reponse
	if err := render.Render(w, r, rs.newSheetResponse(Sheet)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusOK)
}

// EditHandler is public endpoint for
// URL: /courses/{course_id}/sheets/{sheet_id}
// URLPARAM: course_id,integer
// URLPARAM: sheet_id,integer
// METHOD: put
// TAG: sheets
// REQUEST: SheetRequest
// RESPONSE: 204,NotContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  update a specific sheet
func (rs *SheetResource) EditHandler(w http.ResponseWriter, r *http.Request) {
	sheet := r.Context().Value(symbol.CtxKeySheet).(*model.Sheet)

	// start from empty Request
	data := &SheetRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	sheet.Name = data.Name
	sheet.PublishAt = data.PublishAt
	sheet.DueAt = data.DueAt

	// update database entry
	if err := rs.Stores.Sheet.Update(sheet); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// DeleteHandler is public endpoint for
// URL: /courses/{course_id}/sheets/{sheet_id}
// URLPARAM: course_id,integer
// URLPARAM: sheet_id,integer
// METHOD: delete
// TAG: sheets
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  delete a specific sheet
func (rs *SheetResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	Sheet := r.Context().Value(symbol.CtxKeySheet).(*model.Sheet)

	// update database entry
	if err := rs.Stores.Sheet.Delete(Sheet.ID); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// GetFileHandler is public endpoint for
// URL: /courses/{course_id}/sheets/{sheet_id}/file
// URLPARAM: course_id,integer
// URLPARAM: sheet_id,integer
// METHOD: get
// TAG: sheets
// RESPONSE: 200,ZipFile
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get the zip file of a sheet
func (rs *SheetResource) GetFileHandler(w http.ResponseWriter, r *http.Request) {

	sheet := r.Context().Value(symbol.CtxKeySheet).(*model.Sheet)
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	hnd := helper.NewSheetFileHandle(sheet.ID)

	if !hnd.Exists() {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := hnd.WriteToBodyWithName(fmt.Sprintf("%s-%s.zip", course.Name, sheet.Name), w); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
	}

}

// ChangeFileHandler is public endpoint for
// URL: /courses/{course_id}/sheets/{sheet_id}/file
// URLPARAM: course_id,integer
// URLPARAM: sheet_id,integer
// METHOD: post
// TAG: sheets
// REQUEST: zipfile
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  change the zip file of a sheet
func (rs *SheetResource) ChangeFileHandler(w http.ResponseWriter, r *http.Request) {
	// will always be a POST
	sheet := r.Context().Value(symbol.CtxKeySheet).(*model.Sheet)

	// the file will be located
	if _, err := helper.NewSheetFileHandle(sheet.ID).WriteToDisk(r, "file_data"); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
	}
	render.Status(r, http.StatusOK)
}

// PointsHandler is public endpoint for
// URL: /courses/{course_id}/sheets/{sheet_id}/points
// URLPARAM: course_id,integer
// URLPARAM: sheet_id,integer
// METHOD: get
// TAG: sheets
// RESPONSE: 200,newTaskListResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  return all points from a sheet for the request identity
func (rs *SheetResource) PointsHandler(w http.ResponseWriter, r *http.Request) {
	sheet := r.Context().Value(symbol.CtxKeySheet).(*model.Sheet)
	accessClaims := r.Context().Value(symbol.CtxKeyAccessClaims).(*authenticate.AccessClaims)

	taskPoints, err := rs.Stores.Sheet.PointsForUser(accessClaims.LoginID, sheet.ID)
	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// resp := &SheetPointsResponse{SheetPoints: taskPoints}
	if err := render.RenderList(w, r, newTaskPointsListResponse(taskPoints)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusOK)
}

// .............................................................................

// Context middleware is used to load an Sheet object from
// the URL parameter `SheetID` passed through as the request. In case
// the Sheet could not be found, we stop here and return a 404.
// We do NOT check whether the Sheet is authorized to get this Sheet.
func (rs *SheetResource) Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		courseFromURL := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

		var sheetID int64
		var err error

		// try to get id from URL
		if sheetID, err = strconv.ParseInt(chi.URLParam(r, "sheet_id"), 10, 64); err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// find specific Sheet in database
		sheet, err := rs.Stores.Sheet.Get(sheetID)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// public yet?
		if r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole) == authorize.STUDENT && !PublicYet(sheet.PublishAt) {
			render.Render(w, r, ErrBadRequestWithDetails(fmt.Errorf("sheet not published yet")))
			return
		}

		ctx := context.WithValue(r.Context(), symbol.CtxKeySheet, sheet)

		// when there is a sheetID in the url, there is NOT a courseID in the url,
		// BUT: when there is a sheet, there is a course

		course, err := rs.Stores.Sheet.IdentifyCourseOfSheet(sheet.ID)
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
