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

package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/cgtuebingen/infomark-backend/api/app"
	"github.com/cgtuebingen/infomark-backend/api/cronjob"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var log *logrus.Logger

func RunInit() {
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.Out = os.Stdout

	app.RunInit()
}

// Server provides an http.Server.
type Server struct {
	HTTP *http.Server
	Cron *cron.Cron
}

// NewServer creates and configures an APIServer serving all application routes.
func NewServer() (*Server, error) {
	RunInit()
	app.PrepareTokenManager()
	app.InitSubmissionProducer()
	log.Info("configuring server...")

	if viper.GetString("sendmail_binary") != "" {
		log.WithFields(logrus.Fields{"path": viper.GetString("sendmail_binary")}).Info("found sendmail")
		email.DefaultMail = email.SendMail
	}

	db, err := sqlx.Connect("postgres", viper.GetString("database_connection"))
	if err != nil {
		log.WithField("module", "database").Error(err)
		return nil, err
	}

	handler, err := app.New(db, true)
	if err != nil {
		return nil, err
	}

	handler.Handle("/metrics", promhttp.Handler())

	var addr string
	port := viper.GetString("port")

	// allow port to be set as localhost:3000 in env during development to avoid "accept incoming network connection" request on restarts
	if strings.Contains(port, ":") {
		addr = port
	} else {
		addr = ":" + port
	}

	srv := http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    time.Duration(viper.GetInt64("server_read_timeout_sec")) * time.Second,
		WriteTimeout:   time.Duration(viper.GetInt64("server_write_timeout_sec")) * time.Second,
		MaxHeaderBytes: viper.GetInt("server_write_timeout_sec"),
	}

	c := cron.New()
	c.AddJob(fmt.Sprintf("@%s", viper.GetString("cronjob_intervall_submission_zip")), &cronjob.SubmissionFileZipper{
		Stores:    app.NewStores(db),
		DB:        db,
		Directory: viper.GetString("generated_files_dir"),
	})

	return &Server{HTTP: &srv, Cron: c}, nil
}

// Start runs ListenAndServe on the http.Server with graceful shutdown.
func (srv *Server) Start() {
	// log := logrus.StandardLogger()
	log.Info("starting server...")
	go func() {
		if err := srv.HTTP.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}
	}()
	log.WithFields(logrus.Fields{
		"addr": srv.HTTP.Addr,
	}).Info("http is listening")

	log.Info("starting background email sender...")
	go email.BackgroundSend(email.OutgoingEmailsChannel)

	log.Info("starting cronjob for zipping submissions...")
	srv.Cron.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	log.Info("Shutting down server... Reason:", sig)

	// teardown logic...
	srv.Cron.Stop()
	log.Info("Cronjobs gracefully stopped")

	close(email.OutgoingEmailsChannel)
	log.Info("Background email sender gracefully stopped")

	if err := srv.HTTP.Shutdown(context.Background()); err != nil {
		panic(err)
	}
	log.Info("Server gracefully stopped")
}
