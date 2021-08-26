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
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/infomark-org/infomark/api/helper"
	"github.com/infomark-org/infomark/auth/authorize"
	"github.com/infomark-org/infomark/model"
	"github.com/infomark-org/infomark/symbol"
)

// MaterialResource specifies Material management handler.
type MaterialResource struct {
	Stores *Stores
}

// NewMaterialResource create and returns a MaterialResource.
func NewMaterialResource(stores *Stores) *MaterialResource {
	return &MaterialResource{
		Stores: stores,
	}
}

// IndexHandler is public endpoint for
// URL: /courses/{course_id}/materials
// URLPARAM: course_id,integer
// METHOD: get
// TAG: materials
// RESPONSE: 200,MaterialResponseList
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get all materials in course
// DESCRIPTION:
// The materials are ordered by the lecture date.
// Kind means 0: slide, 1: supplementary
func (rs *MaterialResource) IndexHandler(w http.ResponseWriter, r *http.Request) {

	var materials []model.Material
	var err error
	// we use middle to detect whether there is a course given
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	givenRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)
	materials, err = rs.Stores.Material.MaterialsOfCourse(course.ID, givenRole.ToInt())

	if err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	// render JSON response
	if err = render.RenderList(w, r, rs.newMaterialListResponse(givenRole, course.ID, materials)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// CreateHandler is public endpoint for
// URL: /courses/{course_id}/materials
// URLPARAM: course_id,integer
// METHOD: post
// TAG: materials
// REQUEST: MaterialRequest
// RESPONSE: 204,MaterialResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  create a new material
// DESCRIPTION:
// Kind means 0: slide, 1: supplementary
func (rs *MaterialResource) CreateHandler(w http.ResponseWriter, r *http.Request) {

	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

	// start from empty Request
	data := &MaterialRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	material := &model.Material{
		Name:         data.Name,
		Kind:         data.Kind,
		PublishAt:    data.PublishAt,
		LectureAt:    data.LectureAt,
		RequiredRole: data.RequiredRole,
	}

	// create Material entry in database
	newMaterial, err := rs.Stores.Material.Create(material, course.ID)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusCreated)

	// return Material information of created entry
	if err := render.Render(w, r, rs.newMaterialResponse(newMaterial, course.ID)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

}

// GetHandler is public endpoint for
// URL: /courses/{course_id}/materials/{material_id}
// URLPARAM: course_id,integer
// URLPARAM: material_id,integer
// METHOD: get
// TAG: materials
// RESPONSE: 200,MaterialResponse
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get a specific material
// DESCRIPTION:
// Kind means 0: slide, 1: supplementary
func (rs *MaterialResource) GetHandler(w http.ResponseWriter, r *http.Request) {
	// `Material` is retrieved via middle-ware
	material := r.Context().Value(symbol.CtxKeyMaterial).(*model.Material)
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)

	givenRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)
	if givenRole == authorize.STUDENT && !PublicYet(material.PublishAt) {
		render.Render(w, r, ErrUnauthorizedWithDetails(fmt.Errorf("Not public yet")))
		return
	}

	if material.RequiredRole > givenRole.ToInt() {
		render.Render(w, r, ErrUnauthorizedWithDetails(fmt.Errorf("no access allowed with your role")))
		return
	}

	// render JSON response
	if err := render.Render(w, r, rs.newMaterialResponse(material, course.ID)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Status(r, http.StatusOK)
}

// EditHandler is public endpoint for
// URL: /courses/{course_id}/materials/{material_id}
// URLPARAM: course_id,integer
// URLPARAM: material_id,integer
// METHOD: put
// TAG: materials
// REQUEST: MaterialRequest
// RESPONSE: 204,NotContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  update a specific material
// DESCRIPTION:
// Kind means 0: slide, 1: supplementary
func (rs *MaterialResource) EditHandler(w http.ResponseWriter, r *http.Request) {
	// start from empty Request
	data := &MaterialRequest{}

	// parse JSON request into struct
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrBadRequestWithDetails(err))
		return
	}

	material := r.Context().Value(symbol.CtxKeyMaterial).(*model.Material)

	material.Name = data.Name
	material.Kind = data.Kind
	material.PublishAt = data.PublishAt
	material.LectureAt = data.LectureAt
	material.RequiredRole = data.RequiredRole

	// update database entry
	if err := rs.Stores.Material.Update(material); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// DeleteHandler is public endpoint for
// URL: /courses/{course_id}/materials/{material_id}
// URLPARAM: course_id,integer
// URLPARAM: material_id,integer
// METHOD: delete
// TAG: materials
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  delete a specific material
func (rs *MaterialResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	material := r.Context().Value(symbol.CtxKeyMaterial).(*model.Material)

	// update database entry
	if err := rs.Stores.Material.Delete(material.ID); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// GetFileHandler is public endpoint for
// URL: /courses/{course_id}/materials/{material_id}/file
// URLPARAM: course_id,integer
// URLPARAM: material_id,integer
// METHOD: get
// TAG: materials
// RESPONSE: 200,ZipFile
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  get the zip file of a material
func (rs *MaterialResource) GetFileHandler(w http.ResponseWriter, r *http.Request) {

	material := r.Context().Value(symbol.CtxKeyMaterial).(*model.Material)
	course := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
	hnd := helper.NewMaterialFileHandle(material.ID)
	if !hnd.Exists() {
		render.Render(w, r, ErrNotFound)
		return
	}

	givenRole := r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole)
	if givenRole == authorize.STUDENT && !PublicYet(material.PublishAt) {
		render.Render(w, r, ErrNotFound)
		return
	}

	if err := hnd.WriteToBodyWithName(fmt.Sprintf("%s-%s", course.Name, material.Filename), w); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
	}
}

// ChangeFileHandler is public endpoint for
// URL: /courses/{course_id}/materials/{material_id}/file
// URLPARAM: course_id,integer
// URLPARAM: material_id,integer
// METHOD: post
// TAG: materials
// REQUEST: Zipfile
// RESPONSE: 204,NoContent
// RESPONSE: 400,BadRequest
// RESPONSE: 401,Unauthenticated
// RESPONSE: 403,Unauthorized
// SUMMARY:  change the zip file of a sheet
// DESCRIPTION:
// This endpoint will only support pdf or zip files.
func (rs *MaterialResource) ChangeFileHandler(w http.ResponseWriter, r *http.Request) {
	// will always be a POST
	material := r.Context().Value(symbol.CtxKeyMaterial).(*model.Material)

	var (
		err      error
		filename string
	)

	// the file will be located
	if filename, err = helper.NewMaterialFileHandle(material.ID).WriteToDisk(r, "file_data"); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
	}

	material.Filename = filename

	if err := rs.Stores.Material.Update(material); err != nil {
		render.Render(w, r, ErrInternalServerErrorWithDetails(err))
		return
	}

	render.Status(r, http.StatusNoContent)
}

// .............................................................................

// Context middleware is used to load an Material object from
// the URL parameter `MaterialID` passed through as the request. In case
// the Material could not be found, we stop here and return a 404.
// We do NOT check whether the Material is authorized to get this Material.
func (rs *MaterialResource) Context(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		courseFromURL := r.Context().Value(symbol.CtxKeyCourse).(*model.Course)
		// Should be done via another middleware
		var materialID int64
		var err error

		// try to get id from URL
		if materialID, err = strconv.ParseInt(chi.URLParam(r, "material_id"), 10, 64); err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// find specific Material in database
		material, err := rs.Stores.Material.Get(materialID)
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		// public yet?
		if r.Context().Value(symbol.CtxKeyCourseRole).(authorize.CourseRole) == authorize.STUDENT && !PublicYet(material.PublishAt) {
			render.Render(w, r, ErrUnauthorizedWithDetails(fmt.Errorf("material not published yet")))
			return
		}

		ctx := context.WithValue(r.Context(), symbol.CtxKeyMaterial, material)

		// when there is a sheetID in the url, there is NOT a courseID in the url,
		// BUT: when there is a material, there is a course

		course, err := rs.Stores.Material.IdentifyCourseOfMaterial(material.ID)
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
