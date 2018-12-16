package router

import (
	// "context"
	// "fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"
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

func emptyHandler(w http.ResponseWriter, r *http.Request) {}

func apiRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(UserOnly)
	r.Get("/", emptyHandler)

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	// user management
	r.Route("/user", func(r chi.Router) {
		r.Post("/token", emptyHandler)
		r.Delete("/token", emptyHandler)
	})

	r.Route("/users", func(r chi.Router) {
		r.Route("/{userID}", func(r chi.Router) {
			r.Get("/", emptyHandler)
			r.Put("/", emptyHandler)
			r.Delete("/", emptyHandler)
		})
	})

	// course management
	r.Route("/course", func(r chi.Router) {
		r.Route("/{courseID}", func(r chi.Router) {
			// course globals
			r.Get("/", emptyHandler)

			// tasks
			r.Route("/tasks/{taskID}", func(r chi.Router) {
				r.Get("/", emptyHandler)
				r.Put("/", emptyHandler)
				r.Delete("/", emptyHandler)
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

	r.Get("/login", emptyHandler)
	r.Get("/logout", emptyHandler)

	r.Mount("/api", apiRouter())

	return r
}
