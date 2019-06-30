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
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/render"
	"github.com/infomark-org/infomark-backend/auth/authorize"
	"github.com/infomark-org/infomark-backend/model"
)

// SheetResponse is the response payload for Sheet management.
type SheetResponse struct {
	ID        int64     `json:"id" example:"13"`
	Name      string    `json:"name" example:"Blatt 0"`
	FileURL   string    `json:"file_url" example:"/api/v1/sheets/13/file"`
	PublishAt time.Time `json:"publish_at" example:"auto"`
	DueAt     time.Time `json:"due_at" example:"auto"`
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
		FileURL:   fmt.Sprintf("/api/v1/sheets/%s/file", strconv.FormatInt(p.ID, 10)),
	}
}

// newSheetListResponse creates a response from a list of Sheet models.
func (rs *SheetResource) newSheetListResponse(givenRole authorize.CourseRole, Sheets []model.Sheet) []render.Renderer {
	list := []render.Renderer{}
	for k := range Sheets {
		if givenRole == authorize.STUDENT && !PublicYet(Sheets[k].PublishAt) {
			continue
		}
		list = append(list, rs.newSheetResponse(&Sheets[k]))
	}

	return list
}

// TaskPointsResponse returns a performance summary for a task and student
type TaskPointsResponse struct {
	AquiredPoints int `json:"acquired_points" example:"58"`
	MaxPoints     int `json:"max_points" example:"90"`
	TaskID        int `json:"task_id" example:"2"`
}

// Render post-processes a TaskPointsResponse.
func (body *TaskPointsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func newTaskPointsResponse(p *model.TaskPoints) *TaskPointsResponse {
	return &TaskPointsResponse{
		AquiredPoints: p.AquiredPoints,
		MaxPoints:     p.MaxPoints,
		TaskID:        p.TaskID,
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
