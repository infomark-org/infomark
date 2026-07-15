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
	"fmt"
	"log"

	"github.com/infomark-org/infomark/api/helper"
	"github.com/spf13/cobra"
)

var submissionCmd = &cobra.Command{
	Use:   "submission",
	Short: "Manage submissions",
}

// uploadCmd uploads a submission on behalf of a student.
var uploadCmd = &cobra.Command{
	Use:   "upload [courseID] [taskID] [userID] [filename]",
	Short: "Upload a submission on behalf of a student",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {

		courseID := MustInt64Parameter(args[0], "courseID")
		taskID := MustInt64Parameter(args[1], "taskID")
		userID := MustInt64Parameter(args[2], "userID")
		filename := args[3]
		if !helper.FileExists(filename) {
			log.Fatalf("File %s does not exist", filename)
		}

		conn.RequireCredentials()

		url := fmt.Sprintf("/api/v1/courses/%v/tasks/%v/submission", courseID, taskID)
		params := map[string]string{
			"user_id": fmt.Sprintf("%v", userID),
		}

		w, err := remote.UploadWithParameters(url, filename, "application/zip", params, conn)
		if err != nil {
			log.Fatal(err)
		}
		defer w.Close()
		fmt.Println(w.Response.Status)
	},
}

func init() {
	submissionCmd.AddCommand(uploadCmd)
	RootCmd.AddCommand(submissionCmd)
}
