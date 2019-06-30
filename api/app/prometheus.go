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

package app

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	totalFailedLoginsVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "auth",
			Subsystem: "logins",
			Name:      "failed_logins",
			Help:      "Total number of failed logins",
		},
		//
		[]string{},
	)

	totalSubmissionCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "worker",
			Subsystem: "submissions",
			Name:      "pushed_total",
			Help:      "Total number of submissions pushed to the server",
		},
		//
		[]string{"task_id"},
	)

	totalDockerFailExitCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "worker",
			Subsystem: "submissions",
			Name:      "failed_total",
			Help:      "Total number of submissions where docker has unsuccessful exit status",
		},
		//
		[]string{"task_id", "kind"},
	)

	totalDockerSuccessExitCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "worker",
			Subsystem: "submissions",
			Name:      "success_total",
			Help:      "Total number of submissions where docker has successful exit status",
		},
		//
		[]string{"task_id", "kind"},
	)

	totalDockerTimeHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "worker",
			Subsystem: "submissions",
			Name:      "totalTime",
			Help:      "Total time in seconds taken from received upload to finished docker run",
		},
		//
		[]string{"task_id", "kind"},
	)

	totalDockerRunTimeHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "worker",
			Subsystem: "submissions",
			Name:      "totalRunTime",
			Help:      "Total time in seconds taken spent inside docker when running tests",
		},
		//
		[]string{"task_id", "kind"},
	)

	totalDockerWaitTimeHist = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "worker",
			Subsystem: "submissions",
			Name:      "totalWaitTime",
			Help:      "Total time waited between uploading and fetched from worker",
		},
		//
		[]string{"task_id", "kind"},
	)
)

var prometheusIsRegistered = false

func InitPrometheus() {
	if !prometheusIsRegistered {
		// register with the prometheus collector
		prometheus.MustRegister(totalSubmissionCounterVec)
		prometheus.MustRegister(totalDockerFailExitCounterVec)
		prometheus.MustRegister(totalDockerSuccessExitCounterVec)
		prometheus.MustRegister(totalFailedLoginsVec)
		prometheus.MustRegister(totalDockerTimeHist)
		prometheus.MustRegister(totalDockerRunTimeHist)
		prometheus.MustRegister(totalDockerWaitTimeHist)
		prometheusIsRegistered = true
	}
}
