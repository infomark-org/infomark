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
  "errors"
  "fmt"
  "io"
  "net/http"
  "os"
  "path/filepath"
  "strings"
  "time"

  "github.com/cgtuebingen/infomark-backend/auth/authenticate"
  "github.com/cgtuebingen/infomark-backend/auth/authorize"
  "github.com/cgtuebingen/infomark-backend/logging"
  "github.com/go-chi/chi"
  "github.com/go-chi/chi/middleware"
  "github.com/go-chi/cors"
  "github.com/go-chi/render"
  "github.com/jmoiron/sqlx"
  _ "github.com/lib/pq"
  "github.com/spf13/viper"
)

// The golang fork-join multi-threading allows no easy way to cancel started request
// Therefore we limit the amount of data which is read by the server whenever
// we need to parse a JSON request.
func LimitedDecoder(r *http.Request, v interface{}) error {
  var err error

  switch render.GetRequestContentType(r) {
  case render.ContentTypeJSON:
    body := io.LimitReader(r.Body, viper.GetInt64("max_request_json_bytes"))
    err = render.DecodeJSON(body, v)
  default:
    err = errors.New("render: unable to automatically decode the request content type")
  }
  //
  return err
}

func init() {
  render.Decode = LimitedDecoder
}

// Version writes the current API version to the headers.
func VersionMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("X-INFOMARK-VERSION", "0.0.1")
    next.ServeHTTP(w, r)
  })
}

// Secure writes required access headers to all requests.
func SecureMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("X-Frame-Options", "DENY")
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("X-XSS-Protection", "1; mode=block")
    next.ServeHTTP(w, r)
  })
}

// NoCache writes required cache headers to all requests.
func NoCache(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate, value")
    w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
    w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))

    next.ServeHTTP(w, r)
  })
}

// New configures application resources and routes.
func New(db *sqlx.DB, log bool) (*chi.Mux, error) {
  logger := logging.NewLogger()

  if err := db.Ping(); err != nil {
    logger.WithField("module", "database").Error(err)
    return nil, err
  }

  appAPI, err := NewAPI(db)
  if err != nil {
    logger.WithField("module", "app").Error(err)
    return nil, err
  }

  r := chi.NewRouter()
  r.Use(VersionMiddleware)
  r.Use(SecureMiddleware)
  r.Use(NoCache)
  r.Use(middleware.Recoverer)
  r.Use(middleware.RequestID)
  // the following line does not make any sense
  // r.Use(middleware.Timeout(15 * time.Second))
  if log {
    r.Use(logging.NewStructuredLogger(logger))
  }
  r.Use(render.SetContentType(render.ContentTypeJSON))
  r.Use(corsConfig().Handler)

  // r.Use(authenticate.AuthenticateAccessJWT)
  r.Route("/api", func(r chi.Router) {

    r.Route("/v1", func(r chi.Router) {

      // open routes
      r.Group(func(r chi.Router) {

        // we assume 600 students
        // so we reset to 600 request per minute here
        r.Use(authenticate.RateLimitMiddleware("infomark-logins",
          fmt.Sprintf("%d-M", viper.GetInt64("auth_total_requests_per_minute")),
          viper.GetString("redis_url"),
        ))

        r.Post("/auth/token", appAPI.Auth.RefreshAccessTokenHandler)
        r.Post("/auth/sessions", appAPI.Auth.LoginHandler)
        r.Post("/auth/request_password_reset", appAPI.Auth.RequestPasswordResetHandler)
        r.Post("/auth/update_password", appAPI.Auth.UpdatePasswordHandler)
        r.Post("/auth/confirm_email", appAPI.Auth.ConfirmEmailHandler)
        r.Post("/account", appAPI.Account.CreateHandler)
        r.Get("/ping", appAPI.Common.PingHandler)
      })

      // protected routes
      r.Group(func(r chi.Router) {
        r.Use(authenticate.RequiredValidAccessClaims)

        r.Get("/me", appAPI.User.GetMeHandler)
        r.Put("/me", appAPI.User.EditMeHandler)

        r.Route("/users", func(r chi.Router) {
          r.Get("/", appAPI.User.IndexHandler)
          r.Route("/{user_id}", func(r chi.Router) {
            r.Use(appAPI.User.Context)
            r.Get("/", appAPI.User.GetHandler)
            r.Get("/avatar", appAPI.User.GetAvatarHandler)
            r.Put("/", appAPI.User.EditHandler)
            r.Delete("/", appAPI.User.DeleteHandler)
            r.Post("/emails", appAPI.User.SendEmailHandler)
          })
        })

        r.Route("/courses", func(r chi.Router) {
          r.Get("/", appAPI.Course.IndexHandler)
          r.Post("/", authorize.EndpointRequiresRole(appAPI.Course.CreateHandler, authorize.ADMIN))

          r.Route("/{course_id}", func(r chi.Router) {
            r.Use(appAPI.Course.Context)
            r.Use(appAPI.Course.RoleContext)
            r.Post("/enrollments", appAPI.Course.EnrollHandler)

            r.Route("/", func(r chi.Router) {
              r.Use(authorize.RequiresAtLeastCourseRole(authorize.STUDENT))

              r.Get("/", appAPI.Course.GetHandler)
              r.Route("/", func(r chi.Router) {
                r.Use(authorize.RequiresAtLeastCourseRole(authorize.ADMIN))

                r.Post("/emails", appAPI.Course.SendEmailHandler)

                r.Put("/", appAPI.Course.EditHandler)
                r.Delete("/", appAPI.Course.DeleteHandler)
              })

              // we handle permission dependent chnages WITHIN this endpoint
              //

              r.Get("/enrollments", appAPI.Course.IndexEnrollmentsHandler)
              r.Delete("/enrollments", appAPI.Course.DisenrollHandler)

              r.Get("/points", appAPI.Course.PointsHandler)
              r.Get("/bids", appAPI.Course.BidsHandler)

              r.Route("/enrollments/{user_id}", func(r chi.Router) {
                r.Use(authorize.RequiresAtLeastCourseRole(authorize.ADMIN))

                r.Use(appAPI.User.Context)
                r.Get("/", appAPI.Course.GetUserEnrollmentHandler)
                r.Delete("/", appAPI.Course.DeleteUserEnrollmentHandler)
                r.Put("/", appAPI.Course.ChangeRole)

              })

              r.Route("/sheets", func(r chi.Router) {

                r.Get("/", appAPI.Sheet.IndexHandler)
                r.Post("/", authorize.EndpointRequiresRole(appAPI.Sheet.CreateHandler, authorize.ADMIN))

                r.Route("/{sheet_id}", func(r chi.Router) {
                  r.Use(appAPI.Sheet.Context)
                  // r.Use(appAPI.Course.RoleContext)

                  // ensures user is enrolled in the associated course

                  r.Get("/", appAPI.Sheet.GetHandler)

                  r.Get("/tasks", appAPI.Task.IndexHandler)
                  r.Get("/file", appAPI.Sheet.GetFileHandler)
                  r.Get("/points", appAPI.Sheet.PointsHandler)

                  r.Route("/", func(r chi.Router) {
                    r.Use(authorize.RequiresAtLeastCourseRole(authorize.ADMIN))

                    r.Put("/", appAPI.Sheet.EditHandler)
                    r.Delete("/", appAPI.Sheet.DeleteHandler)
                    r.Post("/tasks", appAPI.Task.CreateHandler)
                    r.Post("/file", appAPI.Sheet.ChangeFileHandler)
                  })

                }) // sheet_id
              })

              r.Route("/groups", func(r chi.Router) {
                r.Get("/own", appAPI.Group.GetMineHandler)
                r.Get("/", appAPI.Group.IndexHandler)
                r.Post("/", authorize.EndpointRequiresRole(appAPI.Group.CreateHandler, authorize.ADMIN))
                r.Route("/{group_id}", func(r chi.Router) {
                  r.Use(appAPI.Group.Context)
                  r.Use(appAPI.Course.RoleContext)

                  // ensures user is enrolled in the associated course

                  r.Post("/bids", appAPI.Group.ChangeBidHandler)
                  r.Post("/emails", authorize.EndpointRequiresRole(appAPI.Group.SendEmailHandler, authorize.TUTOR))
                  r.Get("/enrollments", appAPI.Group.IndexEnrollmentsHandler)
                  r.Post("/enrollments", authorize.EndpointRequiresRole(appAPI.Group.EditGroupEnrollmentHandler, authorize.ADMIN))

                  r.Get("/", appAPI.Group.GetHandler)

                  r.Route("/", func(r chi.Router) {
                    r.Use(authorize.RequiresAtLeastCourseRole(authorize.ADMIN))
                    r.Put("/", appAPI.Group.EditHandler)
                    r.Delete("/", appAPI.Group.DeleteHandler)
                  })
                })
              })

              r.Route("/grades", func(r chi.Router) {

                r.Get("/", authorize.EndpointRequiresRole(appAPI.Grade.IndexHandler, authorize.TUTOR))
                // does not require a role
                r.Get("/missing", appAPI.Grade.IndexMissingHandler)
                r.Route("/{grade_id}", func(r chi.Router) {
                  r.Use(appAPI.Grade.Context)
                  r.Use(appAPI.Course.RoleContext)

                  // ensures user is enrolled in the associated course
                  r.Use(authorize.RequiresAtLeastCourseRole(authorize.TUTOR))

                  r.Put("/", appAPI.Grade.EditHandler)
                  r.Get("/", appAPI.Grade.GetByIDHandler)
                  r.Post("/public_result", authorize.EndpointRequiresRole(appAPI.Grade.PublicResultEditHandler, authorize.ADMIN))
                  r.Post("/private_result", authorize.EndpointRequiresRole(appAPI.Grade.PrivateResultEditHandler, authorize.ADMIN))

                })
              })

              r.Route("/materials", func(r chi.Router) {

                r.Get("/", appAPI.Material.IndexHandler)
                r.Post("/", authorize.EndpointRequiresRole(appAPI.Material.CreateHandler, authorize.ADMIN))

                r.Route("/{material_id}", func(r chi.Router) {
                  r.Use(appAPI.Material.Context)
                  r.Use(appAPI.Course.RoleContext)

                  // ensures user is enrolled in the associated course

                  r.Get("/", appAPI.Material.GetHandler)
                  r.Get("/file", appAPI.Material.GetFileHandler)

                  r.Route("/", func(r chi.Router) {
                    r.Use(authorize.RequiresAtLeastCourseRole(authorize.ADMIN))

                    r.Put("/", appAPI.Material.EditHandler)
                    r.Delete("/", appAPI.Material.DeleteHandler)
                    r.Post("/file", appAPI.Material.ChangeFileHandler)
                  })
                })
              })

              r.Route("/submissions", func(r chi.Router) {
                r.Get("/", authorize.EndpointRequiresRole(appAPI.Submission.IndexHandler, authorize.TUTOR))

                r.Route("/{submission_id}", func(r chi.Router) {
                  r.Use(appAPI.Submission.Context)
                  r.Use(appAPI.Course.RoleContext)
                  r.Get("/file", appAPI.Submission.GetFileByIdHandler)
                })
              })

              r.Route("/tasks", func(r chi.Router) {

                r.Get("/missing", appAPI.Task.MissingIndexHandler)
                r.Route("/{task_id}", func(r chi.Router) {
                  r.Use(appAPI.Task.Context)
                  r.Use(appAPI.Course.RoleContext)

                  // ensures user is enrolled in the associated course

                  r.Get("/", appAPI.Task.GetHandler)
                  r.Get("/ratings", appAPI.TaskRating.GetHandler)
                  r.Post("/ratings", appAPI.TaskRating.ChangeHandler)

                  r.Get("/submission", appAPI.Submission.GetFileHandler)
                  r.Post("/submission", appAPI.Submission.UploadFileHandler)

                  r.Get("/result", appAPI.Task.GetSubmissionResultHandler)

                  r.Route("/", func(r chi.Router) {
                    r.Use(authorize.RequiresAtLeastCourseRole(authorize.ADMIN))

                    r.Put("/", appAPI.Task.EditHandler)
                    r.Delete("/", appAPI.Task.DeleteHandler)

                    r.Get("/public_file", appAPI.Task.GetPublicTestFileHandler)
                    r.Get("/private_file", appAPI.Task.GetPrivateTestFileHandler)

                    r.Post("/public_file", appAPI.Task.ChangePublicTestFileHandler)
                    r.Post("/private_file", appAPI.Task.ChangePrivateTestFileHandler)
                  })
                })

              }) // tasks

            })

          }) // course_id
        }) // course

        r.Get("/account", appAPI.Account.GetHandler)
        r.Get("/account/enrollments", appAPI.Account.GetEnrollmentsHandler)
        r.Get("/account/avatar", appAPI.Account.GetAvatarHandler)
        r.Post("/account/avatar", appAPI.Account.ChangeAvatarHandler)
        r.Delete("/account/avatar", appAPI.Account.DeleteAvatarHandler)
        r.Patch("/account", appAPI.Account.EditHandler)
        r.Delete("/auth/sessions", appAPI.Auth.LogoutHandler)

      })

    })
  })

  workDir, _ := os.Getwd()
  filesDir := filepath.Join(workDir, "static")
  FileServer(r, "/", http.Dir(filesDir))

  return r, nil
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
  if strings.ContainsAny(path, "{}*") {
    panic("FileServer does not permit URL parameters.")
  }

  fs := http.StripPrefix(path, http.FileServer(root))

  if path != "/" && path[len(path)-1] != '/' {
    r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
    path += "/"
  }
  path += "*"

  r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    fs.ServeHTTP(w, r)
  }))
}

func corsConfig() *cors.Cors {
  // Basic CORS
  // for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
  return cors.New(cors.Options{
    AllowedOrigins:   []string{"*"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: true,
    MaxAge:           86400, // Maximum value not ignored by any of major browsers
  })
}
