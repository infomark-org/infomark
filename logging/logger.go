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

package logging

import (
  "fmt"
  "log"
  "net/http"
  "time"

  "github.com/go-chi/chi/middleware"
  "github.com/sirupsen/logrus"
  "github.com/spf13/viper"
)

var (
  // Logger is a configured logrus.Logger.
  Logger *logrus.Logger
)

// StructuredLogger is a structured logrus Logger.
type StructuredLogger struct {
  Logger *logrus.Logger
}

// NewLogger creates and configures a new logrus Logger.
func NewLogger() *logrus.Logger {
  Logger = logrus.New()
  if viper.GetBool("log_textlogging") {
    Logger.Formatter = &logrus.TextFormatter{
      DisableTimestamp: true,
    }
  } else {
    Logger.Formatter = &logrus.JSONFormatter{
      DisableTimestamp: true,
    }
  }

  level := viper.GetString("log_level")
  if level == "" {
    level = "error"
  }
  l, err := logrus.ParseLevel(level)
  if err != nil {
    log.Fatal(err)
  }
  Logger.Level = l
  return Logger
}

// NewStructuredLogger implements a custom structured logrus Logger.
func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
  return middleware.RequestLogger(&StructuredLogger{Logger})
}

// NewLogEntry sets default request log fields.
func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
  entry := &StructuredLoggerEntry{Logger: logrus.NewEntry(l.Logger)}
  logFields := logrus.Fields{}

  logFields["ts"] = time.Now().UTC().Format(time.RFC1123)

  if reqID := middleware.GetReqID(r.Context()); reqID != "" {
    logFields["req_id"] = reqID
  }

  scheme := "http"
  if r.TLS != nil {
    scheme = "https"
  }

  logFields["http_scheme"] = scheme
  logFields["http_proto"] = r.Proto
  logFields["http_method"] = r.Method

  logFields["remote_addr"] = r.RemoteAddr
  logFields["user_agent"] = r.UserAgent()

  logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)
  logFields["uri"] = fmt.Sprintf("%s", r.RequestURI)

  entry.Logger = entry.Logger.WithFields(logFields)

  entry.Logger.Infoln("request started")

  return entry
}

// StructuredLoggerEntry is a logrus.FieldLogger.
type StructuredLoggerEntry struct {
  Logger logrus.FieldLogger
}

func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
  l.Logger = l.Logger.WithFields(logrus.Fields{
    "resp_status":       status,
    "resp_bytes_length": bytes,
    "resp_elapsed_ms":   float64(elapsed.Nanoseconds()) / 1000000.0,
  })

  l.Logger.Infoln("request complete")
}

// Panic prints stack trace
func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
  l.Logger = l.Logger.WithFields(logrus.Fields{
    "stack": string(stack),
    "panic": fmt.Sprintf("%+v", v),
  })
}

// Helper methods used by the application to get the request-scoped
// logger entry and set additional fields between handlers.

// GetLogEntry return the request scoped logrus.FieldLogger.
func GetLogEntry(r *http.Request) logrus.FieldLogger {
  entry := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
  return entry.Logger
}

// LogEntrySetField adds a field to the request scoped logrus.FieldLogger.
func LogEntrySetField(r *http.Request, key string, value interface{}) {
  if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
    entry.Logger = entry.Logger.WithField(key, value)
  }
}

// LogEntrySetFields adds multiple fields to the request scoped logrus.FieldLogger.
func LogEntrySetFields(r *http.Request, fields map[string]interface{}) {
  if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
    entry.Logger = entry.Logger.WithFields(fields)
  }
}
