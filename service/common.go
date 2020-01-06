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

package service

import (
	"os"

	"github.com/infomark-org/infomark/configuration"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// Worker is a description for an object getting messages over AMPQ
type Worker interface {
	Setup() error
	Shutdown() error
	HandleLoop(deliveries <-chan amqp.Delivery)
}

// Config contains the settings for AMPQ
type Config struct {
	Tag          string
	Connection   string
	Exchange     string
	ExchangeType string
	Queue        string
	Key          string
}

func NewConfig(config *configuration.RabbitMQConfiguration) *Config {
	return &Config{

		Tag:          "SimpleSubmission",
		Connection:   config.URL(),
		Exchange:     "infomark-worker-exchange",
		ExchangeType: "direct",
		Queue:        "infomark-worker-submissions",
		Key:          config.Key,
	}
}

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.Out = os.Stdout
}
