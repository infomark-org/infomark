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

package console

import (
  "fmt"
  "log"

  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/go-ozzo/ozzo-validation/is"
  "github.com/spf13/cobra"
  null "gopkg.in/guregu/null.v3"
)

func init() {
  UserCmd.AddCommand(UserFindCmd)
  UserCmd.AddCommand(UserConfirmCmd)
  UserCmd.AddCommand(UserSetEmailCmd)
}

var UserCmd = &cobra.Command{
  Use:   "user",
  Short: "Management of users",
}

var UserFindCmd = &cobra.Command{
  Use:   "find [query]",
  Short: "find user by first_name, last_name or email",
  Long:  `List all users matching the query`,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    db, _ := MustConnectAndStores()

    query := fmt.Sprintf("%%%s%%", args[0])

    users := []model.User{}
    err := db.Select(&users, `
SELECT
  *
FROM
  users
WHERE
 last_name LIKE $1
OR
 first_name LIKE $1
OR
 email LIKE $1`, query)
    if err != nil {
      panic(err)
    }

    fmt.Printf("found %v users matching %s\n", len(users), query)
    for k, user := range users {
      fmt.Printf("%4d %20s %20s %50s\n",
        user.ID, user.FirstName, user.LastName, user.Email)
      if k%10 == 0 && k != 0 {
        fmt.Println("")
      }
    }

    fmt.Printf("found %v users matching %s\n", len(users), query)
  },
}

var UserConfirmCmd = &cobra.Command{
  Use:   "confirm [email]",
  Short: "confirms the email address manually",
  Long:  `Will run confirmation procedure for an user `,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    _, stores := MustConnectAndStores()

    email := args[0]
    if err := is.Email.Validate(email); err != nil {
      log.Fatalf("email '%s' is not a valid email\n", email)
    }

    user, err := stores.User.FindByEmail(email)
    if err != nil {
      log.Fatalf("user with email %v not found\n", email)
    }

    user.ConfirmEmailToken = null.String{}
    if err := stores.User.Update(user); err != nil {
      panic(err)
    }

    fmt.Printf("email %s of user %s %s has been confirmed\n",
      email, user.FirstName, user.LastName)
  },
}

var UserSetEmailCmd = &cobra.Command{
  Use:   "set-email [userID] [email]",
  Short: "will alter the email address",
  Long:  `Will change email address of an user without confirmation procedure`,
  Args:  cobra.ExactArgs(2),
  Run: func(cmd *cobra.Command, args []string) {
    userID := MustInt64Parameter(args[0], "userID")
    email := args[1]
    if err := is.Email.Validate(email); err != nil {
      log.Fatalf("email '%s' is not a valid email\n", email)
    }

    _, stores := MustConnectAndStores()

    user, err := stores.User.Get(userID)
    if err != nil {
      fmt.Printf("user with id %v not found\n", userID)
      return
    }

    user.Email = email
    if err := stores.User.Update(user); err != nil {
      panic(err)
    }

    fmt.Printf("email of user %s %s is now %s\n",
      user.FirstName, user.LastName, user.Email)
  },
}
