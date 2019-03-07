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
	null "gopkg.in/guregu/null.v3"
)

// courseResponse is the response payload for course management.
type courseResponse struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	BeginsAt           time.Time `json:"begins_at"`
	EndsAt             time.Time `json:"ends_at"`
	RequiredPercentage int       `json:"required_percentage"`
}

// Render post-processes a courseResponse.
func (body *courseResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newCourseResponse creates a response from a course model.
func (rs *CourseResource) newCourseResponse(p *model.Course) *courseResponse {
	return &courseResponse{
		ID:                 p.ID,
		Name:               p.Name,
		Description:        p.Description,
		BeginsAt:           p.BeginsAt,
		EndsAt:             p.EndsAt,
		RequiredPercentage: p.RequiredPercentage,
	}
}

// newCourseListResponse creates a response from a list of course models.
func (rs *CourseResource) newCourseListResponse(courses []model.Course) []render.Renderer {
	// https://stackoverflow.com/a/36463641/7443104
	list := []render.Renderer{}
	for k := range courses {
		list = append(list, rs.newCourseResponse(&courses[k]))
	}
	return list
}

type SheetPointsResponse struct {
	AquiredPoints int `json:"acquired_points"`
	MaxPoints     int `json:"max_points"`
	SheetID       int `json:"sheet_id"`
}

func (body *SheetPointsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func newSheetPointsResponse(p *model.SheetPoints) *SheetPointsResponse {
	return &SheetPointsResponse{
		AquiredPoints: p.AquiredPoints,
		MaxPoints:     p.MaxPoints,
		SheetID:       p.SheetID,
	}
}

// newCourseListResponse creates a response from a list of course models.
func newSheetPointsListResponse(collection []model.SheetPoints) []render.Renderer {
	list := []render.Renderer{}
	for k := range collection {
		list = append(list, newSheetPointsResponse(&collection[k]))
	}

	return list
}

// .............................................................................
type groupBidsResponse struct {
	ID      int64 `json:"id"`
	UserID  int64 `json:"user_id"`
	GroupID int64 `json:"group_id"`
	Bid     int   `json:"bid"`
}

// Render post-processes a groupBidsResponse.
func (body *groupBidsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newCourseResponse creates a response from a course model.
func newGroupBidsResponse(p *model.GroupBid) *groupBidsResponse {
	return &groupBidsResponse{
		ID:      p.ID,
		UserID:  p.UserID,
		GroupID: p.GroupID,
		Bid:     p.Bid,
	}
}

func newGroupBidsListResponse(collection []model.GroupBid) []render.Renderer {
	list := []render.Renderer{}
	for k := range collection {
		list = append(list, newGroupBidsResponse(&collection[k]))
	}

	return list
}

// .............................................................................

// courseResponse is the response payload for course management.
type enrollmentResponse struct {
	Role int64 `json:"role"`
	User *struct {
		ID            int64       `json:"id"`
		FirstName     string      `json:"first_name"`
		LastName      string      `json:"last_name"`
		AvatarURL     null.String `json:"avatar_url"`
		Email         string      `json:"email"`
		StudentNumber string      `json:"student_number"`
		Semester      int         `json:"semester"`
		Subject       string      `json:"subject"`
		Language      string      `json:"language"`
	} `json:"user"`
}

// Render post-processes a courseResponse.
func (body *enrollmentResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newCourseResponse creates a response from a course model.
func (rs *CourseResource) newEnrollmentResponse(p *model.UserCourse) *enrollmentResponse {

	user := struct {
		ID            int64       `json:"id"`
		FirstName     string      `json:"first_name"`
		LastName      string      `json:"last_name"`
		AvatarURL     null.String `json:"avatar_url"`
		Email         string      `json:"email"`
		StudentNumber string      `json:"student_number"`
		Semester      int         `json:"semester"`
		Subject       string      `json:"subject"`
		Language      string      `json:"language"`
	}{
		ID:            p.ID,
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		AvatarURL:     p.AvatarURL,
		Email:         p.Email,
		StudentNumber: p.StudentNumber,
		Semester:      p.Semester,
		Subject:       p.Subject,
		Language:      p.Language,
	}

	return &enrollmentResponse{
		Role: p.Role,
		User: &user,
	}
}

func (rs *CourseResource) newEnrollmentListResponse(enrollments []model.UserCourse) []render.Renderer {
	list := []render.Renderer{}
	for k := range enrollments {
		list = append(list, rs.newEnrollmentResponse(&enrollments[k]))
	}

	return list
}
