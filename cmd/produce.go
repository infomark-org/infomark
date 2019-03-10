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

	"github.com/cgtuebingen/infomark-backend/api/shared"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// go build infomark.go && ./infomark produce 24 10 24 "test_java_submission:v1"

var produceCmd = &cobra.Command{
	Use:   "produce",
	Short: "start a producer",
	Long:  `produce some workload`,
	Args:  cobra.ExactArgs(4), // taskID, submissionID, gradeID, dockerImage
	Run: func(cmd *cobra.Command, args []string) {

		arg0, err := strconv.Atoi(args[0])
		if err != nil {
			panic(err)
		}

		arg1, err := strconv.Atoi(args[1])
		if err != nil {
			panic(err)
		}

		arg2, err := strconv.Atoi(args[2])
		if err != nil {
			panic(err)
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

func init() {
	RootCmd.AddCommand(produceCmd)
}
