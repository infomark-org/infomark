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
  "io/ioutil"
  "net/http"

  "github.com/cgtuebingen/infomark-backend/api/shared"
  "github.com/cgtuebingen/infomark-backend/tape"
)

type SubmissionHandler interface {
  Handle(body []byte) error
}

type DummySubmissionHandler struct{}
type RealSubmissionHandler struct{}

var DefaultSubmissionHandler SubmissionHandler

func init() {
  DefaultSubmissionHandler = &DummySubmissionHandler{}
}

func (h *DummySubmissionHandler) Handle(workerBody []byte) error {
  // decode incoming message from AMQP
  msg := &shared.SubmissionAMQPWorkerRequest{}
  err := json.Unmarshal(workerBody, msg)

  if err != nil {
    return err
  }
  fmt.Println("msg.SubmissionID", msg.SubmissionID)
  fmt.Println("msg.Token", msg.Token)
  fmt.Println("msg.FileURL", msg.FileURL)
  fmt.Println("msg.ResultURL", msg.ResultURL)

  // generate answer
  workerResp := &shared.SubmissionWorkerResponse{
    Log:    "some data",
    Status: 1,
  }

  // we use a HTTP Request to send the answer
  req := tape.BuildDataRequest("POST", msg.ResultURL, tape.ToH(workerResp))
  req.Header.Add("Authorization", "Bearer "+msg.Token)

  // run request
  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()

  fmt.Println("response Status:", resp.Status)
  fmt.Println("response Headers:", resp.Header)
  body, _ := ioutil.ReadAll(resp.Body)
  fmt.Println("response Body:", string(body))

  return nil
}

func (h *RealSubmissionHandler) Handle(body []byte) error {
  // HandleSubmission is responsible to
  // 1. fecth file from server
  // 2. run docker test
  // 3. push result back to server
  return nil
}
