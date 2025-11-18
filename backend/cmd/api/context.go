package main

import (
	"context"
	"net/http"

	"github.com/autherain/test/internal/user"
)

type contextKey string

const (
	authenticatedUserContextKey = contextKey("authenticatedUser")
)

func contextSetAuthenticatedUser(r *http.Request, u user.User) *http.Request {
	ctx := context.WithValue(r.Context(), authenticatedUserContextKey, u)
	return r.WithContext(ctx)
}

func contextGetAuthenticatedUser(r *http.Request) (user.User, bool) {
	u, ok := r.Context().Value(authenticatedUserContextKey).(user.User)
	return u, ok
}
