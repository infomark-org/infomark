// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019 ComputerGraphics Tuebingen
//               2020-present InfoMark.org
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
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/infomark-org/infomark-backend/cmd/console"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// ConsoleCmd starts the infomark console
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
	ConsoleCmd.AddCommand(console.ConfigurationCmd)

	UtilsCmd.AddCommand(UtilsCompletionCmd)
	UtilsCmd.AddCommand(UtilsDocCmd)
	ConsoleCmd.AddCommand(UtilsCmd)

	RootCmd.AddCommand(ConsoleCmd)

	// doc.GenMarkdownTree(RootCmd, "/tmp")
}

var UtilsCmd = &cobra.Command{
	Use:   "utils",
	Short: "Some helper functions.",
}

var UtilsCompletionCmd = &cobra.Command{
	Use:   "completion [shell]",
	Short: "Output (bash/zsh) shell completion code",
	Long: `Pipe the stdout to your completion collection, e.g.,

./infomark console utils completion zsh > ~/.my-shell/completions/_infomark

`,
	// Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatalln("Expected one argument with the desired shell")
		}

		switch args[0] {
		case "bash":
			RootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			RootCmd.GenZshCompletion(os.Stdout)
		default:
			log.Fatalf("Unknown shell %s, only bash and zsh are available\n", args[0])
		}
	},
}

var UtilsDocCmd = &cobra.Command{
	Use:   "doc",
	Short: "Generates docs for console",
	// Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			log.Fatalln("Expected one argument with the destination dir")
		}

		const fmTemplate = `---
date: %s
title: "%s"
slug: %s
url: %s
lastmod: %s
layout: subpagewithout
---

`

		filePrepender := func(filename string) string {
			now := time.Now().Format(time.RFC3339)
			name := filepath.Base(filename)
			base := strings.TrimSuffix(name, path.Ext(name))
			url := "/guides/console/commands/" + strings.ToLower(base) + "/"
			return fmt.Sprintf(fmTemplate, now, "Console", base, url, now)
		}

		linkHandler := func(name string) string {
			base := strings.TrimSuffix(name, path.Ext(name))
			return "/guides/console/commands/" + strings.ToLower(base) + "/"
		}

		err := doc.GenMarkdownTreeCustom(RootCmd, args[0], filePrepender, linkHandler)
		if err != nil {
			log.Fatal(err)
		}
	},
}
