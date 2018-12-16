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
		// ctx := r.Context()
		// perm, ok := ctx.Value("acl.permission").(YourPermissionType)
		// if !ok || !perm.IsAdmin() {
		// 	http.Error(w, http.StatusText(403), 403)
		// 	return
		// }
		next.ServeHTTP(w, r)
	})
}

func apiRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(UserOnly)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
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

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {})
	r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {})

	r.Mount("/api", apiRouter())

	return r
}
