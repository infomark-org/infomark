// Copyright 2019 ComputerGraphics Tuebingen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ==============================================================================
// Authors: Patrick Wieschollek

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
