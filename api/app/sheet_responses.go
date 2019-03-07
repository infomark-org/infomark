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
	"time"

	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/go-chi/render"
)

// SheetResponse is the response payload for Sheet management.
type SheetResponse struct {
	ID        int64     `json:"id" example:"13"`
	Name      string    `json:"name" example:"Blatt 0"`
	PublishAt time.Time `json:"publish_at"`
	DueAt     time.Time `json:"due_at"`
}

// Render post-processes a SheetResponse.
func (body *SheetResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newSheetResponse creates a response from a Sheet model.
func (rs *SheetResource) newSheetResponse(p *model.Sheet) *SheetResponse {
	return &SheetResponse{
		ID:        p.ID,
		Name:      p.Name,
		PublishAt: p.PublishAt,
		DueAt:     p.DueAt,
	}
}

// newSheetListResponse creates a response from a list of Sheet models.
func (rs *SheetResource) newSheetListResponse(Sheets []model.Sheet) []render.Renderer {
	list := []render.Renderer{}
	for k := range Sheets {
		list = append(list, rs.newSheetResponse(&Sheets[k]))
	}

	return list
}

type TaskPointsResponse struct {
	TaskPoints *model.TaskPoints `json:"task_points"`
}

func (body *TaskPointsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func newTaskPointsResponse(p *model.TaskPoints) *TaskPointsResponse {
	return &TaskPointsResponse{
		TaskPoints: p,
	}
}

// newCourseListResponse creates a response from a list of course models.
func newTaskPointsListResponse(collection []model.TaskPoints) []render.Renderer {
	list := []render.Renderer{}
	for k := range collection {
		list = append(list, newTaskPointsResponse(&collection[k]))
	}

	return list
}
