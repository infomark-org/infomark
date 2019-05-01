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

	"github.com/cgtuebingen/infomark-backend/auth/authorize"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/go-chi/render"
	null "gopkg.in/guregu/null.v3"
)

// .............................................................................

// TaskResponse is the response payload for Task management.
type TaskResponse struct {
	ID                 int64       `json:"id" example:"684"`
	Name               string      `json:"name" example:"Task 1"`
	MaxPoints          int         `json:"max_points" example:"23"`
	PublicDockerImage  null.String `json:"public_docker_image" example:"DefaultJavaTestingImage"`
	PrivateDockerImage null.String `json:"private_docker_image" example:"DefaultJavaTestingImage"`
}

// newTaskResponse creates a response from a Task model.
func newTaskResponse(p *model.Task) *TaskResponse {
	return &TaskResponse{
		ID:                 p.ID,
		Name:               p.Name,
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
	list := []render.Renderer{}
	for k := range Tasks {
		list = append(list, newTaskResponse(&Tasks[k]))
	}
	return list
}

// MissingTaskResponse is the response payload for displaying
type MissingTaskResponse struct {
	Task *struct {
		ID                 int64       `json:"id" example:"684"`
		Name               string      `json:"name" example:"Task 1"`
		MaxPoints          int         `json:"max_points" example:"23"`
		PublicDockerImage  null.String `json:"public_docker_image" example:"DefaultJavaTestingImage"`
		PrivateDockerImage null.String `json:"private_docker_image" example:"DefaultJavaTestingImage"`
	} `json:"task"`
	CourseID int64 `json:"course_id" example:"1"`
	SheetID  int64 `json:"sheet_id" example:"8"`
}

// newTaskResponse creates a response from a Task model.
func newMissingTaskResponse(p *model.MissingTask, givenRole authorize.CourseRole) *MissingTaskResponse {

	publicDockerImage := p.PublicDockerImage
	privateDockerImage := p.PrivateDockerImage

	if givenRole == authorize.STUDENT {
		publicDockerImage = null.StringFrom("")
		privateDockerImage = null.StringFrom("")
	}

	task := struct {
		ID                 int64       `json:"id" example:"684"`
		Name               string      `json:"name" example:"Task 1"`
		MaxPoints          int         `json:"max_points" example:"23"`
		PublicDockerImage  null.String `json:"public_docker_image" example:"DefaultJavaTestingImage"`
		PrivateDockerImage null.String `json:"private_docker_image" example:"DefaultJavaTestingImage"`
	}{
		p.ID,
		p.Name,
		p.MaxPoints,
		publicDockerImage,
		privateDockerImage,
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
func newMissingTaskListResponse(Tasks []model.MissingTask, givenRole authorize.CourseRole) []render.Renderer {
	list := []render.Renderer{}
	for k := range Tasks {
		list = append(list, newMissingTaskResponse(&Tasks[k], givenRole))
	}
	return list
}
