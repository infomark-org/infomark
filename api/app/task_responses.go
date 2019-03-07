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

	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/go-chi/render"
)

// .............................................................................

// TaskResponse is the response payload for Task management.
type TaskResponse struct {
	ID                 int64  `json:"id"`
	MaxPoints          int    `json:"max_points"`
	PublicDockerImage  string `json:"public_docker_image"`
	PrivateDockerImage string `json:"private_docker_image"`
}

// newTaskResponse creates a response from a Task model.
func newTaskResponse(p *model.Task) *TaskResponse {
	return &TaskResponse{
		ID:                 p.ID,
		MaxPoints:          p.MaxPoints,
		PublicDockerImage:  p.PublicDockerImage,
		PrivateDockerImage: p.PrivateDockerImage,
	}
}

// Render post-processes a TaskResponse.
func (body *TaskResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newTaskListResponse creates a response from a list of Task models.
func newTaskListResponse(Tasks []model.Task) []render.Renderer {
	// https://stackoverflow.com/a/36463641/7443104
	list := []render.Renderer{}
	for k := range Tasks {
		list = append(list, newTaskResponse(&Tasks[k]))
	}
	return list
}

// TaskResponse is the response payload for Task management.
type MissingTaskResponse struct {
	Task *struct {
		ID                 int64  `json:"id"`
		MaxPoints          int    `json:"max_points"`
		PublicDockerImage  string `json:"public_docker_image"`
		PrivateDockerImage string `json:"private_docker_image"`
	} `json:"task"`
	CourseID int64 `json:"course_id"`
	SheetID  int64 `json:"sheet_id"`
}

// newTaskResponse creates a response from a Task model.
func newMissingTaskResponse(p *model.MissingTask) *MissingTaskResponse {

	task := struct {
		ID                 int64  `json:"id"`
		MaxPoints          int    `json:"max_points"`
		PublicDockerImage  string `json:"public_docker_image"`
		PrivateDockerImage string `json:"private_docker_image"`
	}{
		p.ID,
		p.MaxPoints,
		p.PublicDockerImage,
		p.PrivateDockerImage,
	}

	r := &MissingTaskResponse{
		Task:     &task,
		CourseID: p.CourseID,
		SheetID:  p.SheetID,
	}

	return r

}

// Render post-processes a TaskResponse.
func (body *MissingTaskResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newTaskListResponse creates a response from a list of Task models.
func newMissingTaskListResponse(Tasks []model.MissingTask) []render.Renderer {
	// https://stackoverflow.com/a/36463641/7443104
	list := []render.Renderer{}
	for k := range Tasks {
		list = append(list, newMissingTaskResponse(&Tasks[k]))
	}
	return list
}
