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

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/database"
	"github.com/cgtuebingen/infomark-backend/logging"
	"github.com/franela/goblin"
	_ "github.com/lib/pq"

	// "github.com/spf13/viper"
	"net/http"
	"testing"
)

func DesterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("here")
		next.ServeHTTP(w, r)
		return
	})
}

func TestAccount(t *testing.T) {

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	userStore := database.NewUserStore(db)
	rs := NewAccountResource(userStore)

	g.Describe("GetAccount", func() {
		g.It("Should require valid claims", func() {

			re := helper.Payload{
				Data:   helper.H{},
				Method: "POST",
			}

			w := helper.SimulateRequest(re, rs.GetHandler, authenticate.RequiredValidAccessClaims)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should return info when claims are invalid", func() {

			re := helper.Payload{
				Data:         helper.H{},
				Method:       "POST",
				AccessClaims: authenticate.NewAccessClaims(0, true), // 0 is invalid
			}

			w := helper.SimulateRequest(re, rs.GetHandler, authenticate.RequiredValidAccessClaims)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should return info when claims are valid", func() {

			re := helper.Payload{
				Data:         helper.H{},
				Method:       "POST",
				AccessClaims: authenticate.NewAccessClaims(1, true),
			}

			w := helper.SimulateRequest(re, rs.GetHandler, authenticate.RequiredValidAccessClaims)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

	})

}

func TestAccountChanges(t *testing.T) {

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	userStore := database.NewUserStore(db)
	rs := NewAccountResource(userStore)

	g.Describe("ChangeAccount", func() {
		g.It("Should require valid access claims", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"first_name": "foo",
						"last_name":  "bar",
					},
					Method: "PATCH",
				},
				rs.EditHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.It("Should not change anything with incorrect password", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"account": helper.H{},
						"user": helper.H{
							"email": "foo",
						},
					},
					Method:       "PATCH",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.EditHandler,
				authenticate.RequiredValidAccessClaims,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should change email with correct password", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"account": helper.H{
							"plain_password": "test",
						},
						"user": helper.H{
							"email": "foo@uni-tuebingen.de",
						},
					},
					Method:       "PATCH",
					AccessClaims: authenticate.NewAccessClaims(1, true),
				},
				rs.EditHandler,
				authenticate.RequiredValidAccessClaims,
			)
			fmt.Println(w)
			g.Assert(w.Code).Equal(http.StatusNoContent)
		})
	})
}
