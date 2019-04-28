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
  CourseCmd.AddCommand(UserEnrollInCourse)
}

var CourseCmd = &cobra.Command{
  Use:   "course",
  Short: "Management of cours assignment",
}

var UserEnrollInCourse = &cobra.Command{
  Use:   "enroll [courseID] [userID] [role]",
  Short: "will enroll a user into course",
  Args:  cobra.ExactArgs(3),
  Run: func(cmd *cobra.Command, args []string) {
    var err error

    courseID := MustInt64Parameter(args[0], "courseID")
    userID := MustInt64Parameter(args[1], "userID")

    role := int64(0)
    switch args[2] {
    case "admin":
      role = int64(2)
    case "tutor":
      role = int64(1)
    case "student":
      role = int64(0)
    default:
      log.Fatalf("role '%s' must be one of 'student', 'tutor', 'admin'\n", args[2])
    }

    _, stores := MustConnectAndStores()

    user, err := stores.User.Get(userID)
    if err != nil {
      log.Fatal("user with id %v not found\n", userID)
    }

    course, err := stores.Course.Get(courseID)
    if err != nil {
      log.Fatal("user with id %v not found\n", userID)
    }

    if err := stores.Course.Enroll(course.ID, user.ID, role); err != nil {
      panic(err)
    }

    fmt.Printf("user %s %s is now enrolled in course %v with role %v\n",
      user.FirstName, user.LastName, course.ID, role)
  },
}
