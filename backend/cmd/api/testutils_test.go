package main

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

type testUser struct {
	id               int
	email            string
	password         string
	hashedPassword   string
	subscriptionTier string
}

var testUsers = map[string]*testUser{
	"alice": {id: 1, email: "alice@example.com", password: "testPass123!", hashedPassword: "$2a$04$27fHaQw5jwiMKYoxhLek4uyj9zp29lxtmLWGuC0MR6tuispXJn9US", subscriptionTier: "free"},
	"bob":   {id: 2, email: "bob@example.com", password: "mySecure456#", hashedPassword: "$2a$04$O6QOPBSFw14SyLBXs64MJuQd8o7GaBKYvbDqeHGgZRi6FN87aXDWC", subscriptionTier: "free"},
}

func newTestClaims() jwt.Claims {
	var c jwt.Claims
	c.Subject = strconv.Itoa(testUsers["alice"].id)
	c.Issued = jwt.NewNumericTime(time.Now())
	c.NotBefore = jwt.NewNumericTime(time.Now())
	c.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	c.Issuer = "https://www.example.com"
	c.Audiences = []string{"https://www.example.com"}
	return c
}

func newTestApplication(t *testing.T, usersStore ...user.UsersReadWriter) *application {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	
	app := &application{
		logger:       logger,
		errorHandler: errors.New(logger),
	}

	// If a mock store is provided, use it
	if len(usersStore) > 0 {
		app.store = &store.Store{
			Users: usersStore[0],
		}
	}

	app.config.jwt.secretKey = "k7mp29rf4qxhwn8vbtaj6pgucmve53y9"
	app.config.baseURL = "https://www.example.com"

	return app
}
