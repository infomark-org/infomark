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

package cmd

import (
  "encoding/json"
  "fmt"
  "log"
  "strconv"

  "github.com/cgtuebingen/infomark-backend/api/app"
  "github.com/cgtuebingen/infomark-backend/api/shared"
  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/service"
  "github.com/go-ozzo/ozzo-validation/is"
  "github.com/jmoiron/sqlx"
  "github.com/spf13/cobra"
  "github.com/spf13/viper"
  null "gopkg.in/guregu/null.v3"
)

var ConsoleCmd = &cobra.Command{
  Use:   "console",
  Short: "infomark console commands",
}

func ConnectAndStores() (*sqlx.DB, *app.Stores, error) {

  db, err := sqlx.Connect("postgres", viper.GetString("database_connection"))
  if err != nil {
    return nil, nil, err
  }

  if err := db.Ping(); err != nil {
    return nil, nil, err
  }

  stores := app.NewStores(db)

  return db, stores, nil

}

var UserPromoteCmd = &cobra.Command{
  Use:   "promote [userID]",
  Short: "set gives an user global admin permission",
  Long:  `Will set the gobal root flag o true for a user bypassing all permission tests`,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {

    arg0, err := strconv.Atoi(args[0])
    if err != nil {
      fmt.Printf("cannot convert userID '%s' to int\n", args[0])
      return
    }
    userID := int64(arg0)

    _, stores, err := ConnectAndStores()
    if err != nil {
      panic(err)
    }

    user, err := stores.User.Get(userID)
    if err != nil {
      fmt.Printf("user with id %v not found\n", userID)
      return
    }
    user.Root = true
    stores.User.Update(user)

  },
}

var UserDowngradeCmd = &cobra.Command{
  Use:   "downgrade [userID]",
  Short: "removes global admin permission from a user",
  Long:  `Will set the gobal root flag to false for a user `,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {

    arg0, err := strconv.Atoi(args[0])
    if err != nil {
      fmt.Printf("cannot convert userID '%s' to int\n", args[0])
      return
    }
    userID := int64(arg0)

    _, stores, err := ConnectAndStores()
    if err != nil {
      panic(err)
    }

    user, err := stores.User.Get(userID)
    if err != nil {
      fmt.Printf("user with id %v not found\n", userID)
      return
    }
    user.Root = false
    stores.User.Update(user)

  },
}

var UserConfirmCmd = &cobra.Command{
  Use:   "confirm [userID]",
  Short: "confirms the email address manually",
  Long:  `Will run confirmation procedure for an user `,
  Args:  cobra.ExactArgs(1),
  Run: func(cmd *cobra.Command, args []string) {

    arg0, err := strconv.Atoi(args[0])
    if err != nil {
      fmt.Printf("cannot convert userID '%s' to int\n", args[0])
      return
    }
    userID := int64(arg0)

    _, stores, err := ConnectAndStores()
    if err != nil {
      panic(err)
    }

    user, err := stores.User.Get(userID)
    if err != nil {
      fmt.Printf("user with id %v not found\n", userID)
      return
    }
    user.ConfirmEmailToken = null.String{}
    stores.User.Update(user)
  },
}

var UserSetEmailCmd = &cobra.Command{
  Use:   "set-email [userID] [email]",
  Short: "will alter the email address",
  Long:  `Will change email address of an user without confirmation procedure`,
  Args:  cobra.ExactArgs(2),
  Run: func(cmd *cobra.Command, args []string) {

    arg0, err := strconv.Atoi(args[0])
    if err != nil {
      fmt.Printf("cannot convert userID '%s' to int\n", args[0])
      return
    }
    userID := int64(arg0)
    email := args[1]

    err = is.Email.Validate(email)
    if err != nil {
      fmt.Printf("email '%s' is not a valid email\n", email)
      return
    }

    _, stores, err := ConnectAndStores()
    if err != nil {
      panic(err)
    }

    user, err := stores.User.Get(userID)
    if err != nil {
      fmt.Printf("user with id %v not found\n", userID)
      return
    }
    user.Email = email
    stores.User.Update(user)
  },
}

var UserCmd = &cobra.Command{
  Use:   "user",
  Short: "Management of users",
}

// -----------------------------------------------------------------------------

// go build infomark.go && ./infomark console submission enqueue 24 10 24 "test_java_submission:v1"
var SubmissionEnqueueCmd = &cobra.Command{

  // ./infomark console submission enqueue 24 10 24 "test_java_submission:v1"
  // cp files/fixtures/unittest.zip files/uploads/tasks/24-public.zip
  // cp files/fixtures/submission.zip files/uploads/submissions/10.zip

  Use:   "enqueue [taskID] [submissionID] [gradeID] [dockerimage]",
  Short: "put submission into testing queue",
  Long:  `will enqueue a submission again into the testing queue`,
  Args:  cobra.ExactArgs(4),
  Run: func(cmd *cobra.Command, args []string) {

    arg0, err := strconv.Atoi(args[0])
    if err != nil {
      fmt.Printf("cannot convert userID '%s' to int\n", args[0])
      return
    }

    arg1, err := strconv.Atoi(args[1])
    if err != nil {
      fmt.Printf("cannot convert submissionID '%s' to int\n", args[1])
      return
    }

    arg2, err := strconv.Atoi(args[2])
    if err != nil {
      fmt.Printf("cannot convert gradeID '%s' to int\n", args[2])
      return
    }

    taskID := int64(arg0)
    submissionID := int64(arg1)
    gradeID := int64(arg2)
    dockerimage := args[3]

    log.Println("starting producer...")

    cfg := &service.Config{
      Connection:   viper.GetString("rabbitmq_connection"),
      Exchange:     viper.GetString("rabbitmq_exchange"),
      ExchangeType: viper.GetString("rabbitmq_exchangeType"),
      Queue:        viper.GetString("rabbitmq_queue"),
      Key:          viper.GetString("rabbitmq_key"),
      Tag:          "SimpleSubmission",
    }

    tokenManager, err := authenticate.NewTokenAuth()
    if err != nil {
      panic(err)
    }
    accessToken, err := tokenManager.CreateAccessJWT(
      authenticate.NewAccessClaims(1, true))
    if err != nil {
      panic(err)
    }

    request := &shared.SubmissionAMQPWorkerRequest{
      SubmissionID: submissionID,
      AccessToken:  accessToken,
      FrameworkFileURL: fmt.Sprintf("%s/api/v1/tasks/%s/private_file",
        viper.GetString("url"),
        strconv.FormatInt(taskID, 10)),
      SubmissionFileURL: fmt.Sprintf("%s/api/v1/submissions/%s/file",
        viper.GetString("url"),
        strconv.FormatInt(submissionID, 10)),
      ResultEndpointURL: fmt.Sprintf("%s/api/v1/grades/%s/private_result",
        viper.GetString("url"),
        strconv.FormatInt(gradeID, 10)),
      DockerImage: dockerimage,
    }

    body, err := json.Marshal(request)
    if err != nil {
      fmt.Errorf("json.Marshal: %s", err)
    }

    producer, _ := service.NewProducer(cfg)
    producer.Publish(body)

  },
}

var SubmissionCmd = &cobra.Command{
  Use:   "submission",
  Short: "Management of submission",
}

func init() {
  UserCmd.AddCommand(UserDowngradeCmd)
  UserCmd.AddCommand(UserPromoteCmd)
  UserCmd.AddCommand(UserSetEmailCmd)
  ConsoleCmd.AddCommand(UserCmd)

  SubmissionCmd.AddCommand(SubmissionEnqueueCmd)
  ConsoleCmd.AddCommand(SubmissionCmd)

  RootCmd.AddCommand(ConsoleCmd)
}
