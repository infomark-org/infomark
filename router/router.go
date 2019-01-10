package router

import (
	// "context"
	// "fmt"
	"net/http"

	"github.com/cgtuebingen/infomark-backend/router/api"
	"github.com/cgtuebingen/infomark-backend/router/helper"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

func UserOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func apiRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(UserOnly)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/", helper.EmptyHandler)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// user management
	r.Route("/user", func(r chi.Router) {
		r.Post("/token", helper.EmptyHandler)
		r.Delete("/token", helper.EmptyHandler)
	})

	r.Mount("/users", api.UserRoutes())

	// course management
	r.Route("/course", func(r chi.Router) {
		r.Route("/{courseID}", func(r chi.Router) {
			// course globals
			r.Get("/", helper.EmptyHandler)

			// tasks
			r.Route("/tasks/{taskID}", func(r chi.Router) {
				r.Get("/", helper.EmptyHandler)
				r.Put("/", helper.EmptyHandler)
				r.Delete("/", helper.EmptyHandler)
			})
		})
	})

	return r
}

func GetRouter() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Get("/login", helper.EmptyHandler)
	r.Get("/logout", helper.EmptyHandler)

	r.Mount("/api", apiRouter())

	return r
}
