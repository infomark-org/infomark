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
	_ "fmt"
	"io/ioutil"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/database"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/cgtuebingen/infomark-backend/logging"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/franela/goblin"
	_ "github.com/lib/pq"

	// "github.com/spf13/viper"
	"net/http"
	"testing"
)

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

	courseStore := database.NewCourseStore(db)
	rs := NewCourseResource(courseStore)

	g.Describe("IndexCourses", func() {
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

			courses := []model.Course{}
			temp, err := ioutil.ReadAll(w.Body)
			g.Assert(err).Equal(nil)
			err = json.Unmarshal(temp, &courses)
			g.Assert(err).Equal(nil)
			g.Assert(len(courses)).Equal(2)

		})

		g.It("Should get a specific course", func() {

			course_expected, err := courseStore.Get(1)
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
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						ctx := context.WithValue(r.Context(), "course", course_expected)
						next.ServeHTTP(w, r.WithContext(ctx))
					})
				},
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			course_actual := &model.Course{}
			json.NewDecoder(w.Body).Decode(course_actual)

			g.Assert(course_actual.ID).Equal(course_expected.ID)
			g.Assert(course_actual.Name).Equal(course_expected.Name)
			g.Assert(course_actual.Description).Equal(course_expected.Description)
			g.Assert(course_actual.BeginsAt.Equal(course_expected.BeginsAt)).Equal(true)
			g.Assert(course_actual.EndsAt.Equal(course_expected.EndsAt)).Equal(true)
			g.Assert(course_actual.RequiredPercentage).Equal(course_expected.RequiredPercentage)

		})

	})

}
