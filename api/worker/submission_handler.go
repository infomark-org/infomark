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
  "crypto/sha256"
  "encoding/json"
  "fmt"
  "io"
  "net/http"
  "os"
  "strings"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/api/shared"
  "github.com/cgtuebingen/infomark-backend/logging"
  "github.com/cgtuebingen/infomark-backend/service"
  "github.com/cgtuebingen/infomark-backend/tape"
  "github.com/google/uuid"
  "github.com/sirupsen/logrus"
  "github.com/spf13/viper"
)

type SubmissionHandler interface {
  Handle(body []byte) error
}

type DummySubmissionHandler struct{}
type RealSubmissionHandler struct{}

var DefaultSubmissionHandler SubmissionHandler
var DefaultLogger *logrus.Logger

func init() {
  // fmt.Println(viper.GetString("rabbitmq_connection"))
  // fmt.Println("worker_void", viper.GetBool("worker_void"))
  if viper.GetBool("worker_void") {
    DefaultSubmissionHandler = &DummySubmissionHandler{}
  } else {
    DefaultSubmissionHandler = &RealSubmissionHandler{}
  }

  DefaultLogger = logging.NewLogger()
  file, err := os.OpenFile("submission_handler.log", os.O_APPEND|os.O_WRONLY, 0666)
  if err == nil {
    DefaultLogger.Out = file
  } else {
    DefaultLogger.Info("Failed to log to file, using default stderr")
  }

}

func (h *DummySubmissionHandler) Handle(workerBody []byte) error {
  // decode incoming message from AMQP
  msg := &shared.SubmissionAMQPWorkerRequest{}
  err := json.Unmarshal(workerBody, msg)

  if err != nil {
    DefaultLogger.WithFields(logrus.Fields{
      "SubmissionID":      msg.SubmissionID,
      "AccessToken":       msg.AccessToken,
      "SubmissionFileURL": msg.SubmissionFileURL,
      "FrameworkFileURL":  msg.FrameworkFileURL,
      "ResultEndpointURL": msg.ResultEndpointURL,
      "Sha256":            msg.Sha256,
    }).Warn(err)
    return err
  }

  fmt.Println("--> void")

  return nil
}

func verifySha256(filePath string, expectedChecksum string) error {
  f, err := os.Open(filePath)
  if err != nil {
    DefaultLogger.Printf("error: %v\n", err)
    return err
  }
  defer f.Close()

  h := sha256.New()
  if _, err := io.Copy(h, f); err != nil {
    DefaultLogger.Printf("error: %v\n", err)
    return err
  }

  actualChecksum := fmt.Sprintf("%x", h.Sum(nil))

  if actualChecksum != expectedChecksum {
    return fmt.Errorf("Sha256 missmatch, actual %s vs. expected %s for file %s",
      actualChecksum,
      expectedChecksum,
      filePath,
    )
  }

  return nil
}

func downloadFile(r *http.Request, dst string) error {
  client := &http.Client{}

  w, err := client.Do(r)
  if err != nil {
    DefaultLogger.Printf("error: %v\n", err)
    return err
  }
  defer w.Body.Close()

  out, err := os.Create(dst)
  if err != nil {
    DefaultLogger.Printf("error: %v\n", err)
    return err
  }
  defer out.Close()

  _, err = io.Copy(out, w.Body)
  return err
}

func cleanDockerOutput(stdout string) string {
  logStart := "--- BEGIN --- INFOMARK -- WORKER"
  logEnd := "--- END --- INFOMARK -- WORKER"

  rsl := strings.Split(stdout, logStart)

  if len(rsl) > 1 {
    rsl = strings.Split(rsl[1], logEnd)
    return rsl[0]

  } else {
    return stdout
  }
}

func (h *RealSubmissionHandler) Handle(body []byte) error {
  // HandleSubmission is responsible to
  // 1. parse request
  msg := &shared.SubmissionAMQPWorkerRequest{}
  err := json.Unmarshal(body, msg)
  if err != nil {
    DefaultLogger.Printf("error: %v\n", err)
    return err
  }

  DefaultLogger.WithFields(logrus.Fields{
    "SubmissionID":      msg.SubmissionID,
    "AccessToken":       msg.AccessToken,
    "SubmissionFileURL": msg.SubmissionFileURL,
    "FrameworkFileURL":  msg.FrameworkFileURL,
    "ResultEndpointURL": msg.ResultEndpointURL,
    "Sha256":            msg.Sha256,
  }).Info("handle")

  uuid, err := uuid.NewRandom()
  if err != nil {
    DefaultLogger.Printf("error: %v\n", err)
    return err
  }
  submission_path := fmt.Sprintf("%s/%s-submission.zip", viper.GetString("worker_workdir"), uuid)
  framework_path := fmt.Sprintf("%s/%s-framework.zip", viper.GetString("worker_workdir"), uuid)

  // 2. fetch submission file from server
  r, err := http.NewRequest("GET", msg.SubmissionFileURL, nil)
  r.Header.Add("Authorization", "Bearer "+msg.AccessToken)
  if err := downloadFile(r, submission_path); err != nil {
    DefaultLogger.Printf("error: %v\n", err)
    return err
  }

  defer helper.FileDelete(submission_path)

  // 3. fetch framework file from server
  r, err = http.NewRequest("GET", msg.FrameworkFileURL, nil)
  r.Header.Add("Authorization", "Bearer "+msg.AccessToken)
  if err := downloadFile(r, framework_path); err != nil {
    DefaultLogger.Printf("error: %v\n", err)
    return err
  }
  defer helper.FileDelete(framework_path)

  // we use a HTTP Request to send the answer
  r = tape.BuildDataRequest("POST", msg.ResultEndpointURL, tape.ToH(&shared.SubmissionWorkerResponse{
    Log:    "submission is currently being tested ...",
    Status: 0,
  }))
  r.Header.Add("Authorization", "Bearer "+msg.AccessToken)

  // 4. verify checksums to avoid race conditions
  if err := verifySha256(submission_path, msg.Sha256); err != nil {
    DefaultLogger.WithFields(logrus.Fields{
      "SubmissionID":      msg.SubmissionID,
      "SubmissionFileURL": msg.SubmissionFileURL,
      "FrameworkFileURL":  msg.FrameworkFileURL,
      "Sha256":            msg.Sha256,
    }).Warn(err)
    return err
  }

  // 5. run docker test
  ds := service.NewDockerService()
  defer ds.Client.Close()

  stdout, exit, err := ds.Run(msg.DockerImage, submission_path, framework_path, viper.GetInt64("worker_docker_memory_bytes"))
  if err != nil {
    DefaultLogger.WithFields(logrus.Fields{
      "SubmissionID": msg.SubmissionID,
      "stdout":       stdout,
      "exitcode":     exit,
    }).Warn(err)
    return err
  }

  stdout = cleanDockerOutput(stdout)

  // 3. push result back to server
  workerResp := &shared.SubmissionWorkerResponse{
    Log:    stdout,
    Status: int(exit),
  }

  // we use a HTTP Request to send the answer
  r = tape.BuildDataRequest("POST", msg.ResultEndpointURL, tape.ToH(workerResp))
  r.Header.Add("Authorization", "Bearer "+msg.AccessToken)

  // run request
  client := &http.Client{}
  resp, err := client.Do(r)
  if err != nil {
    DefaultLogger.WithFields(logrus.Fields{
      "action":       "send result to backend",
      "SubmissionID": msg.SubmissionID,
      "stdout":       stdout,
      "exitcode":     exit,
    }).Warn(err)

    return err
  }
  defer resp.Body.Close()

  return nil
}
