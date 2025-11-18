package context

import (
	"context"
	"net/http"

	"github.com/autherain/test/internal/user"
)

type Key string

const (
	AuthenticatedUserKey = Key("authenticatedUser")
)

func SetAuthenticatedUser(r *http.Request, u user.User) *http.Request {
	ctx := context.WithValue(r.Context(), AuthenticatedUserKey, u)
	return r.WithContext(ctx)
}

func GetAuthenticatedUser(r *http.Request) (user.User, bool) {
	u, ok := r.Context().Value(AuthenticatedUserKey).(user.User)
	return u, ok
}
