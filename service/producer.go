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

package service

import (
  "fmt"
  "log"

  "github.com/streadway/amqp"
)

// Producer is an object which can emit a AMPQ messages
type Producer struct {
  Config *Config

  conn    *amqp.Connection
  channel *amqp.Channel
  done    chan error
}

// NewProducer creates a new producer which can emit AMPQ messages
func NewProducer(cfg *Config) (*Producer, error) {

  producer := &Producer{
    conn:    nil,
    channel: nil,
    done:    make(chan error),

    Config: cfg,
  }

  return producer, nil
}

// Publish emits an AMPQ message
func (c *Producer) Publish(body []byte) error {

  // This function dials, connects, declares, publishes, and tears down,
  // all in one go. In a real service, you probably want to maintain a
  // long-lived connection as state, and publish against that.

  log.Printf("dialing %s", c.Config.Connection)
  connection, err := amqp.Dial(c.Config.Connection)
  if err != nil {
    return fmt.Errorf("Dial: %s", err)
  }
  defer connection.Close()

  log.Printf("got Connection, getting Channel")
  channel, err := connection.Channel()
  if err != nil {
    return fmt.Errorf("Channel: %s", err)
  }

  log.Printf("got Channel, declaring %q Exchange (%s)", c.Config.ExchangeType, c.Config.Exchange)
  if err := channel.ExchangeDeclare(
    c.Config.Exchange,     // name
    c.Config.ExchangeType, // type
    false,                 // durable
    false,                 // auto-deleted
    false,                 // internal
    false,                 // noWait
    nil,                   // arguments
  ); err != nil {
    return fmt.Errorf("Exchange Declare: %s", err)
  }

  // Prepare this message to be persistent.  Your publishing requirements may
  // be different.
  msg := amqp.Publishing{
    Headers:         amqp.Table{},
    ContentType:     "application/json",
    ContentEncoding: "",
    Body:            body,
    DeliveryMode:    1, // 1=non-persistent, 2=persistent
    Priority:        0, // 0-9
  }

  log.Printf("declared Exchange, publishing %dB body (%s)", len(body), body)
  if err = channel.Publish(
    c.Config.Exchange, // publish to an exchange
    c.Config.Key,      // routing to 0 or more queues
    false,             // mandatory
    false,             // immediate
    msg,
  ); err != nil {
    return fmt.Errorf("Exchange Publish: %s", err)
  }

  return nil

}
