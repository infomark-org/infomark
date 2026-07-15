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
	"github.com/infomark-org/infomark/clients/cli/bridge"
	"github.com/spf13/cobra"
)

// pingCmd represents the serve command.
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "tests if endpoint is available",
	Run: func(cmd *cobra.Command, args []string) {
		w := remote.Get("/api/v1/ping")
		defer w.Close()
		if w.OK() {
			bridge.Green(w.Plain())
		}
	},
}

func init() {
	RootCmd.AddCommand(pingCmd)
}
