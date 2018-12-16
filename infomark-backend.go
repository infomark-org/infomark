package main

import (
	"fmt"
	"github.com/cgtuebingen/infomark-backend/router"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"net/http"
	"os"
)

var flags = []cli.Flag{
	cli.BoolFlag{
		EnvVar: "INFOMARK_DEBUG",
		Name:   "debug",
		Usage:  "enable server debug mode",
	},
}

func server(c *cli.Context) error {
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	logrus.Debug("launch router")
	r := router.GetRouter()
	http.ListenAndServe(":3000", r)

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
