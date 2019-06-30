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
	"log"

	"github.com/infomark-org/infomark-backend/api"

	"github.com/spf13/cobra"
)

var numWorkers = 1

var workCmd = &cobra.Command{
	Use:   "work",
	Short: "start a worker",
	Long: `Starts a background worker which will use docker to test submissions.
Can be used with the flag "-n" to start multiple workers within one process.
`,
	Run: func(cmd *cobra.Command, args []string) {

		worker, err := api.NewWorker(numWorkers)
		if err != nil {
			log.Fatal(err)
		}
		worker.Start()
	},
}

func init() {

	workCmd.Flags().IntVarP(&numWorkers, "number", "n", 1, "number of workers within one routine")
	RootCmd.AddCommand(workCmd)
}
