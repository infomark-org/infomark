// InfoMark - a platform for managing Sheets with
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
  "errors"
  "net/http"
  "strconv"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/model"
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

// .............................................................................

// SheetRequest is the request payload for Sheet management.
type SheetRequest struct {
  *model.Sheet
  ProtectedID int64 `json:"id"`
}

// SheetResponse is the response payload for Sheet management.
type SheetResponse struct {
  *model.Sheet
  Tasks []model.Task `json:"tasks"`
}

// newSheetResponse creates a response from a Sheet model.
func (rs *SheetResource) newSheetResponse(p *model.Sheet) *SheetResponse {
  return &SheetResponse{
    Sheet: p,
  }
}

// newSheetListResponse creates a response from a list of Sheet models.
func (rs *SheetResource) newSheetListResponse(Sheets []model.Sheet) []render.Renderer {
  // https://stackoverflow.com/a/36463641/7443104
  list := []render.Renderer{}
  for k := range Sheets {
    list = append(list, rs.newSheetResponse(&Sheets[k]))
  }

  return list
}

// Bind preprocesses a SheetRequest.
func (body *SheetRequest) Bind(r *http.Request) error {

  if body.Sheet == nil {
    return errors.New("Empty body")
  }

  // Sending the id via request-body is invalid.
  // The id should be submitted in the url.
  body.ProtectedID = 0

  return nil

}

// Render post-processes a SheetResponse.
func (body *SheetResponse) Render(w http.ResponseWriter, r *http.Request) error {
  return nil
}

// IndexHandler is the enpoint for retrieving all Sheets if claim.root is true.
func (rs *SheetResource) IndexHandler(w http.ResponseWriter, r *http.Request) {

  var sheets []model.Sheet
  var err error
  // we use middle to detect whether there is a course given

  course := r.Context().Value("course").(*model.Course)
  sheets, err = rs.Stores.Sheet.SheetsOfCourse(course.ID, false)

  // render JSON reponse
  if err = render.RenderList(w, r, rs.newSheetListResponse(sheets)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }
}

// CreateHandler is the enpoint for retrieving all Sheets if claim.root is true.
func (rs *SheetResource) CreateHandler(w http.ResponseWriter, r *http.Request) {

  course := r.Context().Value("course").(*model.Course)

  // start from empty Request
  data := &SheetRequest{}

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // validate final model
  if err := data.Sheet.Validate(); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // create Sheet entry in database
  newSheet, err := rs.Stores.Sheet.Create(data.Sheet, course.ID)
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

// GetHandler is the enpoint for retrieving a specific Sheet.
func (rs *SheetResource) GetHandler(w http.ResponseWriter, r *http.Request) {
  // `Sheet` is retrieved via middle-ware
  Sheet := r.Context().Value("sheet").(*model.Sheet)

  // render JSON reponse
  if err := render.Render(w, r, rs.newSheetResponse(Sheet)); err != nil {
    render.Render(w, r, ErrRender(err))
    return
  }

  render.Status(r, http.StatusOK)
}

// PatchHandler is the endpoint fro updating a specific Sheet with given id.
func (rs *SheetResource) EditHandler(w http.ResponseWriter, r *http.Request) {
  // start from empty Request
  data := &SheetRequest{
    Sheet: r.Context().Value("sheet").(*model.Sheet),
  }

  // parse JSON request into struct
  if err := render.Bind(r, data); err != nil {
    render.Render(w, r, ErrBadRequestWithDetails(err))
    return
  }

  // update database entry
  if err := rs.Stores.Sheet.Update(data.Sheet); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

func (rs *SheetResource) DeleteHandler(w http.ResponseWriter, r *http.Request) {
  Sheet := r.Context().Value("sheet").(*model.Sheet)

  // update database entry
  if err := rs.Stores.Sheet.Delete(Sheet.ID); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    return
  }

  render.Status(r, http.StatusNoContent)
}

func (rs *SheetResource) GetFileHandler(w http.ResponseWriter, r *http.Request) {

  sheet := r.Context().Value("sheet").(*model.Sheet)
  hnd := helper.NewSheetFileHandle(sheet.ID)

  if !hnd.Exists() {
    render.Render(w, r, ErrNotFound)
    return
  } else {
    if err := hnd.WriteToBody(w); err != nil {
      render.Render(w, r, ErrInternalServerErrorWithDetails(err))
    }
  }
}

func (rs *SheetResource) ChangeFileHandler(w http.ResponseWriter, r *http.Request) {
  // will always be a POST
  sheet := r.Context().Value("sheet").(*model.Sheet)

  // the file will be located
  if err := helper.NewSheetFileHandle(sheet.ID).WriteToDisk(r, "file_data"); err != nil {
    render.Render(w, r, ErrInternalServerErrorWithDetails(err))
  }
  render.Status(r, http.StatusOK)
}

// .............................................................................
// Context middleware is used to load an Sheet object from
// the URL parameter `SheetID` passed through as the request. In case
// the Sheet could not be found, we stop here and return a 404.
// We do NOT check whether the Sheet is authorized to get this Sheet.
func (d *SheetResource) Context(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // TODO: check permission if inquirer of request is allowed to access this Sheet
    // Should be done via another middleware
    var Sheet_id int64
    var err error

    // try to get id from URL
    if Sheet_id, err = strconv.ParseInt(chi.URLParam(r, "sheetID"), 10, 64); err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // find specific Sheet in database
    Sheet, err := d.Stores.Sheet.Get(Sheet_id)
    if err != nil {
      render.Render(w, r, ErrNotFound)
      return
    }

    // serve next
    ctx := context.WithValue(r.Context(), "sheet", Sheet)
    next.ServeHTTP(w, r.WithContext(ctx))
  })
}
