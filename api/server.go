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

package api

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"github.com/infomark-org/infomark/api/app"
	"github.com/infomark-org/infomark/api/cronjob"
	"github.com/infomark-org/infomark/auth/authenticate"
	"github.com/infomark-org/infomark/configuration"
	"github.com/infomark-org/infomark/email"
	"github.com/infomark-org/infomark/migration"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
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
	HTTP           *http.Server
	Cron           *cron.Cron
	Configuration  *configuration.ServerConfigurationSchema
	Authentication *authenticate.TokenAuth
}

// NewServer creates and configures an APIServer serving all application routes.
func NewServer(config *configuration.ServerConfigurationSchema) (*Server, error) {
	RunInit()

	app.InitSubmissionProducer()
	log.WithField("url", config.URL()).Info("configuring server...")

	if config.SendEmail() {
		log.WithFields(logrus.Fields{"path": config.Email.SendmailBinary}).Info("found sendmail")
		email.SendMail = email.NewSendMailer(config.Email.SendmailBinary)
		email.DefaultMail = email.SendMail
	} else {
		email.DefaultMail = email.TerminalMail
	}

	db, err := sqlx.Connect("postgres", config.PostgresURL())
	if err != nil {
		log.WithField("module", "database").Error(err)
		return nil, err
	}

	migration.UpdateDatabase(db, log)

	handler, err := app.New(db, promhttp.Handler(), true)
	if err != nil {
		return nil, err
	}

	srv := http.Server{
		Addr:           config.HTTPAddr(),
		Handler:        handler,
		ReadTimeout:    config.HTTP.Timeouts.Read,
		WriteTimeout:   config.HTTP.Timeouts.Write,
		MaxHeaderBytes: int(config.HTTP.Limits.MaxHeader),
	}

	c := cron.New()
	c.AddJob(config.CronjobsZipSubmissionsIntervall(), &cronjob.SubmissionFileZipper{
		Stores:    app.NewStores(db),
		DB:        db,
		Directory: config.Paths.GeneratedFiles,
	})

	return &Server{
		HTTP:           &srv,
		Cron:           c,
		Configuration:  config,
		Authentication: authenticate.NewTokenAuth(&config.Authentication)}, nil
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
