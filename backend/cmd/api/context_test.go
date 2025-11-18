package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/autherain/test/internal/user"
	"github.com/stretchr/testify/assert"
)

func TestContextSetAuthenticatedUser(t *testing.T) {
	testUser := user.User{
		ID:             123,
		Created:        time.Now(),
		Email:          "alice@example.com",
		HashedPassword: "$2a$12$testhashedpassword",
	}

	t.Run("Returns new request with user set and original unchanged", func(t *testing.T) {
		originalReq, err := http.NewRequest(http.MethodGet, "/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		modifiedReq := contextSetAuthenticatedUser(originalReq, testUser)

		retrievedUser, found := originalReq.Context().Value(authenticatedUserContextKey).(user.User)
		assert.False(t, found)
		assert.Equal(t, user.User{}, retrievedUser)

		retrievedUser, found = modifiedReq.Context().Value(authenticatedUserContextKey).(user.User)
		assert.True(t, found)
		assert.Equal(t, testUser, retrievedUser)
	})
}

func TestContextGetAuthenticatedUser(t *testing.T) {
	testUser := user.User{
		ID:             123,
		Created:        time.Now(),
		Email:          "alice@example.com",
		HashedPassword: "$2a$12$testhashedpassword",
	}

	t.Run("Successfully returns user when set", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.WithValue(req.Context(), authenticatedUserContextKey, testUser)
		req = req.WithContext(ctx)

		retrievedUser, found := contextGetAuthenticatedUser(req)
		assert.True(t, found)
		assert.Equal(t, testUser, retrievedUser)
	})

	t.Run("Returns zero user and false when not set", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		if err != nil {
			t.Fatal(err)
		}

		retrievedUser, found := contextGetAuthenticatedUser(req)
		assert.False(t, found)
		assert.Equal(t, user.User{}, retrievedUser)
	})

	t.Run("Returns zero user and false if wrong type", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.WithValue(req.Context(), authenticatedUserContextKey, 123)
		req = req.WithContext(ctx)

		retrievedUser, found := contextGetAuthenticatedUser(req)
		assert.False(t, found)
		assert.Equal(t, user.User{}, retrievedUser)
	})
}
