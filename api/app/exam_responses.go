// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  Infomark Authors
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

	"github.com/go-chi/render"
	"github.com/infomark-org/infomark/model"
)

// ExamResponse is the response payload for course management.
type ExamResponse struct {
	ID          int64     `json:"id" example:"1"`
	Name        string    `json:"name" example:"Info2"`
	Description string    `json:"description" example:"Some course description here"`
	ExamTime    time.Time `json:"exam_time" example:"auto"`
	CourseID    int64     `json:"course_id" example:"1"`
}

// Render post-processes a ExamResponse.
func (body *ExamResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newExamResponse creates a response from a course model.
func (rs *ExamResource) newExamResponse(p *model.Exam) *ExamResponse {
	return &ExamResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		ExamTime:    p.ExamTime,
		CourseID:    p.CourseID,
	}
}

// newExamListResponse creates a response from a list of course models.
func (rs *ExamResource) newExamListResponse(courses []model.Exam) []render.Renderer {
	list := []render.Renderer{}
	for k := range courses {
		list = append(list, rs.newExamResponse(&courses[k]))
	}
	return list
}

// ExamEnrollmentResponse is the response payload for course management.
type ExamEnrollmentResponse struct {
	Status   int    `json:"status" example:"1"`
	Mark     string `json:"mark" example:"1"`
	UserID   int64  `json:"user_id" example:"42"`
	CourseID int64  `json:"course_id" example:"1"`
	ExamID   int64  `json:"exam_id" example:"1"`
}

// Render post-processes a ExamEnrollmentResponse.
func (body *ExamEnrollmentResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// newExamEnrollmentResponse creates a response from a course model.
func newExamEnrollmentResponse(p *model.UserExam) *ExamEnrollmentResponse {

	return &ExamEnrollmentResponse{
		Status:   p.Status,
		Mark:     p.Mark,
		UserID:   p.UserID,
		CourseID: p.CourseID,
		ExamID:   p.ExamID,
	}
}

func newExamEnrollmentListResponse(enrollments []model.UserExam) []render.Renderer {
	list := []render.Renderer{}
	for k := range enrollments {
		list = append(list, newExamEnrollmentResponse(&enrollments[k]))
	}

	return list
}
