package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.NotFound(app.errorHandler.NotFound)
	mux.MethodNotAllowed(app.errorHandler.MethodNotAllowed)

	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(app.authenticate)
	mux.Use(middleware.Heartbeat("/health"))

	mux.Post("/users", app.userServer.HandleCreateUser)
	mux.Post("/authentication-tokens", app.createAuthenticationToken)

	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireAuthenticatedUser)

	})

	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireBasicAuthentication)

	})

	return mux
}

func (app *application) createAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	app.userServer.HandleCreateAuthenticationToken(w, r, app.newAuthenticationToken)
}
