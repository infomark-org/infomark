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

package app

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/infomark-org/infomark/auth"
	"github.com/infomark-org/infomark/auth/authenticate"
	"github.com/infomark-org/infomark/auth/authorize"
	"github.com/infomark-org/infomark/configuration"
	"github.com/infomark-org/infomark/symbol"
	"github.com/jmoiron/sqlx"
	"github.com/markbates/pkger"
	"github.com/sirupsen/logrus"
)

// LimitedDecoder limits the amount of data a client can send in a JSON data request.
// The golang fork-join multi-threading allows no easy way to cancel started requests.
// Therefore we limit the amount of data which is read by the server whenever
// we need to parse a JSON request.
func LimitedDecoder(r *http.Request, v interface{}) error {
	var err error

	switch render.GetRequestContentType(r) {
	case render.ContentTypeJSON:
		body := io.LimitReader(r.Body, int64(configuration.Configuration.Server.HTTP.Limits.MaxRequestJSON))
		err = render.DecodeJSON(body, v)
	default:
		err = errors.New("render: unable to automatically decode the request content type")
	}
	//
	return err
}

var log *logrus.Logger

func RunInit() {
	log = logrus.New()
	render.Decode = LimitedDecoder

	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.Out = os.Stdout
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		end := time.Now()
		if r.RequestURI != "/metrics" {
			log.WithFields(logrus.Fields{
				"method": r.Method,
				// "proto":   r.Proto,
				"agent":   r.UserAgent(),
				"remote":  r.RemoteAddr,
				"latency": end.Sub(start),
				"time":    end.Format(time.RFC3339),
			}).Info(r.RequestURI)
		}
	})
}

func BasicAuthMiddleware(realm string, credentials map[string]string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
			username, password, ok := r.BasicAuth()
			if !ok {
				w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
				render.Render(w, r, auth.ErrUnauthenticated)
				return
			}
			validPassword, userFound := credentials[username]
			if !userFound {
				w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
				render.Render(w, r, auth.ErrUnauthenticated)
				return
			}

			if subtle.ConstantTimeCompare([]byte(password), []byte(validPassword)) == 1 {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
			render.Render(w, r, auth.ErrUnauthenticated)
		})
	}
}

// VersionMiddleware writes the current API version to the headers.
func VersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-INFOMARK-VERSION", symbol.Version.String())
		w.Header().Set("X-INFOMARK-Commit", symbol.GitCommit)
		next.ServeHTTP(w, r)
	})
}

// SecureMiddleware writes required access headers to all requests.
func SecureMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		if r.UserAgent() == "" {
			render.Render(w, r, ErrBadRequestWithDetails(
				fmt.Errorf(`Request forbidden by administrative rules.
Please make sure your request has a User-Agent header`)))
			return
		}

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
func New(db *sqlx.DB, promhttp http.Handler, log bool) (*chi.Mux, error) {
	logger := logrus.StandardLogger()

	config := &configuration.Configuration.Server

	InitPrometheus()

	tokenAuth := authenticate.NewTokenAuth(&config.Authentication)
	sessionAuth := authenticate.NewSessionAuth(&config.Authentication)

	if err := db.Ping(); err != nil {
		logger.WithField("module", "database").Error(err)
		return nil, err
	}

	appAPI, err := NewAPI(db, tokenAuth, sessionAuth)
	if err != nil {
		logger.WithField("module", "app").Error(err)
		return nil, err
	}

	loginLimiter, err := authenticate.NewLoginLimiter("infomark-logins",
		fmt.Sprintf("%d-M", config.Authentication.TotalRequestsPerMinute),
		config.RedisURL())
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
	if log {
		r.Use(LoggingMiddleware)
	}
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(corsConfig().Handler)

	basicAuth := BasicAuthMiddleware("Restricted", map[string]string{
		configuration.Configuration.Server.Services.Prometheus.User: configuration.Configuration.Server.Services.Prometheus.Password,
	})

	r.Group(func(r chi.Router) {
		r.Use(basicAuth)
		r.Handle("/metrics", promhttp)
	})

	// r.Use(authenticate.AuthenticateAccessJWT)
	r.Route("/api", func(r chi.Router) {

		r.Route("/v1", func(r chi.Router) {

			// open routes
			r.Group(func(r chi.Router) {

				// we assume 600 students
				// so we reset to 600 request per minute here
				r.Use(authenticate.RateLimitMiddleware(loginLimiter))

				r.Post("/auth/token", appAPI.Auth.RefreshAccessTokenHandler)
				r.Post("/auth/sessions", appAPI.Auth.LoginHandler)
				r.Post("/auth/request_password_reset", appAPI.Auth.RequestPasswordResetHandler)
				r.Post("/auth/update_password", appAPI.Auth.UpdatePasswordHandler)
				r.Post("/auth/confirm_email", appAPI.Auth.ConfirmEmailHandler)
				r.Post("/account", appAPI.Account.CreateHandler)
				r.Get("/ping", appAPI.Common.PingHandler)
				r.Get("/version", appAPI.Common.VersionHandler)
				r.Get("/privacy_statement", appAPI.Common.PrivacyStatementHandler)
			})

			// protected routes
			r.Group(func(r chi.Router) {
				r.Use(authenticate.RequiredValidAccessClaims(sessionAuth, config))

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
					r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Get("/find", appAPI.User.Find)
				})

				r.Route("/courses", func(r chi.Router) {
					r.Get("/", appAPI.Course.IndexHandler)
					r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Post("/", appAPI.Course.CreateHandler)

					r.Route("/{course_id}", func(r chi.Router) {
						r.Use(appAPI.Course.Context)
						r.Use(appAPI.Course.RoleContext)


						r.Post("/enrollments", appAPI.Course.EnrollHandler)

						r.Route("/", func(r chi.Router) {
							r.Use(authorize.RequiresAtLeastCourseRole(authorize.STUDENT))

							r.Get("/", appAPI.Course.GetHandler)

							r.Route("/", func(r chi.Router) {
								r.Use(authorize.RequiresAtLeastCourseRole(authorize.TUTOR))

								r.Get("/teamcount", appAPI.Course.TeamCountHandler)

								r.Route("/", func(r chi.Router) {
									r.Use(authorize.RequiresAtLeastCourseRole(authorize.ADMIN))

									r.Post("/emails", appAPI.Course.SendEmailHandler)
									r.Put("/", appAPI.Course.EditHandler)
									r.Delete("/", appAPI.Course.DeleteHandler)
								})
							})

							r.Get("/team", appAPI.Team.IndexTeamHandler)
							r.Get("/teams", appAPI.Team.IncompleteTeamsHandler)

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
								r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Post("/", appAPI.Sheet.CreateHandler)

								r.Route("/{sheet_id}", func(r chi.Router) {
									r.Use(appAPI.Sheet.Context)

									r.Get("/", appAPI.Sheet.GetHandler)
									r.Route("/tasks", func(r chi.Router) {
										r.Get("/", appAPI.Task.IndexHandler)
										r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Post("/", appAPI.Task.CreateHandler)
									})

									r.Get("/file", appAPI.Sheet.GetFileHandler)
									r.Get("/points", appAPI.Sheet.PointsHandler)

									r.Route("/", func(r chi.Router) {
										r.Use(authorize.RequiresAtLeastCourseRole(authorize.ADMIN))

										r.Put("/", appAPI.Sheet.EditHandler)
										r.Delete("/", appAPI.Sheet.DeleteHandler)
										r.Post("/file", appAPI.Sheet.ChangeFileHandler)
									})
								})
							})

							r.Route("/groups", func(r chi.Router) {
								r.Get("/own", appAPI.Group.GetMineHandler)
								r.Get("/", appAPI.Group.IndexHandler)
								r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Post("/", appAPI.Group.CreateHandler)

								r.Route("/{group_id}", func(r chi.Router) {
									r.Use(appAPI.Group.Context)

									r.Post("/bids", appAPI.Group.ChangeBidHandler)
									r.With(authorize.RequiresAtLeastCourseRole(authorize.TUTOR)).Post("/emails", appAPI.Group.SendEmailHandler)
									r.Get("/enrollments", appAPI.Group.IndexEnrollmentsHandler)
									r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Post("/enrollments", appAPI.Group.EditGroupEnrollmentHandler)
									r.Get("/", appAPI.Group.GetHandler)

									r.Route("/", func(r chi.Router) {
										r.Use(authorize.RequiresAtLeastCourseRole(authorize.ADMIN))

										r.Put("/", appAPI.Group.EditHandler)
										r.Delete("/", appAPI.Group.DeleteHandler)
									})
								})
							})

							r.Route("/exams", func(r chi.Router) {
								r.Get("/", appAPI.Exam.IndexHandler)
								r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Post("/", appAPI.Exam.CreateHandler)

								r.Route("/{exam_id}", func(r chi.Router) {
									r.Use(appAPI.Exam.Context)

									r.Get("/", appAPI.Exam.GetHandler)
									r.Route("/", func(r chi.Router) {
										r.Use(authorize.RequiresAtLeastCourseRole(authorize.ADMIN))

										r.Put("/", appAPI.Exam.EditHandler)
										r.Delete("/", appAPI.Exam.DeleteHandler)
									})

									r.Route("/enrollments", func(r chi.Router) {
										r.Post("/", appAPI.Exam.EnrollExamHandler)
										r.Delete("/", appAPI.Exam.DisenrollExamHandler)
										r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Put("/", appAPI.Exam.UpdateEnrollExamHandler)
										r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Get("/", appAPI.Exam.GetExamEnrollmentsHandler)
									})

								})
							})

							r.Route("/grades", func(r chi.Router) {
								r.With(authorize.RequiresAtLeastCourseRole(authorize.TUTOR)).Get("/", appAPI.Grade.IndexHandler)
								r.With(authorize.RequiresAtLeastCourseRole(authorize.TUTOR)).Get("/summary", appAPI.Grade.IndexSummaryHandler)
								r.Get("/missing", appAPI.Grade.IndexMissingHandler)

								r.Route("/{grade_id}", func(r chi.Router) {
									r.Use(appAPI.Grade.Context)
									r.Use(authorize.RequiresAtLeastCourseRole(authorize.TUTOR))

									r.Put("/", appAPI.Grade.EditHandler)
									r.Get("/", appAPI.Grade.GetByIDHandler)
									r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Post("/public_result", appAPI.Grade.PublicResultEditHandler)
									r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Post("/private_result", appAPI.Grade.PrivateResultEditHandler)
								})
							})

							r.Route("/materials", func(r chi.Router) {
								r.Get("/", appAPI.Material.IndexHandler)
								r.With(authorize.RequiresAtLeastCourseRole(authorize.ADMIN)).Post("/", appAPI.Material.CreateHandler)

								r.Route("/{material_id}", func(r chi.Router) {
									r.Use(appAPI.Material.Context)

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
								r.With(authorize.RequiresAtLeastCourseRole(authorize.TUTOR)).Get("/", appAPI.Submission.IndexHandler)

								r.Route("/{submission_id}", func(r chi.Router) {
									r.Use(appAPI.Submission.Context)

									r.Get("/file", appAPI.Submission.GetFileByIDHandler)
								})
							})

							r.Route("/tasks", func(r chi.Router) {
								r.Get("/missing", appAPI.Task.MissingIndexHandler)

								r.Route("/{task_id}", func(r chi.Router) {
									r.Use(appAPI.Task.Context)

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

									r.Route("/groups/{group_id}", func(r chi.Router) {
										r.Use(authorize.RequiresAtLeastCourseRole(authorize.TUTOR))
										r.Use(appAPI.Group.Context)

										r.Get("/file", appAPI.Submission.GetCollectionFileHandler)
										r.Get("/", appAPI.Submission.GetCollectionHandler)
									})
								})
							})
						}) // at least required authorize.STUDENT
					})
				})

				r.Get("/account", appAPI.Account.GetHandler)
				r.Get("/account/enrollments", appAPI.Account.GetEnrollmentsHandler)
				r.Get("/account/exams/enrollments", appAPI.Account.GetExamEnrollmentsHandler)
				r.Get("/account/avatar", appAPI.Account.GetAvatarHandler)
				r.Post("/account/avatar", appAPI.Account.ChangeAvatarHandler)
				r.Delete("/account/avatar", appAPI.Account.DeleteAvatarHandler)
				r.Patch("/account", appAPI.Account.EditHandler)
				r.Delete("/auth/sessions", appAPI.Auth.LogoutHandler)

			})

		})
	})

	FileServer(r, "/", pkger.Dir("/static"))

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
