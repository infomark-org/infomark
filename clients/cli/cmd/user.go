// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  Patrick Wieschollek
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

package cmd

import (
	"fmt"

	"github.com/infomark-org/infomark/api/app"
	"github.com/spf13/cobra"
)

// meCmd shows the account information of the authenticated user.
var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show my account information",
	Run: func(cmd *cobra.Command, args []string) {
		conn.RequireCredentials()
		w := remote.Get("/api/v1/account", conn)
		defer w.Close()

		data := &app.UserResponse{}
		w.DecodeJSON(data)
		fmt.Println("ID            ", data.ID)
		fmt.Println("FirstName     ", data.FirstName)
		fmt.Println("LastName      ", data.LastName)
		// AvatarURL became a nullable column in the current backend
		// (guregu null.String). Render its underlying value so the output
		// stays a plain URL string as in the original client.
		fmt.Println("AvatarURL     ", data.AvatarURL.ValueOrZero())
		fmt.Println("Email         ", data.Email)
		fmt.Println("StudentNumber ", data.StudentNumber)
		fmt.Println("Semester      ", data.Semester)
		fmt.Println("Subject       ", data.Subject)
		fmt.Println("Language      ", data.Language)
		fmt.Println("Root          ", data.Root)
	},
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
}

var findCmd = &cobra.Command{
	Use:   "find [query]",
	Short: "find a user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		conn.RequireCredentials()

		url := fmt.Sprintf("/api/v1/users/find?query=%s", query)
		w := remote.Get(url, conn)
		defer w.Close()

		users := []app.UserResponse{}
		w.DecodeJSON(&users)

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

func init() {
	RootCmd.AddCommand(meCmd)

	userCmd.AddCommand(findCmd)
	RootCmd.AddCommand(userCmd)
}
