// InfoMark - a platform for managing users with
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

	// "io/ioutil"

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

func TestUser(t *testing.T) {

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
	rs := NewUserResource(stores)

	g.Describe("User Query", func() {
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

		g.It("Should list all users", func() {
			users_expected, err := rs.Stores.User.GetAll()
			g.Assert(err).Equal(nil)

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

			users_actual := []model.User{}
			err = json.NewDecoder(w.Body).Decode(&users_actual)
			g.Assert(err).Equal(nil)
			g.Assert(len(users_actual)).Equal(len(users_expected))

		})

		g.It("Should get a specific user", func() {

			user_expected, err := rs.Stores.User.Get(1)
			g.Assert(err).Equal(nil)

			w := helper.SimulateRequest(
				helper.Payload{
					Data:         helper.H{},
					Method:       "GET",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.GetHandler,
				// set user
				authenticate.RequiredValidAccessClaims,
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						ctx := context.WithValue(r.Context(), "user", user_expected)
						next.ServeHTTP(w, r.WithContext(ctx))
					})
				},
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			user_actual := &model.User{}
			err = json.NewDecoder(w.Body).Decode(user_actual)
			g.Assert(err).Equal(nil)

			g.Assert(user_actual.ID).Equal(user_expected.ID)

			g.Assert(user_actual.FirstName).Equal(user_expected.FirstName)
			g.Assert(user_actual.FirstName).Equal(user_expected.FirstName)
			g.Assert(user_actual.LastName).Equal(user_expected.LastName)
			g.Assert(user_actual.AvatarPath).Equal(user_expected.AvatarPath)
			g.Assert(user_actual.Email).Equal(user_expected.Email)
			g.Assert(user_actual.StudentNumber).Equal(user_expected.StudentNumber)
			g.Assert(user_actual.Semester).Equal(user_expected.Semester)
			g.Assert(user_actual.Subject).Equal(user_expected.Subject)
			g.Assert(user_actual.Language).Equal(user_expected.Language)

		})

	})

}
