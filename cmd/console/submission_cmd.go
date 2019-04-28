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

package console

import (
  "encoding/json"
  "fmt"
  "log"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/api/shared"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/service"
  "github.com/spf13/cobra"
  "github.com/spf13/viper"
)

func init() {
  SubmissionCmd.AddCommand(SubmissionEnqueueCmd)
  SubmissionCmd.AddCommand(SubmissionRunCmd)

}

var SubmissionCmd = &cobra.Command{
  Use:   "submission",
  Short: "Management of submission",
}

var SubmissionEnqueueCmd = &cobra.Command{
  Use:   "enqeue [submissionID]",
  Short: "put submission into testing queue",
  Long:  `will enqueue a submission again into the testing queue`,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {
    submissionID := MustInt64Parameter(args[0], "submissionID")

    _, stores := MustConnectAndStores()

    submission, err := stores.Submission.Get(submissionID)
    fail(err)

    task, err := stores.Task.Get(submission.TaskID)
    fail(err)

    sheet, err := stores.Task.IdentifySheetOfTask(submission.TaskID)
    fail(err)

    course, err := stores.Sheet.IdentifyCourseOfSheet(sheet.ID)
    fail(err)

    grade, err := stores.Grade.GetForSubmission(submission.ID)
    fail(err)

    log.Println("starting producer...")

    cfg := &service.Config{
      Connection:   viper.GetString("rabbitmq_connection"),
      Exchange:     viper.GetString("rabbitmq_exchange"),
      ExchangeType: viper.GetString("rabbitmq_exchangeType"),
      Queue:        viper.GetString("rabbitmq_queue"),
      Key:          viper.GetString("rabbitmq_key"),
      Tag:          "SimpleSubmission",
    }

    sha256, err := helper.NewSubmissionFileHandle(submission.ID).Sha256()
    fail(err)

    tokenManager, err := authenticate.NewTokenAuth()
    fail(err)
    accessToken, err := tokenManager.CreateAccessJWT(
      authenticate.NewAccessClaims(1, true))
    fail(err)

    body_public, err := json.Marshal(shared.NewSubmissionAMQPWorkerRequest(
      course.ID, task.ID, submission.ID, grade.ID,
      accessToken, viper.GetString("url"), task.PublicDockerImage.String, sha256, "public"))
    if err != nil {
      log.Fatalf("json.Marshal: %s", err)
    }

    body_private, err := json.Marshal(shared.NewSubmissionAMQPWorkerRequest(
      course.ID, task.ID, submission.ID, grade.ID,
      accessToken, viper.GetString("url"), task.PrivateDockerImage.String, sha256, "private"))
    if err != nil {
      log.Fatalf("json.Marshal: %s", err)
    }

    producer, _ := service.NewProducer(cfg)
    producer.Publish(body_public)
    producer.Publish(body_private)

  },
}

var SubmissionRunCmd = &cobra.Command{
  Use:   "run [submissionID]",
  Short: "run tests for a submission without writing to db",
  Long:  `will enqueue a submission again into the testing queue`,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {

    submissionID := MustInt64Parameter(args[0], "submissionID")

    _, stores := MustConnectAndStores()

    submission, err := stores.Submission.Get(submissionID)
    fail(err)

    task, err := stores.Task.Get(submission.TaskID)
    fail(err)

    log.Println("starting docker...")

    ds := service.NewDockerService()
    defer ds.Client.Close()

    var exit int64
    var stdout string

    submission_hnd := helper.NewSubmissionFileHandle(submission.ID)
    if !submission_hnd.Exists() {
      log.Fatalf("submission file %s for id %v is missing", submission_hnd.Path(), submission.ID)
    }

    // run public test
    if task.PublicDockerImage.Valid {
      framework_hnd := helper.NewPublicTestFileHandle(task.ID)
      if framework_hnd.Exists() {

        log.Printf("use docker image \"%v\"\n", task.PublicDockerImage.String)
        log.Printf("use framework file \"%v\"\n", framework_hnd.Path())
        stdout, exit, err = ds.Run(
          task.PublicDockerImage.String,
          submission_hnd.Path(),
          framework_hnd.Path(),
          viper.GetInt64("worker_docker_memory_bytes"),
        )
        if err != nil {
          log.Fatal(err)
        }

        fmt.Println(" --- STDOUT -- BEGIN ---")
        fmt.Println(stdout)
        fmt.Println(" --- STDOUT -- END   ---")
        fmt.Printf("exit-code: %v\n", exit)
      } else {
        fmt.Println("skip public test, there is no framework file")

      }

    } else {
      fmt.Println("skip public test, there is no docker file")
    }

    // run private test
    if task.PrivateDockerImage.Valid {
      framework_hnd := helper.NewPrivateTestFileHandle(task.ID)
      if framework_hnd.Exists() {

        log.Printf("use docker image \"%v\"\n", task.PrivateDockerImage.String)
        log.Printf("use framework file \"%v\"\n", framework_hnd.Path())
        stdout, exit, err = ds.Run(
          task.PrivateDockerImage.String,
          submission_hnd.Path(),
          framework_hnd.Path(),
          viper.GetInt64("worker_docker_memory_bytes"),
        )
        if err != nil {
          log.Fatal(err)
        }

        fmt.Println(" --- STDOUT -- BEGIN ---")
        fmt.Println(stdout)
        fmt.Println(" --- STDOUT -- END   ---")
        fmt.Printf("exit-code: %v\n", exit)
      } else {
        fmt.Println("skip private test, there is no framework file")

      }

    } else {
      fmt.Println("skip private test, there is no docker file")
    }

  },
}
