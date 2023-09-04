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

package background

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/infomark-org/infomark/api/app"
	"github.com/infomark-org/infomark/api/helper"
	"github.com/infomark-org/infomark/api/shared"
	"github.com/infomark-org/infomark/configuration"
	"github.com/infomark-org/infomark/service"
	"github.com/infomark-org/infomark/symbol"
	"github.com/infomark-org/infomark/tape"
	"github.com/sirupsen/logrus"
)

// SubmissionHandler is any handler capable to work on submissions
type SubmissionHandler interface {
	Handle(body []byte) error
}

// DummySubmissionHandler is doing nothing (for testing)
type DummySubmissionHandler struct{}

// RealSubmissionHandler is starting docker to test submissions
type RealSubmissionHandler struct{}

// DefaultSubmissionHandler is the default submission handler
var DefaultSubmissionHandler SubmissionHandler

// DefaultLogger is the logger writing worker-logs to file
var DefaultLogger *logrus.Logger

func init() {

	DefaultLogger = logrus.StandardLogger()
	DefaultLogger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	DefaultLogger.Out = os.Stdout
	// file, err := os.OpenFile("submission_handler.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	// if err == nil {
	// 	DefaultLogger.Out = file
	// } else {
	// 	fmt.Println(err)
	// 	DefaultLogger.Info("Failed to log to file, using default stderr")
	// }
	// defer file.Close()

}

// Handle reads message and does nothing
func (h *DummySubmissionHandler) Handle(workerBody []byte) error {
	// decode incoming message from AMQP
	msg := &shared.SubmissionAMQPWorkerRequest{}
	err := json.Unmarshal(workerBody, msg)

	if err != nil {
		DefaultLogger.WithFields(logrus.Fields{
			"submissionID":      msg.SubmissionID,
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

func newHTTPClientSingleRequest() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: 30 * time.Second,
			DisableKeepAlives:   true,
		},
		Timeout: 1 * time.Minute,
	}
}

func downloadFile(r *http.Request, dst string) error {
	client := newHTTPClientSingleRequest()

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

	}
	return stdout
}

// Handle reads message and test submission using docker
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
		"submissionID": msg.SubmissionID,
		// "AccessToken":       msg.AccessToken,
		// "SubmissionFileURL": msg.SubmissionFileURL,
		"FrameworkFileURL": msg.FrameworkFileURL,
		"image":            msg.DockerImage,
		// "ResultEndpointURL": msg.ResultEndpointURL,
		"Sha256": msg.Sha256,
	}).Info("start processing")

	uuid, err := uuid.NewRandom()
	if err != nil {
		DefaultLogger.Printf("error: %v\n", err)
		return err
	}
	submissionPath := fmt.Sprintf("%s/%s-submission.zip", configuration.Configuration.Worker.Workdir, uuid)
	frameworkPath := fmt.Sprintf("%s/%s-framework.zip", configuration.Configuration.Worker.Workdir, uuid)

	// 2. fetch submission file from server
	r, err := http.NewRequest("GET", msg.SubmissionFileURL, nil)
	if err != nil {
		DefaultLogger.Printf("error: %v\n", err)
		return err
	}

	r.Header.Add("Authorization", "Bearer "+msg.AccessToken)
	if err := downloadFile(r, submissionPath); err != nil {
		DefaultLogger.Printf("error: %v\n", err)
		return err
	}

	defer helper.FileDelete(submissionPath)

	// 3. fetch framework file from server
	r, err = http.NewRequest("GET", msg.FrameworkFileURL, nil)
	if err != nil {
		DefaultLogger.Printf("error: %v\n", err)
		return err
	}
	r.Header.Add("Authorization", "Bearer "+msg.AccessToken)

	if err := downloadFile(r, frameworkPath); err != nil {
		DefaultLogger.Printf("error: %v\n", err)
		return err
	}
	defer helper.FileDelete(frameworkPath)

	// Under circumstances there is no guarantee that the following request will be issues
	// BEFORE the actual test result.
	// we use a HTTP Request to send the answer
	// r = tape.BuildDataRequest("POST", msg.ResultEndpointURL, tape.ToH(&app.GradeFromWorkerRequest{
	//   Log:    "submission is currently being tested ...",
	//   Status: 1,
	// }))
	// r.Header.Add("Authorization", "Bearer "+msg.AccessToken)
	// resp, err := client.Do(r)
	// if err != nil {
	//   DefaultLogger.WithFields(logrus.Fields{
	//     "action":       "send result to backend",
	//     "submissionID": msg.SubmissionID,
	//     "resp":         resp,
	//   }).Warn(err)

	//   return err
	// }

	// 4. verify checksums to avoid race conditions
	if err := verifySha256(submissionPath, msg.Sha256); err != nil {
		DefaultLogger.WithFields(logrus.Fields{
			"submissionID":      msg.SubmissionID,
			"SubmissionFileURL": msg.SubmissionFileURL,
			"FrameworkFileURL":  msg.FrameworkFileURL,
			"Sha256":            msg.Sha256,
		}).Warn(err)
		return err
	}

	// 5. run docker test
	ds, err := service.NewDockerServiceWithTimeout(configuration.Configuration.Worker.Docker.Timeout)
	if err != nil {
		DefaultLogger.Printf("error: %v\n", err)
		return err
	}
	defer ds.Client.Close()

	var exit int64
	var stdout string

	workerResp := &app.GradeFromWorkerRequest{}
	workerResp.EnqueuedAt = msg.EnqueuedAt
	workerResp.StartedAt = time.Now()

	stdout, exit, err = ds.Run(
		msg.DockerImage,
		submissionPath,
		frameworkPath,
		int64(configuration.Configuration.Worker.Docker.MaxMemory),
	)
	/*
	if err != nil {
		DefaultLogger.WithFields(logrus.Fields{
			"submissionID": msg.SubmissionID,
			"stdout":       stdout,
			"exitcode":     exit,
			"image":        msg.DockerImage,
		}).Warn(err)
		return err
	}
	*/

	if err == nil && exit == symbol.TestingResultSuccess.AsInt64() {
		stdout = cleanDockerOutput(stdout)
		// 3. push result back to server
		workerResp.Log = stdout
		workerResp.Status = symbol.TestingResultSuccess
		workerResp.FinishedAt = time.Now()

	} else {

		DefaultLogger.WithFields(logrus.Fields{
			"submissionID": msg.SubmissionID,
			"stdout":       stdout,
			"exitcode":     exit,
			"image":        msg.DockerImage,
		}).Warn(err)

		workerResp.Log = fmt.Sprintf(`There has been an issue during testing your upload (The ID is %v).
        The testing-framework has failed (not the server).\nError code: %v\n`,
			msg.SubmissionID, exit)
		workerResp.Status = symbol.TestingResultFailed
		workerResp.FinishedAt = time.Now()

	}

	// we use a HTTP Request to send the answer
	r = tape.BuildDataRequest("POST", msg.ResultEndpointURL, tape.ToH(workerResp))
	r.Header.Add("Authorization", "Bearer "+msg.AccessToken)

	DefaultLogger.WithFields(logrus.Fields{
		"submissionID":      msg.SubmissionID,
		"exitcode":          exit,
		"image":             msg.DockerImage,
		"resultEndpointURL": msg.ResultEndpointURL,
	}).Info("send result to backend")

	// run request
	client := newHTTPClientSingleRequest()
	resp, err := client.Do(r)
	if err != nil {
		DefaultLogger.WithFields(logrus.Fields{
			"action":            "send result to backend",
			"submissionID":      msg.SubmissionID,
			"ResultEndpointURL": msg.ResultEndpointURL,
			"stdout":            stdout,
			"exitcode":          exit,
			"resp":              resp,
			"image":             msg.DockerImage,
		}).Warn(err)

		return err
	}
	defer resp.Body.Close()

	return nil
}
