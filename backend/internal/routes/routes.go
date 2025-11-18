package routes

import (
	"net/http"

	"github.com/autherain/test/internal/app"
	"github.com/autherain/test/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func Setup(app *app.Application) http.Handler {
	mux := chi.NewRouter()

	mux.NotFound(app.ErrorHandler.NotFound)
	mux.MethodNotAllowed(app.ErrorHandler.MethodNotAllowed)

	mux.Use(chimw.RealIP)
	mux.Use(chimw.Logger)
	mux.Use(chimw.Recoverer)
	mux.Use(middleware.Authenticate(app))
	mux.Use(chimw.Heartbeat("/health"))

	mux.Post("/users", app.UserServer.HandleCreateUser)
	mux.Post("/authentication-tokens", func(w http.ResponseWriter, r *http.Request) {
		app.UserServer.HandleCreateAuthenticationToken(w, r, app.NewAuthenticationToken)
	})

	mux.Group(func(mux chi.Router) {
		mux.Use(middleware.RequireAuthenticatedUser(app))

	})

	mux.Group(func(mux chi.Router) {
		mux.Use(middleware.RequireBasicAuthentication(app))

	})

	return mux
}
