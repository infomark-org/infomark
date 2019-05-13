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
	"os"
	"os/signal"

	background "github.com/cgtuebingen/infomark-backend/api/worker"
	"github.com/cgtuebingen/infomark-backend/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Worker provides a background worker
type Worker struct {
	NumInstances int
}

// NewWorker creates and configures an background worker
func NewWorker(numInstances int) (*Worker, error) {
	log.Println("configuring worker...")
	return &Worker{NumInstances: numInstances}, nil
}

// Start runs ListenAndServe on the http.Worker with graceful shutdown.
func (srv *Worker) Start() {
	log.Println("starting Worker...")

	cfg := &service.Config{
		Connection:   viper.GetString("rabbitmq_connection"),
		Exchange:     "infomark-worker-exchange",
		ExchangeType: "direct",
		Queue:        "infomark-worker-submissions",
		Key:          viper.GetString("rabbitmq_key"),
		Tag:          "SimpleSubmission",
	}

	consumers := []*service.Consumer{}

	for i := 0; i < srv.NumInstances; i++ {
		log.WithFields(logrus.Fields{"instance": i}).Info("start")
		consumer, _ := service.NewConsumer(cfg, background.DefaultSubmissionHandler.Handle, i)
		deliveries, err := consumer.Setup()
		if err != nil {
			panic(err)
		}
		consumers = append(consumers, consumer)
		go consumers[i].HandleLoop(deliveries)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	log.Println("Shutting down Worker... Reason:", sig)

	for i := 0; i < srv.NumInstances; i++ {
		consumers[i].Shutdown()
	}

	log.Println("Worker gracefully stopped")
}
