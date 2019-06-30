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

// CourseResponse is the response payload for course management.
type CourseResponse struct {
	ID                 int64     `json:"id" example:"1"`
	Name               string    `json:"name" example:"Info2"`
	Description        string    `json:"description" example:"Some course description here"`
	BeginsAt           time.Time `json:"begins_at" example:"auto"`
	EndsAt             time.Time `json:"ends_at" example:"auto"`
	RequiredPercentage int       `json:"required_percentage" example:"80"`
}

// Render post-processes a CourseResponse.
func (body *CourseResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newCourseResponse creates a response from a course model.
func (rs *CourseResource) newCourseResponse(p *model.Course) *CourseResponse {
	return &CourseResponse{
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
	list := []render.Renderer{}
	for k := range courses {
		list = append(list, rs.newCourseResponse(&courses[k]))
	}
	return list
}

// SheetPointsResponse is reponse for performance on a specific exercise sheet
type SheetPointsResponse struct {
	AquiredPoints int `json:"acquired_points" example:"58"`
	MaxPoints     int `json:"max_points" example:"90"`
	SheetID       int `json:"sheet_id" example:"2"`
}

// Render postprocesses a SheetPointsResponse before marshalling to JSON.
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
type GroupBidsResponse struct {
	ID      int64 `json:"id" example:"512"`
	UserID  int64 `json:"user_id" example:"112"`
	GroupID int64 `json:"group_id" example:"2"`
	Bid     int   `json:"bid" example:"6" minval:"0" maxval:"10"`
}

// Render postprocesses a GroupBidsResponse before marshalling to JSON.
func (body *GroupBidsResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newCourseResponse creates a response from a course model.
func newGroupBidsResponse(p *model.GroupBid) *GroupBidsResponse {
	return &GroupBidsResponse{
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

// CourseResponse is the response payload for course management.
type EnrollmentResponse struct {
	Role int64 `json:"role" example:"1"`
	User *struct {
		ID            int64       `json:"id" example:"13"`
		FirstName     string      `json:"first_name" example:"Max"`
		LastName      string      `json:"last_name" example:"Mustermensch"`
		AvatarURL     null.String `json:"avatar_url" example:"/example.com/file"`
		Email         string      `json:"email" example:"test@uni-tuebingen.de"`
		StudentNumber string      `json:"student_number" example:"0816"`
		Semester      int         `json:"semester" example:"8" minval:"1"`
		Subject       string      `json:"subject" example:"informatik"`
		Language      string      `json:"language" example:"de" len:"2"`
	} `json:"user"`
}

// Render post-processes a CourseResponse.
func (body *EnrollmentResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newCourseResponse creates a response from a course model.
func newEnrollmentResponse(p *model.UserCourse) *EnrollmentResponse {

	user := struct {
		ID            int64       `json:"id" example:"13"`
		FirstName     string      `json:"first_name" example:"Max"`
		LastName      string      `json:"last_name" example:"Mustermensch"`
		AvatarURL     null.String `json:"avatar_url" example:"/example.com/file"`
		Email         string      `json:"email" example:"test@uni-tuebingen.de"`
		StudentNumber string      `json:"student_number" example:"0816"`
		Semester      int         `json:"semester" example:"8" minval:"1"`
		Subject       string      `json:"subject" example:"informatik"`
		Language      string      `json:"language" example:"de" len:"2"`
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

	return &EnrollmentResponse{
		Role: p.Role,
		User: &user,
	}
}

func newEnrollmentListResponse(enrollments []model.UserCourse) []render.Renderer {
	list := []render.Renderer{}
	for k := range enrollments {
		list = append(list, newEnrollmentResponse(&enrollments[k]))
	}

	return list
}
