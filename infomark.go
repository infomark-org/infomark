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

//go:generate go run docs/generate.go

package main

import (
	"embed"
	"io/fs"

	"github.com/infomark-org/infomark/api/app"
	"github.com/infomark-org/infomark/cmd"
)

// staticFiles embeds the compiled single-page UI from the repository-root
// static directory. This package is the only one positioned above that
// directory, and go:embed cannot reach into parent directories, so the embed
// declaration has to live here.
//
//go:embed all:static
var staticFiles embed.FS

func main() {
	// Strip the leading "static/" path segment so the assets are served from
	// the site root, exactly as the previous pkger.Dir("/static") handler did,
	// then hand the filesystem to the app package before any command runs.
	staticRoot, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	app.StaticFiles = staticRoot

	cmd.Execute()
}
