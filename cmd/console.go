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

package cmd

import (
  "github.com/cgtuebingen/infomark-backend/cmd/console"
  "github.com/spf13/cobra"
)

var ConsoleCmd = &cobra.Command{
  Use:   "console",
  Short: "infomark console commands",
}

func init() {
  ConsoleCmd.AddCommand(console.AdminCmd)
  ConsoleCmd.AddCommand(console.UserCmd)
  ConsoleCmd.AddCommand(console.CourseCmd)
  ConsoleCmd.AddCommand(console.SubmissionCmd)
  ConsoleCmd.AddCommand(console.GroupCmd)
  ConsoleCmd.AddCommand(console.DatabaseCmd)
  RootCmd.AddCommand(ConsoleCmd)
}
