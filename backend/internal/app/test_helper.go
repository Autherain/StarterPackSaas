package app

import (
	"io"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/autherain/test/internal/errors"
	"github.com/autherain/test/internal/store"
	"github.com/autherain/test/internal/user"
	"github.com/pascaldekloe/jwt"
)

type TestUser struct {
	ID               int
	Email            string
	Password         string
	HashedPassword   string
	SubscriptionTier string
}

var TestUsers = map[string]*TestUser{
	"alice": {ID: 1, Email: "alice@example.com", Password: "testPass123!", HashedPassword: "$2a$04$27fHaQw5jwiMKYoxhLek4uyj9zp29lxtmLWGuC0MR6tuispXJn9US", SubscriptionTier: "free"},
	"bob":   {ID: 2, Email: "bob@example.com", Password: "mySecure456#", HashedPassword: "$2a$04$O6QOPBSFw14SyLBXs64MJuQd8o7GaBKYvbDqeHGgZRi6FN87aXDWC", SubscriptionTier: "free"},
}

func NewTestClaims() jwt.Claims {
	var c jwt.Claims
	c.Subject = strconv.Itoa(TestUsers["alice"].ID)
	c.Issued = jwt.NewNumericTime(time.Now())
	c.NotBefore = jwt.NewNumericTime(time.Now())
	c.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	c.Issuer = "https://www.example.com"
	c.Audiences = []string{"https://www.example.com"}
	return c
}

func NewTestApplication(t *testing.T, usersStore ...user.UsersReadWriter) *Application {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	app := &Application{
		Logger:       logger,
		ErrorHandler: errors.New(logger),
	}

	// If a mock store is provided, use it
	if len(usersStore) > 0 {
		app.Store = &store.Store{
			Users: usersStore[0],
		}
	}

	app.Config.JWT.SecretKey = "k7mp29rf4qxhwn8vbtaj6pgucmve53y9"
	app.Config.BaseURL = "https://www.example.com"

	return app
}
