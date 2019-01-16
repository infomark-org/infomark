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

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/cgtuebingen/infomark-backend/router"
	"github.com/cgtuebingen/infomark-backend/router/auth"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var flags = []cli.Flag{
	cli.BoolFlag{
		EnvVar: "INFOMARK_DEBUG",
		Name:   "debug",
		Usage:  "enable server debug mode",
	},
	cli.StringFlag{
		EnvVar: "INFOMARK_SERVER_ADDR",
		Name:   "server-addr",
		Usage:  "server address",
		Value:  ":3000",
	},
	cli.StringFlag{
		EnvVar: "INFOMARK_SECRET",
		Name:   "jwt-secret",
		Usage:  "secret phrase",
		Value:  "demo123",
	},
}

func server(c *cli.Context) error {
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}
	auth.InitializeJWT(c.String("jwt-secret"))
	logrus.Info("Infomark-server is listening at", c.String("server-addr"))

	r := router.GetRouter()
	http.ListenAndServe(c.String("server-addr"), r)

	return nil
}

func main() {
	logrus.Info("Start Infomark")
	app := cli.NewApp()
	app.Name = "infomark-server"
	app.Version = "0.0.1alpha"
	app.Usage = "infomark server"
	app.Action = server
	app.Flags = flags

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
