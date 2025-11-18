package main

import (
	"strconv"
	"testing"
	"time"

	"github.com/autherain/test/internal/assert"
	"github.com/pascaldekloe/jwt"
)

func TestNewAuthenticationToken(t *testing.T) {
	app := newTestApplication(t)

	t.Run("generates valid JWT token and expiry time", func(t *testing.T) {
		userID := 123

		token, expiry, err := app.newAuthenticationToken(userID)
		assert.Nil(t, err)
		assert.MatchesRegexp(t, token, `^eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`)
		assert.True(t, expiry.After(time.Now()))
	})

	t.Run("token contains correct claims", func(t *testing.T) {
		userID := 456
		token, _, err := app.newAuthenticationToken(userID)
		assert.Nil(t, err)

		claims, err := jwt.HMACCheck([]byte(token), []byte(app.config.jwt.secretKey))
		assert.Nil(t, err)
		assert.Equal(t, claims.Subject, strconv.Itoa(userID))
		assert.Equal(t, claims.Issuer, app.config.baseURL)
		assert.Equal(t, len(claims.Audiences), 1)
		assert.Equal(t, claims.Audiences[0], app.config.baseURL)

		assert.True(t, time.Since(claims.Issued.Time()) < time.Second)
		assert.True(t, time.Since(claims.NotBefore.Time()) < time.Second)

		duration := time.Until(claims.Expires.Time())
		assert.False(t, duration < 23*time.Hour+59*time.Minute || duration > 24*time.Hour+1*time.Minute)
	})

	t.Run("generates different tokens on subsequent calls", func(t *testing.T) {
		userID := 100
		token1, _, err := app.newAuthenticationToken(userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		token2, _, err := app.newAuthenticationToken(userID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if token1 == token2 {
			t.Error("expected different tokens for subsequent calls (different issued times)")
		}
	})
}
