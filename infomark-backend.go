package main

//go:generate sqlboiler --wipe psql

import (
	"fmt"
	"github.com/cgtuebingen/infomark-backend/router"
	_ "github.com/lib/pq"
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
	cli.StringFlag{
		EnvVar: "INFOMARK_SERVER_ADDR",
		Name:   "server-addr",
		Usage:  "server address",
		Value:  ":3000",
	},
}

func server(c *cli.Context) error {
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

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
