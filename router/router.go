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

package router

import (
	"net/http"

	"github.com/cgtuebingen/infomark-backend/router/api"
	"github.com/cgtuebingen/infomark-backend/router/auth"
	"github.com/cgtuebingen/infomark-backend/router/helper"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
)

func apiV1Router() http.Handler {

	tokenAuth := auth.GetTokenAuth()

	r := chi.NewRouter()
	// Seek, verify and validate JWT tokens
	r.Use(jwtauth.Verifier(tokenAuth))

	// Handle valid / invalid tokens.
	r.Use(auth.AuthenticatorCtx)

	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/", helper.EmptyHandler)

	r.Mount("/login", api.LoginRoutes())
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

	r.Route("/api", func(r chi.Router) {
		// Whoever's reading this: You'll thank me for that.
		r.Mount("/v1", apiV1Router())

		// Health status
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("pong"))
		})
	})

	return r
}
