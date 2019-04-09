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
	"time"

	"github.com/cgtuebingen/infomark-backend/auth/authorize"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/go-chi/render"
	"github.com/spf13/viper"
)

// MaterialResponse is the response payload for Material management.
type MaterialResponse struct {
	ID        int64     `json:"id" example:"55"`
	Name      string    `json:"name" example:"Schleifen und Bedingungen"`
	FileURL   string    `json:"file_url" example:"/api/v1/materials/55/file"`
	Kind      int       `json:"kind" example:"0"`
	PublishAt time.Time `json:"publish_at" example:"auto"`
	LectureAt time.Time `json:"lecture_at" example:"auto"`
}

// newMaterialResponse creates a response from a Material model.
func (rs *MaterialResource) newMaterialResponse(p *model.Material, courseID int64) *MaterialResponse {
	return &MaterialResponse{
		ID:        p.ID,
		Name:      p.Name,
		Kind:      p.Kind,
		PublishAt: p.PublishAt,
		LectureAt: p.LectureAt,
		FileURL: fmt.Sprintf("%s/api/v1/courses/%d/materials/%d/file",
			viper.GetString("url"),
			courseID,
			p.ID,
		),
	}
}

// newMaterialListResponse creates a response from a list of Material models.
func (rs *MaterialResource) newMaterialListResponse(givenRole authorize.CourseRole, courseID int64, Materials []model.Material) []render.Renderer {
	list := []render.Renderer{}
	for k := range Materials {
		if givenRole == authorize.STUDENT && !PublicYet(Materials[k].PublishAt) {
			continue
		}
		list = append(list, rs.newMaterialResponse(&Materials[k], courseID))

	}
	return list
}

// Render post-processes a MaterialResponse.
func (body *MaterialResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
