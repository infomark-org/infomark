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

  "github.com/spf13/cobra"
)

func init() {
  AdminCmd.AddCommand(AdminRemoveCmd)
  AdminCmd.AddCommand(AdminAddCmd)
}

var AdminCmd = &cobra.Command{
  Use:   "admin",
  Short: "Management of global admins.",
}

var AdminAddCmd = &cobra.Command{
  Use:   "add [userID]",
  Short: "set gives an user global admin permission",
  Long: `Will set the gobal root flag to "true" for a given user
 bypassing all permission tests`,
  Args: cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    userID := MustInt64Parameter(args[0], "userID")

    _, stores := MustConnectAndStores()

    user, err := stores.User.Get(userID)
    if err != nil {
      log.Fatalf("user with id %v not found\n", userID)
    }

    user.Root = true
    if err := stores.User.Update(user); err != nil {
      panic(err)
    }

    fmt.Printf("The user %s %s (id:%v) has now global admin privileges\n",
      user.FirstName, user.LastName, user.ID)
  },
}

var AdminRemoveCmd = &cobra.Command{
  Use:   "remove [userID]",
  Short: "removes global admin permission from a user",
  Long:  `Will set the gobal root flag to false for a user `,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    userID := MustInt64Parameter(args[0], "userID")

    _, stores := MustConnectAndStores()

    user, err := stores.User.Get(userID)
    if err != nil {
      log.Fatalf("user with id %v not found\n", userID)
    }
    user.Root = false
    if err := stores.User.Update(user); err != nil {
      panic(err)
    }

    fmt.Printf("user %s %s (%v) is not an admin anymore\n", user.FirstName, user.LastName, user.ID)

  },
}
