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
	"context"
	"encoding/json"
	"fmt"
	_ "fmt"
	"time"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/cgtuebingen/infomark-backend/logging"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/franela/goblin"
	_ "github.com/lib/pq"

	// "github.com/spf13/viper"
	"net/http"
	"testing"
)

func SetCourseContext(course *model.Course) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "course", course)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TestCourse(t *testing.T) {

	email.DefaultMail = email.VoidMail

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	stores := NewStores(db)
	rs := NewCourseResource(stores)

	g.Describe("Course Query", func() {
		g.It("Should require claims", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data:   helper.H{},
					Method: "GET",
				},
				rs.IndexHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should list all courses", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.IndexHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			courses_actual := []model.Course{}

			err = json.NewDecoder(w.Body).Decode(&courses_actual)
			g.Assert(err).Equal(nil)
			g.Assert(len(courses_actual)).Equal(2)

		})

		g.It("Should get a specific course", func() {

			course_expected, err := stores.Course.Get(1)
			g.Assert(err).Equal(nil)

			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.GetHandler,
				// set course
				authenticate.RequiredValidAccessClaims,
				SetCourseContext(course_expected),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			course_actual := &model.Course{}
			err = json.NewDecoder(w.Body).Decode(course_actual)
			g.Assert(err).Equal(nil)

			g.Assert(course_actual.ID).Equal(course_expected.ID)
			g.Assert(course_actual.Name).Equal(course_expected.Name)
			g.Assert(course_actual.Description).Equal(course_expected.Description)
			g.Assert(course_actual.BeginsAt.Equal(course_expected.BeginsAt)).Equal(true)
			g.Assert(course_actual.EndsAt.Equal(course_expected.EndsAt)).Equal(true)
			g.Assert(course_actual.RequiredPercentage).Equal(course_expected.RequiredPercentage)

		})

	})

}

func TestCourseCreation(t *testing.T) {

	email.DefaultMail = email.VoidMail

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	stores := NewStores(db)
	rs := NewCourseResource(stores)

	g.Describe("Course Creation", func() {
		g.It("Should require claims", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data:   helper.H{},
					Method: "POST",
				},
				rs.CreateHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should require body", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "POST",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.CreateHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should create valid course", func() {

			courses_before, err := rs.Stores.Course.GetAll()
			g.Assert(err).Equal(nil)

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"name":                "Info2_new",
						"description":         "Lorem Ipsum_new",
						"begins_at":           "2019-02-01T01:02:03Z",
						"ends_at":             "2019-07-30T23:59:59Z",
						"required_percentage": 43,
					},
					Method:       "POST",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.CreateHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusCreated)

			expectedBegin, _ := time.Parse(time.RFC3339, "2019-02-01T01:02:03Z")
			expectedEnd, _ := time.Parse(time.RFC3339, "2019-07-30T23:59:59Z")

			course_return := &model.Course{}
			err = json.NewDecoder(w.Body).Decode(&course_return)
			g.Assert(course_return.Name).Equal("Info2_new")
			g.Assert(course_return.Description).Equal("Lorem Ipsum_new")
			g.Assert(course_return.BeginsAt.Equal(expectedBegin)).Equal(true)
			g.Assert(course_return.EndsAt.Equal(expectedEnd)).Equal(true)
			g.Assert(course_return.RequiredPercentage).Equal(43)

			courses_after, err := rs.Stores.Course.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(courses_after)).Equal(len(courses_before) + 1)
		})

	})

}

func TestCourseChanges(t *testing.T) {

	email.DefaultMail = email.VoidMail

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	stores := NewStores(db)
	rs := NewCourseResource(stores)

	courses_before, err := rs.Stores.Course.GetAll()
	g.Assert(err).Equal(nil)
	course_before, err := stores.Course.Get(1)
	g.Assert(err).Equal(nil)

	g.Describe("Course Changes", func() {
		g.It("Should require claims", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data:   helper.H{},
					Method: "PATCH",
				},
				rs.EditHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should perform updates", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"name":                "Info2_update",
						"description":         "Lorem Ipsum_update",
						"begins_at":           "2020-02-01T01:02:03Z",
						"ends_at":             "2020-07-30T23:59:59Z",
						"required_percentage": 99,
					},
					Method:       "PATCH",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.EditHandler,
				authenticate.RequiredValidAccessClaims,
				SetCourseContext(course_before),
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			expectedBegin, _ := time.Parse(time.RFC3339, "2020-02-01T01:02:03Z")
			expectedEnd, _ := time.Parse(time.RFC3339, "2020-07-30T23:59:59Z")

			course_after, err := stores.Course.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(course_after.Name).Equal("Info2_update")
			g.Assert(course_after.Description).Equal("Lorem Ipsum_update")
			g.Assert(course_after.BeginsAt.Equal(expectedBegin)).Equal(true)
			g.Assert(course_after.EndsAt.Equal(expectedEnd)).Equal(true)
			g.Assert(course_after.RequiredPercentage).Equal(99)
		})

		g.It("Should delete", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "DELETE",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.DeleteHandler,
				authenticate.RequiredValidAccessClaims,
				SetCourseContext(course_before),
			)
			// TODO()
			fmt.Println(w.Body)
			g.Assert(w.Code).Equal(http.StatusOK)

			courses_after, err := rs.Stores.Course.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(courses_after)).Equal(len(courses_before) - 1)
		})
	})

}
