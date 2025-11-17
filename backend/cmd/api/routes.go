package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.NotFound(app.notFound)
	mux.MethodNotAllowed(app.methodNotAllowed)

	mux.Use(app.logAccess)
	mux.Use(app.recoverPanic)
	mux.Use(app.authenticate)

	mux.Get("/status", app.status)
	mux.Post("/users", app.createUser)
	mux.Post("/authentication-tokens", app.createAuthenticationToken)

	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireAuthenticatedUser)

		mux.Get("/restricted", app.restricted)
	})

	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireBasicAuthentication)

		mux.Get("/restricted-basic-auth", app.restricted)
	})

	// Example: Pro tier or higher required
	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireAuthenticatedUser)
		mux.Use(app.requireTierOrHigher("pro"))

		mux.Get("/pro-features", app.restricted)
	})

	// Example: Enterprise tier only
	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireAuthenticatedUser)
		mux.Use(app.requireTier("enterprise"))

		mux.Get("/enterprise-features", app.restricted)
	})

	// Example: Multiple tiers allowed (pro or enterprise)
	mux.Group(func(mux chi.Router) {
		mux.Use(app.requireAuthenticatedUser)
		mux.Use(app.requireTier("pro", "enterprise"))

		mux.Get("/premium-features", app.restricted)
	})

	return mux
}
