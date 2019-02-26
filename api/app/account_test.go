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
  "net/http/httptest"
  "testing"

  "github.com/cgtuebingen/infomark-backend/auth"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/franela/goblin"
  "github.com/spf13/viper"
)

func TestAccount(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var w *httptest.ResponseRecorder
  var stores *Stores

  g.Describe("Account", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      tape.Router, _ = New(tape.DB, false)
      stores = NewStores(tape.DB)
      _ = stores
    })

    g.It("Query should require valid claims", func() {
      w = tape.Play("GET", "/api/v1/account")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.PlayWithClaims("GET", "/api/v1/account", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

    })

    g.Xit("Query should not return info when claims are invalid", func() {
      // we removed that endpoint
      w = tape.PlayWithClaims("GET", "/api/v1/account", 0, true)
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.It("Should get all enrollments", func() {
      enrollments_expected, err := stores.User.GetEnrollments(1)
      g.Assert(err).Equal(nil)

      w = tape.PlayWithClaims("GET", "/api/v1/account/enrollments", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      enrollments_actual := []model.Enrollment{}
      err = json.NewDecoder(w.Body).Decode(&enrollments_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(enrollments_actual)).Equal(len(enrollments_expected))

      for j := 0; j < len(enrollments_expected); j++ {
        // fmt.Println(j)
        g.Assert(enrollments_actual[j].Role).Equal(enrollments_expected[j].Role)
        g.Assert(enrollments_actual[j].CourseID).Equal(enrollments_expected[j].CourseID)
        g.Assert(enrollments_actual[j].ID).Equal(int64(0))
      }
    })

    g.It("Should not create invalid accounts (missing user data)", func() {
      w = tape.PlayData("POST", "/api/v1/account",
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

      min_len := viper.GetInt("min_password_length")
      too_short_password := auth.GenerateToken(min_len - 1)

      w = tape.PlayData("POST", "/api/v1/account",
        H{
          "account": H{
            "email":          "foo@test.com",
            "plain_password": too_short_password,
          },
          "user": H{
            "first_name": "Data",
          },
        })
      g.Assert(w.Code).Equal(http.StatusBadRequest)
    })

    g.It("Should create valid accounts", func() {

      min_len := viper.GetInt("min_password_length")
      ok_password := auth.GenerateToken(min_len)

      w = tape.PlayData("POST", "/api/v1/account",
        H{
          "account": H{
            "email":          "foo@test.com",
            "plain_password": ok_password,
          },
          "user": H{
            "first_name":     "Max",
            "last_name":      "Mustermensch",
            "semester":       2,
            "student_number": "0815",
            "subject":        "math",
            "email":          "foo@test.com",
          },
        })
      g.Assert(w.Code).Equal(http.StatusCreated)

      user_after, err := stores.User.FindByEmail("foo@test.com")
      g.Assert(err).Equal(nil)
      g.Assert(user_after.Email).Equal("foo@test.com")
      g.Assert(user_after.ConfirmEmailToken.Valid).Equal(true)
      g.Assert(user_after.FirstName).Equal("Max")
      g.Assert(user_after.LastName).Equal("Mustermensch")
      g.Assert(user_after.Semester).Equal(2)
      g.Assert(user_after.StudentNumber).Equal("0815")
      g.Assert(user_after.Subject).Equal("math")

      password_valid := auth.CheckPasswordHash(ok_password, user_after.EncryptedPassword)
      g.Assert(password_valid).Equal(true)
    })

    g.It("Changes should require valid access-claims", func() {

      data := H{
        "account": H{
          "email":          "foo@uni-tuebingen.de",
          "plain_password": "new_pass",
        },
        "old_plain_password": "test",
      }

      w = tape.PlayData("PUT", "/api/v1/account", data)
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.PlayDataWithClaims("PUT", "/api/v1/account", data, 1, true)
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

      w = tape.PlayDataWithClaims("PUT", "/api/v1/account", data, 1, true)
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

      w = tape.PlayDataWithClaims("PUT", "/api/v1/account", data, 1, true)
      g.Assert(w.Code).Equal(http.StatusNoContent)

      user_after, err := stores.User.Get(1)
      g.Assert(err).Equal(nil)
      g.Assert(user_after.Email).Equal("foo@uni-tuebingen.de")

      password_valid := auth.CheckPasswordHash("new_pass", user_after.EncryptedPassword)
      g.Assert(password_valid).Equal(true)
      g.Assert(user_after.ConfirmEmailToken.Valid).Equal(true)
    })

    g.It("Should only change email when correct old password ", func() {

      data := H{
        "account": H{
          "email": "foo@uni-tuebingen.de",
        },
        "old_plain_password": "test",
      }

      w = tape.PlayDataWithClaims("PUT", "/api/v1/account", data, 1, true)
      g.Assert(w.Code).Equal(http.StatusNoContent)

      user_after, err := stores.User.Get(1)
      g.Assert(err).Equal(nil)
      g.Assert(user_after.Email).Equal("foo@uni-tuebingen.de")

      password_valid := auth.CheckPasswordHash("test", user_after.EncryptedPassword)
      g.Assert(password_valid).Equal(true)
      g.Assert(user_after.ConfirmEmailToken.Valid).Equal(true)
    })

    g.It("Should only require valid email when correct old password ", func() {

      data := H{
        "account": H{
          "email": "foo@uni-tuebingen",
        },
        "old_plain_password": "test",
      }

      w = tape.PlayDataWithClaims("PUT", "/api/v1/account", data, 1, true)
      g.Assert(w.Code).Equal(http.StatusBadRequest)
    })

    g.It("Should only change password when correct old password ", func() {

      data := H{
        "account": H{
          "plain_password": "fooerrr",
        },
        "old_plain_password": "test",
      }

      w = tape.PlayDataWithClaims("PUT", "/api/v1/account", data, 1, true)
      g.Assert(w.Code).Equal(http.StatusNoContent)

      user_after, err := stores.User.Get(1)
      g.Assert(err).Equal(nil)
      g.Assert(user_after.Email).Equal("test@uni-tuebingen.de")

      password_valid := auth.CheckPasswordHash("fooerrr", user_after.EncryptedPassword)
      g.Assert(password_valid).Equal(true)
      g.Assert(user_after.ConfirmEmailToken.Valid).Equal(false)
    })

    g.It("Should have empty avatar url when no avatar is given", func() {

      // defer helper.NewAvatarFileHandle(1).Delete()

      // no avatar by default
      w = tape.PlayWithClaims("GET", "/api/v1/account", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      user_return := model.User{}
      err := json.NewDecoder(w.Body).Decode(&user_return)
      g.Assert(err).Equal(nil)
      g.Assert(user_return.AvatarURL.Valid).Equal(false)

      // upload avatar
      avatar_filename := fmt.Sprintf("%s/default-avatar.jpg", viper.GetString("fixtures_dir"))
      body, ct, err := tape.CreateFileRequestBody(avatar_filename, "image/jpg")
      g.Assert(err).Equal(nil)

      r, _ := http.NewRequest("POST", "/api/v1/account/avatar", body)
      r.Header.Set("Content-Type", ct)
      w := tape.PlayRequestWithClaims(r, 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      // there should be now an avatar
      w = tape.PlayWithClaims("GET", "/api/v1/account", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
      err = json.NewDecoder(w.Body).Decode(&user_return)
      g.Assert(err).Equal(nil)
      g.Assert(user_return.AvatarURL.Valid).Equal(true)

    })

    g.AfterEach(func() {
      tape.AfterEach()
    })
  })

}
