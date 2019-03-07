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
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-chi/chi"
  "github.com/go-chi/render"
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

// IndexHandler is the enpoint for retrieving all Materials if claim.root is true.
func (rs *MaterialResource) IndexHandler(w http.ResponseWriter, r *http.Request) {

  var materials []model.Material
  var err error
  // we use middle to detect whether there is a course given

  course := r.Context().Value("course").(*model.Course)
  materials, err = rs.Stores.Material.MaterialsOfCourse(course.ID, false)

  if err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  // render JSON reponse
  if err = render.RenderList(w, r, rs.newMaterialListResponse(materials)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// CreateHandler is the enpoint for retrieving all Materials if claim.root is true.
func (rs *MaterialResource) CreateHandler(w http.ResponseWriter, r *http.Request) {

  course := r.Context().Value("course").(*model.Course)

  // start from empty Request
  data := &MaterialRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  material := &model.Material{
    Name:      data.Name,
    Kind:      data.Kind,
    Filename:  data.Filename,
    PublishAt: data.PublishAt,
    LectureAt: data.LectureAt,
  }

  // create Material entry in database
  newMaterial, err := rs.Stores.Material.Create(material, course.ID)
  if err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusCreated)

  // return Material information of created entry
  if err := render.Render(w, r, rs.newMaterialResponse(newMaterial)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

}

// GetHandler is the enpoint for retrieving a specific Material.
func (rs *MaterialResource) GetHandler(w http.ResponseWriter, r *http.Request) {
  // `Material` is retrieved via middle-ware
  material := r.Context().Value("material").(*model.Material)

  // render JSON reponse
  if err := render.Render(w, r, rs.newMaterialResponse(material)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// PatchHandler is the endpoint fro updating a specific Material with given id.
func (rs *MaterialResource) EditHandler(w http.ResponseWriter, r *http.Request) {
  // start from empty Request
  data := &MaterialRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  material := r.Context().Value("material").(*model.Material)

  material.Name = data.Name
  material.Kind = data.Kind
  material.Filename = data.Filename
  material.PublishAt = data.PublishAt
  material.LectureAt = data.LectureAt

  // update database entry
  if err := rs.Stores.Material.Update(material); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

func (rs *MaterialResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
  material := r.Context().Value("material").(*model.Material)

  // update database entry
  if err := rs.Stores.Material.Delete(material.ID); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

func (rs *MaterialResource) GetFileHandler(w http.ResponseWriter, r *http.Request) {

  material := r.Context().Value("material").(*model.Material)
  hnd := helper.NewMaterialFileHandle(material.ID)

  if !hnd.Exists() {
    render.Render(w, r, ErrNotFound)
    return
  } else {
    if err := hnd.WriteToBody(w); err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    }
  }
}

func (rs *MaterialResource) ChangeFileHandler(w http.ResponseWriter, r *http.Request) {
  // will always be a POST
  material := r.Context().Value("material").(*model.Material)

  // the file will be located
  if err := helper.NewMaterialFileHandle(material.ID).WriteToDisk(r, "file_data"); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
  }
  render.Status(r, http.StatusOK)
}

// .............................................................................
// Context middleware is used to load an Material object from
// the URL parameter `MaterialID` passed through as the request. In case
// the Material could not be found, we stop here and return a 404.
// We do NOT check whether the Material is authorized to get this Material.
func (rs *MaterialResource) Context(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // TODO: check permission if inquirer of request is allowed to access this Material
    // Should be done via another middleware
    var materialID int64
    var err error

    // try to get id from URL
    if materialID, err = strconv.ParseInt(chi.URLParam(r, "materialID"), 10, 64); err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // find specific Material in database
    material, err := rs.Stores.Material.Get(materialID)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    ctx := context.WithValue(r.Context(), "material", material)

    // when there is a sheetID in the url, there is NOT a courseID in the url,
    // BUT: when there is a material, there is a course

    course, err := rs.Stores.Material.IdentifyCourseOfMaterial(material.ID)
    if err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
      return
    }

    ctx = context.WithValue(ctx, "course", course)

    // serve next
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
