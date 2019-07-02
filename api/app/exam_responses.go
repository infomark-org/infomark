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
	"github.com/infomark-org/infomark-backend/model"
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

// // ExamEnrollmentResponse is the response payload for course management.
// type ExamEnrollmentResponse struct {
// 	Status int `json:"status" example:"1"`
// 	Mark   int `json:"mark" example:"1"`
// 	User   *struct {
// 		ID            int64       `json:"id" example:"13"`
// 		FirstName     string      `json:"first_name" example:"Max"`
// 		LastName      string      `json:"last_name" example:"Mustermensch"`
// 		AvatarURL     null.String `json:"avatar_url" example:"/example.com/file"`
// 		Email         string      `json:"email" example:"test@uni-tuebingen.de"`
// 		StudentNumber string      `json:"student_number" example:"0816"`
// 		Semester      int         `json:"semester" example:"8" minval:"1"`
// 		Subject       string      `json:"subject" example:"informatik"`
// 		Language      string      `json:"language" example:"de" len:"2"`
// 	} `json:"user"`
// }

// // Render post-processes a ExamEnrollmentResponse.
// func (body *ExamEnrollmentResponse) Render(w http.ResponseWriter, r *http.Request) error {
// 	return nil
// }

// // newExamEnrollmentResponse creates a response from a course model.
// func newExamEnrollmentResponse(p *model.UserExamView) *ExamEnrollmentResponse {

// 	user := struct {
// 		ID            int64       `json:"id" example:"13"`
// 		FirstName     string      `json:"first_name" example:"Max"`
// 		LastName      string      `json:"last_name" example:"Mustermensch"`
// 		AvatarURL     null.String `json:"avatar_url" example:"/example.com/file"`
// 		Email         string      `json:"email" example:"test@uni-tuebingen.de"`
// 		StudentNumber string      `json:"student_number" example:"0816"`
// 		Semester      int         `json:"semester" example:"8" minval:"1"`
// 		Subject       string      `json:"subject" example:"informatik"`
// 		Language      string      `json:"language" example:"de" len:"2"`
// 	}{
// 		ID:            p.User.ID,
// 		FirstName:     p.FirstName,
// 		LastName:      p.LastName,
// 		AvatarURL:     p.AvatarURL,
// 		Email:         p.Email,
// 		StudentNumber: p.StudentNumber,
// 		Semester:      p.Semester,
// 		Subject:       p.Subject,
// 		Language:      p.Language,
// 	}

// 	return &ExamEnrollmentResponse{
// 		Status: p.Status,
// 		Mark:   p.Mark,
// 		User:   &user,
// 	}
// }

// func newExamEnrollmentListResponse(enrollments []model.UserExamView) []render.Renderer {
// 	list := []render.Renderer{}
// 	for k := range enrollments {
// 		list = append(list, newExamEnrollmentResponse(&enrollments[k]))
// 	}

// 	return list
// }
