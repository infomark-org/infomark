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

package api

import (
  "encoding/json"
  "net/http"
  "time"

  "github.com/cgtuebingen/infomark-backend/api/app"
  "github.com/cgtuebingen/infomark-backend/logging"
  "github.com/go-chi/chi"
  "github.com/go-chi/chi/middleware"
  "github.com/go-chi/cors"
  "github.com/go-chi/render"
  "github.com/jmoiron/sqlx"
  _ "github.com/lib/pq"
)

type H map[string]interface{}

func WriteJSON(w http.ResponseWriter, obj interface{}) error {
  // writeContentType(w, []string{"application/json; charset=utf-8"})
  jsonBytes, err := json.Marshal(obj)
  if err != nil {
    return err
  }
  w.Write(jsonBytes)
  return nil
}

func EmptyHandler(w http.ResponseWriter, r *http.Request) {
  WriteJSON(w, H{"response": "empty"})
}

// New configures application resources and routes.
func New() (*chi.Mux, error) {
  logger := logging.NewLogger()

  // db, err := sqlx.Connect("sqlite3", "__deleteme.db")
  db, err := sqlx.Connect("postgres", "user=postgres dbname=infomark password=postgres sslmode=disable")
  if err != nil {
    logger.WithField("module", "database").Error(err)
    return nil, err
  }

  appAPI, err := app.NewAPI(db)
  if err != nil {
    logger.WithField("module", "app").Error(err)
    return nil, err
  }

  r := chi.NewRouter()
  r.Use(middleware.Recoverer)
  r.Use(middleware.RequestID)
  // TODO (patwie): This overrides the status code
  // r.Use(middleware.DefaultCompress)
  r.Use(middleware.Timeout(15 * time.Second))
  r.Use(logging.NewStructuredLogger(logger))
  r.Use(render.SetContentType(render.ContentTypeJSON))
  r.Use(corsConfig().Handler)

  r.Route("/v1", func(r chi.Router) {
    // users
    r.Route("/users", func(r chi.Router) {
      r.Get("/", appAPI.User.Index)
      r.Route("/{userID}", func(r chi.Router) {
        r.Use(appAPI.User.Context)
        r.Get("/", appAPI.User.Get)
        r.Patch("/", appAPI.User.Patch)
      })
    })
    // login
    r.Route("/account", func(r chi.Router) {
      r.Get("/", EmptyHandler)
      r.Patch("/", EmptyHandler)
      r.Post("/", EmptyHandler)
    })

    // r.Route("/account", func(r chi.Router) {
    //   r.Get("/", EmptyHandler)
    //   r.Patch("/", EmptyHandler)
    //   r.Post("/", EmptyHandler)
    // })

  })

  r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("pong"))
  })

  return r, nil
}

func corsConfig() *cors.Cors {
  // Basic CORS
  // for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
  return cors.New(cors.Options{
    // AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
    AllowedOrigins: []string{"*"},
    // AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
    AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: true,
    MaxAge:           86400, // Maximum value not ignored by any of major browsers
  })
}
