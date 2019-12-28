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

package shared

import (
	"fmt"
	"time"
)

// SubmissionAMQPWorkerRequest is the message which is handed over to the background workers
type SubmissionAMQPWorkerRequest struct {
	SubmissionID      int64     `json:"submission_id"`
	AccessToken       string    `json:"access_token"`
	FrameworkFileURL  string    `json:"framework_file_url"`
	SubmissionFileURL string    `json:"submission_file_url"`
	ResultEndpointURL string    `json:"result_endpoint_url"`
	DockerImage       string    `json:"docker_image"`
	Sha256            string    `json:"sha_256"`
	EnqueuedAt        time.Time `json:"enqueued_at"`
}

// // SubmissionWorkerResponse is the message handed from the workers to the server
// type SubmissionWorkerResponse struct {
// 	Log        string    `json:"log"`
// 	Status     int       `json:"status"`
// 	EnqueuedAt time.Time `json:"enqueued_at"`
// 	StartedAt  time.Time `json:"started_at"`
// 	FinishedAt time.Time `json:"finished_at"`
// }

// NewSubmissionAMQPWorkerRequest creates a new message for the workers
func NewSubmissionAMQPWorkerRequest(
	courseID int64, taskID int64, submissionID int64, gradeID int64,
	accessToken string, url string, dockerimage string, sha256 string, visibility string) *SubmissionAMQPWorkerRequest {

	return &SubmissionAMQPWorkerRequest{
		SubmissionID: submissionID,
		EnqueuedAt:   time.Now(),
		AccessToken:  accessToken,
		FrameworkFileURL: fmt.Sprintf("%s/api/v1/courses/%d/tasks/%d/%s_file",
			url,
			courseID,
			taskID,
			visibility),
		SubmissionFileURL: fmt.Sprintf("%s/api/v1/courses/%d/submissions/%d/file",
			url,
			courseID,
			submissionID),
		ResultEndpointURL: fmt.Sprintf("%s/api/v1/courses/%d/grades/%d/%s_result",
			url,
			courseID,
			gradeID,
			visibility),
		DockerImage: dockerimage,
		Sha256:      sha256,
	}
}
