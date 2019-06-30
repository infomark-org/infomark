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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/auth"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/franela/goblin"
	"github.com/spf13/viper"
)

func TestAccount(t *testing.T) {
	PrepareTests()
	g := goblin.Goblin(t)
	email.DefaultMail = email.VoidMail

	tape := &Tape{}

	var stores *Stores
	adminJWT := NewJWTRequest(1, true)
	noAdminJWT := NewJWTRequest(1, false)

	g.Describe("Account", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
			stores = NewStores(tape.DB)
		})

		g.It("Query should require valid claims", func() {
			w := tape.Get("/api/v1/account")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Get("/api/v1/account", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

		})

		g.It("Should get all enrollments", func() {
			enrollmentsExpected, err := stores.User.GetEnrollments(1)
			g.Assert(err).Equal(nil)

			w := tape.Get("/api/v1/account/enrollments", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			enrollmentsActual := []UserEnrollmentResponse{}
			err = json.NewDecoder(w.Body).Decode(&enrollmentsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(enrollmentsActual)).Equal(len(enrollmentsExpected))

			for j := 0; j < len(enrollmentsExpected); j++ {
				g.Assert(enrollmentsActual[j].Role).Equal(enrollmentsExpected[j].Role)
				g.Assert(enrollmentsActual[j].CourseID).Equal(enrollmentsExpected[j].CourseID)
				g.Assert(enrollmentsActual[j].ID).Equal(int64(0))
			}
		})

		g.It("Should not create invalid accounts (missing user data)", func() {
			w := tape.Post("/api/v1/account",
				H{
					"account": H{
						"email":          "foo@test.com",
						"plain_password": "bar",
					},
					"user": H{
						"first_name": "",
					},
				})
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should not create accounts with too short password", func() {

			minLen := viper.GetInt("min_password_length")
			tooShortPassword := auth.GenerateToken(minLen - 1)

			w := tape.Post("/api/v1/account",
				H{
					"account": H{
						"email":          "foo@test.com",
						"plain_password": tooShortPassword,
					},
					"user": H{
						"first_name": "Data",
					},
				})
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should create valid accounts", func() {

			minLen := viper.GetInt("min_password_length")
			validPassword := auth.GenerateToken(minLen)

			request := H{
				"user": H{
					"first_name":     "Max  ",
					"last_name":      "  Mustermensch",   // contains whitespaces
					"email":          "max@Mensch.com  ", // contains uppercase
					"student_number": "0815",
					"semester":       2,
					"subject":        "bio2",
					"language":       "de",
				},
				"account": H{
					"email":          "max@Mensch.com  ",
					"plain_password": validPassword,
				},
			}

			w := tape.Post("/api/v1/account", request)
			g.Assert(w.Code).Equal(http.StatusCreated)

			userAfter, err := stores.User.FindByEmail("max@mensch.com")
			g.Assert(err).Equal(nil)

			g.Assert(userAfter.FirstName).Equal("Max")
			g.Assert(userAfter.LastName).Equal("Mustermensch")
			g.Assert(userAfter.Email).Equal("max@mensch.com")
			g.Assert(userAfter.StudentNumber).Equal("0815")
			g.Assert(userAfter.Semester).Equal(2)
			g.Assert(userAfter.Subject).Equal("bio2")
			g.Assert(userAfter.Language).Equal("de")
			g.Assert(userAfter.Root).Equal(false)

			g.Assert(userAfter.ConfirmEmailToken.Valid).Equal(true)
			g.Assert(userAfter.ResetPasswordToken.Valid).Equal(false)
			g.Assert(userAfter.AvatarURL.Valid).Equal(false)

			g.Assert(auth.CheckPasswordHash(validPassword, userAfter.EncryptedPassword)).Equal(true)
		})

		g.It("Changes should require valid access-claims", func() {

			data := H{
				"account": H{
					"email":          "foo@uni-tuebingen.de",
					"plain_password": "new_pass",
				},
				"old_plain_password": "test",
			}

			w := tape.Patch("/api/v1/account", data)
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.Patch("/api/v1/account", data, adminJWT)
			g.Assert(w.Code).Equal(http.StatusNoContent)
		})

		g.It("Changes should require valid credentials", func() {

			data := H{
				"account": H{
					"email":          "foo@uni-tuebingen.de",
					"plain_password": "new_pass",
				},
				"old_plain_password": "test_false",
			}

			w := tape.Patch("/api/v1/account", data, adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should change email and password when correct password ", func() {

			data := H{
				"account": H{
					"email":          "foo@uni-tuebingen.de",
					"plain_password": "new_pass",
				},
				"old_plain_password": "test",
			}

			w := tape.Patch("/api/v1/account", data, adminJWT)
			g.Assert(w.Code).Equal(http.StatusNoContent)

			userAfter, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userAfter.Email).Equal("foo@uni-tuebingen.de")

			isPasswordValid := auth.CheckPasswordHash("new_pass", userAfter.EncryptedPassword)
			g.Assert(isPasswordValid).Equal(true)
			g.Assert(userAfter.ConfirmEmailToken.Valid).Equal(true)
		})

		g.It("Should only change email when correct old password ", func() {

			data := H{
				"account": H{
					"email": "foo@uni-tuebingen.de",
				},
				"old_plain_password": "test",
			}

			w := tape.Patch("/api/v1/account", data, adminJWT)
			g.Assert(w.Code).Equal(http.StatusNoContent)

			userAfter, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userAfter.Email).Equal("foo@uni-tuebingen.de")

			isPasswordValid := auth.CheckPasswordHash("test", userAfter.EncryptedPassword)
			g.Assert(isPasswordValid).Equal(true)
			g.Assert(userAfter.ConfirmEmailToken.Valid).Equal(true)
		})

		g.It("Should only require valid email when correct old password ", func() {

			data := H{
				"account": H{
					"email": "foo@uni-tuebingen",
				},
				"old_plain_password": "test",
			}

			w := tape.Patch("/api/v1/account", data, adminJWT)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
		})

		g.It("Should only change password when correct old password ", func() {

			data := H{
				"account": H{
					"plain_password": "fooerrr",
				},
				"old_plain_password": "test",
			}

			w := tape.Patch("/api/v1/account", data, adminJWT)
			g.Assert(w.Code).Equal(http.StatusNoContent)

			userAfter, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(userAfter.Email).Equal("test@uni-tuebingen.de")

			isPasswordValid := auth.CheckPasswordHash("fooerrr", userAfter.EncryptedPassword)
			g.Assert(isPasswordValid).Equal(true)
			g.Assert(userAfter.ConfirmEmailToken.Valid).Equal(false)
		})

		g.It("should change avatar (jpg)", func() {
			defer helper.NewAvatarFileHandle(1).Delete()

			// no file so far
			g.Assert(helper.NewAvatarFileHandle(1).Exists()).Equal(false)

			// no avatar by default
			w := tape.Get("/api/v1/account", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			userReturned := &UserResponse{}
			err := json.NewDecoder(w.Body).Decode(userReturned)
			g.Assert(err).Equal(nil)
			g.Assert(userReturned.AvatarURL.Valid).Equal(false)

			// upload avatar
			avatarFilename := fmt.Sprintf("%s/default-avatar.jpg", viper.GetString("fixtures_dir"))
			w, err = tape.Upload("/api/v1/account/avatar", avatarFilename, "image/jpg", adminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)
			g.Assert(helper.NewAvatarFileHandle(1).Exists()).Equal(true)

			user, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user.AvatarURL.Valid).Equal(true)

			// there should be now an avatar
			w = tape.Get("/api/v1/account", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&userReturned)
			g.Assert(err).Equal(nil)
			g.Assert(userReturned.AvatarURL.Valid).Equal(true)

			w = tape.Get("/api/v1/account/avatar", adminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)

			if !strings.HasSuffix(w.Header().Get("Content-Type"), "jpeg") {
				g.Assert(strings.HasSuffix(w.Header().Get("Content-Type"), "jpg")).Equal(true)
			}

		})

		g.It("should change avatar (png)", func() {
			defer helper.NewAvatarFileHandle(1).Delete()

			// no file so far
			g.Assert(helper.NewAvatarFileHandle(1).Exists()).Equal(false)

			// no avatar by default
			w := tape.Get("/api/v1/account", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			userReturned := &UserResponse{}
			err := json.NewDecoder(w.Body).Decode(userReturned)
			g.Assert(err).Equal(nil)
			g.Assert(userReturned.AvatarURL.Valid).Equal(false)

			// upload avatar
			avatarFilename := fmt.Sprintf("%s/default-avatar.png", viper.GetString("fixtures_dir"))
			w, err = tape.Upload("/api/v1/account/avatar", avatarFilename, "image/png", adminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)
			g.Assert(helper.NewAvatarFileHandle(1).Exists()).Equal(true)

			user, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(user.AvatarURL.Valid).Equal(true)

			// there should be now an avatar
			w = tape.Get("/api/v1/account", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)
			err = json.NewDecoder(w.Body).Decode(&userReturned)
			g.Assert(err).Equal(nil)
			g.Assert(userReturned.AvatarURL.Valid).Equal(true)

			w = tape.Get("/api/v1/account/avatar", adminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)
			g.Assert(strings.HasSuffix(w.Header().Get("Content-Type"), "png")).Equal(true)

		})

		g.It("reject to large avatars (jpg)", func() {
			defer helper.NewAvatarFileHandle(1).Delete()

			// create 10MB file > 5kb
			f, err := os.Create("/tmp/foo.jpg")
			if err != nil {
				log.Fatal(err)
			}
			if err := f.Truncate(1e7); err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			defer func() {
				os.Remove("/tmp/foo.jpg")
			}()

			// no file so far
			g.Assert(helper.NewAvatarFileHandle(1).Exists()).Equal(false)

			// no avatar by default
			w := tape.Get("/api/v1/account", adminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			userReturned := &UserResponse{}
			err = json.NewDecoder(w.Body).Decode(userReturned)
			g.Assert(err).Equal(nil)
			g.Assert(userReturned.AvatarURL.Valid).Equal(false)

			// upload avatar
			w, err = tape.Upload("/api/v1/account/avatar", "/tmp/foo.jpg", "image/jpg", adminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusBadRequest)
			g.Assert(helper.NewAvatarFileHandle(1).Exists()).Equal(false)

		})

		g.It("Should have a way to delete own avatar", func() {
			defer helper.NewAvatarFileHandle(1).Delete()

			// no file so far
			g.Assert(helper.NewAvatarFileHandle(1).Exists()).Equal(false)

			// upload avatar
			avatarFilename := fmt.Sprintf("%s/default-avatar.jpg", viper.GetString("fixtures_dir"))
			w, err := tape.Upload("/api/v1/account/avatar", avatarFilename, "image/jpg", noAdminJWT)
			g.Assert(err).Equal(nil)
			g.Assert(w.Code).Equal(http.StatusOK)
			g.Assert(helper.NewAvatarFileHandle(1).Exists()).Equal(true)

			// there should be now an avatar
			w = tape.Get("/api/v1/account", noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			userReturned := &UserResponse{}
			err = json.NewDecoder(w.Body).Decode(userReturned)
			g.Assert(err).Equal(nil)
			g.Assert(userReturned.AvatarURL.Valid).Equal(true)

			// delete
			w = tape.Delete("/api/v1/account/avatar", noAdminJWT)
			g.Assert(w.Code).Equal(http.StatusOK)

			g.Assert(helper.NewAvatarFileHandle(1).Exists()).Equal(false)

		})

		g.AfterEach(func() {
			tape.AfterEach()
		})
	})

}
