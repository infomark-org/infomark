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
	"net/http/httptest"
	"testing"

	"github.com/cgtuebingen/infomark-backend/auth"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/franela/goblin"
	redis "github.com/go-redis/redis"
	"github.com/spf13/viper"
	null "gopkg.in/guregu/null.v3"
)

func TestAuth(t *testing.T) {
	PrepareTests()
	g := goblin.Goblin(t)
	email.DefaultMail = email.VoidMail

	tape := &Tape{}

	var w *httptest.ResponseRecorder
	var stores *Stores

	option, err := redis.ParseURL(viper.GetString("redis_url"))
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(option)

	g.Describe("Auth", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
			tape.Router, _ = New(tape.DB, false)
			stores = NewStores(tape.DB)
			_ = stores
		})

		g.It("Not existent user cannot log in", func() {

			w = tape.Post("/api/v1/auth/sessions",
				H{
					"email":          "peter.zwegat@uni-tuebingen.de",
					"plain_password": "",
				},
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Wrong credentials should fail", func() {

			w = tape.Post("/api/v1/auth/sessions",
				H{
					"email":          "test@uni-tuebingen.de",
					"plain_password": "testOops",
				},
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should not login when confirm email token is set", func() {

			// tamper confirmation token reset token
			userBefore, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userBefore.Email).Equal("test@uni-tuebingen.de")
			g.Assert(userBefore.ConfirmEmailToken.Valid).Equal(false)
			userBefore.ConfirmEmailToken = null.StringFrom("testtoken")
			stores.User.Update(userBefore)

			w = tape.Post("/api/v1/auth/sessions",
				H{
					"email":          "test@uni-tuebingen.de",
					"plain_password": "test",
				},
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Correct credentials should log in", func() {

			w = tape.Post("/api/v1/auth/sessions",
				H{
					"email":          "test@uni-tuebingen.de",
					"plain_password": "test",
				},
			)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Password-Reset will fail if email invalid", func() {

			w = tape.Post("/api/v1/auth/request_password_reset",
				H{
					"email": "test2@uni-tuebingen.de",
				},
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Correct Password-Reset-Token will change password", func() {

			// state before
			userBefore, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userBefore.Email).Equal("test@uni-tuebingen.de")
			g.Assert(userBefore.ResetPasswordToken.Valid).Equal(false)

			w = tape.Post("/api/v1/auth/request_password_reset",
				H{
					"email": "test@uni-tuebingen.de",
				},
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			// state after request
			userAfter, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userAfter.Email).Equal("test@uni-tuebingen.de")
			g.Assert(userAfter.ResetPasswordToken.Valid).Equal(true)

			// use token to reset password
			w = tape.Post("/api/v1/auth/update_password",
				H{
					"reset_password_token": userAfter.ResetPasswordToken.String,
					"plain_password":       "new_password",
					"email":                "test@uni-tuebingen.de",
				},
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			userAfter2, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userAfter2.Email).Equal("test@uni-tuebingen.de")
			g.Assert(userAfter2.ResetPasswordToken.Valid).Equal(false)

			isPasswordValid := auth.CheckPasswordHash("new_password", userAfter2.EncryptedPassword)
			g.Assert(isPasswordValid).Equal(true)

			isPasswordValid = auth.CheckPasswordHash("test", userAfter2.EncryptedPassword)
			g.Assert(isPasswordValid).Equal(false)
		})

		g.It("Invalid Password-Reset-Token is denied", func() {

			w = tape.Post("/api/v1/auth/update_password",
				H{
					"reset_password_token": "invalid_string",
					"plain_password":       "new_password",
				},
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			userAfter, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userAfter.Email).Equal("test@uni-tuebingen.de")
		})

		g.It("Invalid Email-Confirmation-Token is denied", func() {

			// setup confirmation token
			userBefore, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userBefore.Email).Equal("test@uni-tuebingen.de")
			userBefore.ConfirmEmailToken = null.StringFrom("testtoken")
			stores.User.Update(userBefore)
			g.Assert(userBefore.ConfirmEmailToken.Valid).Equal(true)

			w = tape.Post("/api/v1/auth/confirm_email",
				H{
					"email":              "test@uni-tuebingen.de",
					"confirmation_token": "testtoken_wrong",
				},
			)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			userAfter, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userAfter.ConfirmEmailToken.Valid).Equal(true)
		})

		g.It("Correct Email-Confirmation-Token will confirm email", func() {

			// setup confirmation token
			userBefore, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userBefore.Email).Equal("test@uni-tuebingen.de")
			userBefore.ConfirmEmailToken = null.StringFrom("testtoken")
			stores.User.Update(userBefore)
			g.Assert(userBefore.ConfirmEmailToken.Valid).Equal(true)

			w = tape.Post("/api/v1/auth/confirm_email",
				H{
					"email":              "test@uni-tuebingen.de",
					"confirmation_token": "testtoken",
				},
			)
			g.Assert(w.Code).Equal(http.StatusOK)

			userAfter, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userAfter.ConfirmEmailToken.Valid).Equal(false)
		})

		g.It("Should limit requests per minute to do an login", func() {
			payload := H{
				"email":          "test@uni-tuebingen.de",
				"plain_password": "test",
			}

			for i := 0; i < 10; i++ {
				w = tape.Post("/api/v1/auth/sessions", payload)
			}
			g.Assert(w.Code).Equal(http.StatusOK)
			w = tape.Post("/api/v1/auth/sessions", payload)
			g.Assert(w.Code).Equal(http.StatusTooManyRequests)
		})

		g.AfterEach(func() {
			tape.AfterEach()
			err := redisClient.Set("infomark-logins:1.2.3.4-infomark-logins", "0", 0).Err()
			g.Assert(err).Equal(nil)
		})

	})

}
