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
	"github.com/cgtuebingen/infomark-backend/auth"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/cgtuebingen/infomark-backend/logging"
	"github.com/franela/goblin"
	_ "github.com/lib/pq"
	null "gopkg.in/guregu/null.v3"

	// "github.com/spf13/viper"
	"net/http"
	"testing"
)

func TestAuthComponent(t *testing.T) {

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
	rs := NewAuthResource(stores)

	g.Describe("LoginHandler", func() {
		g.It("Not existent user should fail", func() {
			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"email":          "peter.zwegat@uni-tuebingen.de",
						"plain_password": "",
					},
					Method: "POST",
				},
				rs.LoginHandler,
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
				rs.LoginHandler,
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
				rs.LoginHandler,
			)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Password-Reset will fail if email invalid", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"email": "test2@uni-tuebingen.de",
					},
					Method: "POST",
				},
				rs.RequestPasswordResetHandler,
			)
			g.Assert(w.Code).Equal(http.StatusNotFound)
		})

		g.It("Invalid Password-Reset-Token is denied", func() {

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"reset_password_token": "invalid_string",
						"plain_password":       "new_password",
					},
					Method: "POST",
				},
				rs.UpdatePasswordHandler,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			user_after, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user_after.Email).Equal("test@uni-tuebingen.de")
		})

		g.It("Correct Password-Reset-Token will change password", func() {

			// fetch reset token
			user_before, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user_before.Email).Equal("test@uni-tuebingen.de")
			g.Assert(user_before.ResetPasswordToken.Valid).Equal(false)

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"email": "test@uni-tuebingen.de",
					},
					Method: "POST",
				},
				rs.RequestPasswordResetHandler,
			)
			fmt.Println(w.Body)
			g.Assert(w.Code).Equal(http.StatusOK)

			user_after, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user_after.Email).Equal("test@uni-tuebingen.de")
			g.Assert(user_after.ResetPasswordToken.Valid).Equal(true)

			// use token to reset password
			w = helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"reset_password_token": user_after.ResetPasswordToken.String,
						"plain_password":       "new_password",
						"email":                "test@uni-tuebingen.de",
					},
					Method: "POST",
				},
				rs.UpdatePasswordHandler,
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			user_after2, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user_after2.Email).Equal("test@uni-tuebingen.de")
			g.Assert(user_after2.ResetPasswordToken.Valid).Equal(false)

			password_valid := auth.CheckPasswordHash("new_password", user_after2.EncryptedPassword)
			g.Assert(password_valid).Equal(true)

			password_valid = auth.CheckPasswordHash("test", user_after2.EncryptedPassword)
			g.Assert(password_valid).Equal(false)

		})

		g.It("Should not login when confirm email token is set", func() {

			// fetch reset token
			user_before, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user_before.Email).Equal("test@uni-tuebingen.de")
			g.Assert(user_before.ConfirmEmailToken.Valid).Equal(false)
			user_before.ConfirmEmailToken = null.StringFrom("testtoken")
			stores.User.Update(user_before)

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"email":          "test@uni-tuebingen.de",
						"plain_password": "test",
					},
					Method: "POST",
				},
				rs.LoginHandler,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

		})

		g.It("Should not confirm email with incorrect token", func() {

			// fetch reset token
			user_before, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user_before.Email).Equal("test@uni-tuebingen.de")
			user_before.ConfirmEmailToken = null.StringFrom("testtoken")
			stores.User.Update(user_before)
			g.Assert(user_before.ConfirmEmailToken.Valid).Equal(true)

			// cannot login

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"email":              "test@uni-tuebingen.de",
						"confirmation_token": "testtoken_wrong",
					},
					Method: "POST",
				},
				rs.ConfirmEmailHandler,
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			user_after, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user_after.ConfirmEmailToken.Valid).Equal(true)

		})

		g.It("Should confirm email with correct token", func() {

			// fetch reset token
			user_before, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user_before.Email).Equal("test@uni-tuebingen.de")
			user_before.ConfirmEmailToken = null.StringFrom("testtoken")
			stores.User.Update(user_before)
			g.Assert(user_before.ConfirmEmailToken.Valid).Equal(true)

			// cannot login

			w := helper.SimulateRequest(
				helper.Payload{
					Data: helper.H{
						"email":              "test@uni-tuebingen.de",
						"confirmation_token": "testtoken",
					},
					Method: "POST",
				},
				rs.ConfirmEmailHandler,
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			user_after, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user_after.ConfirmEmailToken.Valid).Equal(false)

		})
	})

}
