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

package background

import (
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "os"
  "strings"

  "github.com/cgtuebingen/infomark-backend/api/shared"
  "github.com/cgtuebingen/infomark-backend/service"
  "github.com/cgtuebingen/infomark-backend/tape"
  "github.com/google/uuid"
  "github.com/spf13/viper"
)

type SubmissionHandler interface {
  Handle(body []byte) error
}

type DummySubmissionHandler struct{}
type RealSubmissionHandler struct{}

var DefaultSubmissionHandler SubmissionHandler

func init() {
  fmt.Println(viper.GetString("rabbitmq_connection"))
  fmt.Println("worker_void", viper.GetBool("worker_void"))
  if viper.GetBool("worker_void") {
    DefaultSubmissionHandler = &DummySubmissionHandler{}
  } else {
    DefaultSubmissionHandler = &RealSubmissionHandler{}
  }
}

func (h *DummySubmissionHandler) Handle(workerBody []byte) error {
  // decode incoming message from AMQP
  msg := &shared.SubmissionAMQPWorkerRequest{}
  err := json.Unmarshal(workerBody, msg)

  if err != nil {
    return err
  }
  fmt.Println("msg.SubmissionID", msg.SubmissionID)
  fmt.Println("msg.AccessToken", msg.AccessToken)
  fmt.Println("msg.SubmissionFileURL", msg.SubmissionFileURL)
  fmt.Println("msg.FrameworkFileURL", msg.FrameworkFileURL)
  fmt.Println("msg.ResultEndpointURL", msg.ResultEndpointURL)

  fmt.Println("--> void")

  return nil
}

func downloadFile(r *http.Request, dst string) error {
  client := &http.Client{}

  w, err := client.Do(r)
  if err != nil {
    return err
  }
  defer w.Body.Close()

  out, err := os.Create(dst)
  if err != nil {
    return err
  }
  defer out.Close()

  _, err = io.Copy(out, w.Body)
  fmt.Println("write to", dst)
  return err
}

func cleanDockerOutput(stdout string) string {
  logStart := "--- BEGIN --- INFOMARK -- WORKER"
  logEnd := "--- END --- INFOMARK -- WORKER"

  rsl := strings.Split(stdout, logStart)
  rsl = strings.Split(rsl[1], logEnd)
  return rsl[0]
}

func (h *RealSubmissionHandler) Handle(body []byte) error {
  // HandleSubmission is responsible to
  // 1. parse request
  msg := &shared.SubmissionAMQPWorkerRequest{}
  err := json.Unmarshal(body, msg)
  if err != nil {
    return err
  }

  uuid, err := uuid.NewRandom()
  if err != nil {
    return err
  }

  // 2. fetch submission file from server
  r, err := http.NewRequest("GET", msg.SubmissionFileURL, nil)
  r.Header.Add("Authorization", "Bearer "+msg.AccessToken)
  if err := downloadFile(r, fmt.Sprintf("%s/%s-submission.zip", viper.GetString("worker_workdir"), uuid)); err != nil {
    return err
  }

  // 3. fetch framework file from server
  r, err = http.NewRequest("GET", msg.FrameworkFileURL, nil)
  r.Header.Add("Authorization", "Bearer "+msg.AccessToken)
  if err := downloadFile(r, fmt.Sprintf("%s/%s-framework.zip", viper.GetString("worker_workdir"), uuid)); err != nil {
    return err
  }

  // 2. run docker test

  ds := service.NewDockerService()
  defer ds.Client.Close()

  stdout, exit, err := ds.Run(
    msg.DockerImage,
    fmt.Sprintf("%s/%s-submission.zip", viper.GetString("worker_workdir"), uuid),
    fmt.Sprintf("%s/%s-framework.zip", viper.GetString("worker_workdir"), uuid),
  )
  if err != nil {
    return err
  }

  // fmt.Println("docker", stdout)
  stdout = cleanDockerOutput(stdout)
  // fmt.Println("docker", stdout)

  // fmt.Println("docker", exit)
  // fmt.Println("docker", err)

  // 3. push result back to server
  workerResp := &shared.SubmissionWorkerResponse{
    Log:    stdout,
    Status: int(exit),
  }

  // fmt.Println(workerResp.Log)
  // fmt.Println(workerResp.Status)

  // we use a HTTP Request to send the answer
  r = tape.BuildDataRequest("POST", msg.ResultEndpointURL, tape.ToH(workerResp))
  r.Header.Add("Authorization", "Bearer "+msg.AccessToken)

  // run request
  client := &http.Client{}
  resp, err := client.Do(r)
  if err != nil {
    return err
  }
  defer resp.Body.Close()
  fmt.Println("response Status:", resp.Status)

  // fmt.Println(resp.Body)

  return nil
}
