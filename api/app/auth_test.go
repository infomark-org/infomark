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
	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/database"
	"github.com/cgtuebingen/infomark-backend/logging"
	"github.com/franela/goblin"
	_ "github.com/lib/pq"

	// "github.com/spf13/viper"
	"net/http"
	"testing"
)

func TestLogin(t *testing.T) {

	logger := logging.NewLogger()
	g := goblin.Goblin(t)

	db, err := helper.TransactionDB()
	defer db.Close()
	if err != nil {
		logger.WithField("module", "database").Error(err)
		return
	}

	userStore := database.NewUserStore(db)
	auth := NewAuthResource(userStore)

	g.Describe("LoginHandlers", func() {
		g.It("Not existent user should fail", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"email":          "peter.zwegat@uni-tuebingen.de",
						"plain_password": "",
					},
					Method: "POST",
				},
				auth.LoginHandler,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Wrong credentials should fail", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"email":          "test@uni-tuebingen.de",
						"plain_password": "testOops",
					},
					Method: "POST",
				},
				auth.LoginHandler,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Correct credentials should not fail", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"email":          "test@uni-tuebingen.de",
						"plain_password": "test",
					},
					Method: "POST",
				},
				auth.LoginHandler,
			)
			g.Assert(w.Code).Equal(http.StatusOK)
		})
	})

}
