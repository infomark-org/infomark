package router

import (
	// "context"
	// "fmt"

	"net/http"

	"github.com/cgtuebingen/infomark-backend/router/api"
	"github.com/cgtuebingen/infomark-backend/router/auth"
	"github.com/cgtuebingen/infomark-backend/router/helper"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
)

// JWT code

func apiRouter() http.Handler {

	tokenAuth := auth.GetTokenAuth()

	r := chi.NewRouter()
	// Seek, verify and validate JWT tokens
	r.Use(jwtauth.Verifier(tokenAuth))

	// Handle valid / invalid tokens. In this example, we use
	// the provided authenticator middleware, but you can write your
	// own very easily, look at the Authenticator method in jwtauth.go
	// and tweak it, its not scary.
	r.Use(auth.AuthenticatorCtx)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/", helper.EmptyHandler)

	r.Mount("/users", api.UsersRoutes())

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

	// login (get JWT token)
	r.Route("/login", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))
		r.Post("/", api.Login)

	})

	r.Mount("/api", apiRouter())

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	return r
}
