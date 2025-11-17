package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/autherain/test/internal/assert"
	"github.com/autherain/test/internal/database"

	"github.com/pascaldekloe/jwt"
)

func TestRecoverPanic(t *testing.T) {
	t.Run("Allows normal requests to proceed", func(t *testing.T) {
		app := newTestApplication(t)
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, app.recoverPanic(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})

	t.Run("Recovers from panic and sends a 500 response", func(t *testing.T) {
		var buf bytes.Buffer
		app := newTestApplication(t)
		app.logger = slog.New(slog.NewTextHandler(&buf, nil))

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("something went wrong")
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, app.recoverPanic(next))
		assert.Equal(t, res.StatusCode, http.StatusInternalServerError)
		assert.Equal(t, res.BodyFields["Error"], "The server encountered a problem and could not process your request")
	})
}

func TestLogAccess(t *testing.T) {
	t.Run("Logs the request and response details", func(t *testing.T) {
		var buf bytes.Buffer
		app := newTestApplication(t)
		app.logger = slog.New(slog.NewTextHandler(&buf, nil))

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
			w.Write([]byte(`{"Message": "I'm a test teapot"}`))
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, app.logAccess(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
		assert.True(t, strings.Contains(buf.String(), "level=INFO"))
		assert.True(t, strings.Contains(buf.String(), "msg=access"))
		assert.True(t, strings.Contains(buf.String(), "request.method=GET"))
		assert.True(t, strings.Contains(buf.String(), "request.url=/test"))
		assert.True(t, strings.Contains(buf.String(), "response.status=418"))
		assert.True(t, strings.Contains(buf.String(), "response.size=32"))
	})
}

func TestAuthenticate(t *testing.T) {
	t.Run("Adds valid authenticated user to request context", func(t *testing.T) {
		app := newTestApplication(t)

		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		var capturedUser database.User
		var capturedFound bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedUser, capturedFound = contextGetAuthenticatedUser(r)
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
		assert.True(t, capturedFound)
		assert.Equal(t, capturedUser.ID, testUsers["alice"].id)
		assert.Equal(t, capturedUser.Email, testUsers["alice"].email)
	})

	t.Run("Does not add user when no authenticated user ID in request JWT", func(t *testing.T) {
		app := newTestApplication(t)

		var capturedFound bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, capturedFound = contextGetAuthenticatedUser(r)
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, app.authenticate(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
		assert.False(t, capturedFound)
	})

	t.Run("Does not add user when user ID not found in database", func(t *testing.T) {
		app := newTestApplication(t)

		jwt, _, err := app.newAuthenticationToken(999)
		if err != nil {
			t.Fatal(err)
		}

		var capturedFound bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, capturedFound = contextGetAuthenticatedUser(r)
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
		assert.False(t, capturedFound)
	})

	t.Run("Returns a 401 response for malformed JWT bearer token", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer bad_jwt")

		res := send(t, req, app.authenticate(next))
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
		assert.Equal(t, res.BodyFields["Error"], "Invalid authentication token")
	})

	t.Run("Returns a 401 response for JWT created with invalid secret key", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		claims := newTestClaims()

		jwt, err := claims.HMACSign(jwt.HS256, []byte("this-is-the-wrong-key"))
		if err != nil {
			t.Fatal(err)
		}

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+string(jwt))

		res := send(t, req, app.authenticate(next))
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
		assert.Equal(t, res.BodyFields["Error"], "Invalid authentication token")
	})

	t.Run("Returns a 401 response for JWT created with invalid issuer", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		claims := newTestClaims()
		claims.Issuer = "https://wrong.example.com"

		jwt, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secretKey))
		if err != nil {
			t.Fatal(err)
		}

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+string(jwt))

		res := send(t, req, app.authenticate(next))
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
		assert.Equal(t, res.BodyFields["Error"], "Invalid authentication token")
	})

	t.Run("Returns a 401 response for JWT created with invalid audience", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		claims := newTestClaims()
		claims.Audiences = []string{"https://wrong.example.com"}

		jwt, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secretKey))
		if err != nil {
			t.Fatal(err)
		}

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+string(jwt))

		res := send(t, req, app.authenticate(next))
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
		assert.Equal(t, res.BodyFields["Error"], "Invalid authentication token")
	})

	t.Run("Returns a 401 response for expired JWT", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		claims := newTestClaims()
		claims.Issued = jwt.NewNumericTime(time.Now().Add(-1 * time.Hour))
		claims.NotBefore = jwt.NewNumericTime(time.Now().Add(-1 * time.Hour))
		claims.Expires = jwt.NewNumericTime(time.Now().Add(-1 * time.Second))

		jwt, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secretKey))
		if err != nil {
			t.Fatal(err)
		}

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+string(jwt))

		res := send(t, req, app.authenticate(next))
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
		assert.Equal(t, res.BodyFields["Error"], "Invalid authentication token")
	})

	t.Run("Returns a 401 response for not-yet issued JWT", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		claims := newTestClaims()
		claims.Issued = jwt.NewNumericTime(time.Now().Add(time.Second))
		claims.NotBefore = jwt.NewNumericTime(time.Now().Add(time.Second))
		claims.Expires = jwt.NewNumericTime(time.Now().Add(time.Hour))

		jwt, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secretKey))
		if err != nil {
			t.Fatal(err)
		}

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+string(jwt))

		res := send(t, req, app.authenticate(next))
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
		assert.Equal(t, res.BodyFields["Error"], "Invalid authentication token")
	})
}

func TestRequireAuthenticatedUser(t *testing.T) {
	t.Run("Allows authenticated user to proceed", func(t *testing.T) {
		app := newTestApplication(t)

		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/restricted", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(app.requireAuthenticatedUser(next)))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})

	t.Run("Sends unauthenticated user a 401 response", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, app.authenticate(app.requireAuthenticatedUser(next)))
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
		assert.Equal(t, res.BodyFields["Error"], "You must be authenticated to access this resource")
	})
}

func TestRequireBasicAuthentication(t *testing.T) {
	t.Run("Allows user with valid basic auth credentials to proceed", func(t *testing.T) {
		app := newTestApplication(t)
		authUsername := "admin"
		authPassword := "placeholder*77"
		validHashedPassword := "$2a$04$HLvpR86.wXVT.2KHHkUbFe4/ou3wYGnc9FD7VcKaixofed5enOS.W"

		app.config.basicAuth.username = authUsername
		app.config.basicAuth.hashedPassword = validHashedPassword

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.SetBasicAuth(authUsername, authPassword)

		res := send(t, req, app.requireBasicAuthentication(next))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})

	t.Run("Sends a 401 response including WWW-Authenticate header for invalid authentication", func(t *testing.T) {
		validUsername := "admin"
		validPassword := "placeholder*77"
		validHashedPassword := "$2a$04$HLvpR86.wXVT.2KHHkUbFe4/ou3wYGnc9FD7VcKaixofed5enOS.W"

		tests := []struct {
			name         string
			setAuth      bool
			authUsername string
			authPassword string
		}{
			{
				name:    "No basic auth credentials provided",
				setAuth: false,
			},
			{
				name:         "Invalid username provided",
				setAuth:      true,
				authUsername: "wronguser",
				authPassword: validPassword,
			},
			{
				name:         "Invalid password provided",
				setAuth:      true,
				authUsername: validUsername,
				authPassword: "wrongpassword",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				app := newTestApplication(t)

				app.config.basicAuth.username = validUsername
				app.config.basicAuth.hashedPassword = validHashedPassword

				next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				})

				req := newTestRequest(t, http.MethodGet, "/test", nil)
				if tt.setAuth {
					req.SetBasicAuth(tt.authUsername, tt.authPassword)
				}

				res := send(t, req, app.requireBasicAuthentication(next))
				assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
				assert.Equal(t, res.Header.Get("WWW-Authenticate"), `Basic realm="restricted", charset="UTF-8"`)
				assert.Equal(t, res.BodyFields["Error"], "You must be authenticated to access this resource")
			})
		}
	})
}

func TestRequireTier(t *testing.T) {
	t.Run("Allows user with matching tier to proceed", func(t *testing.T) {
		app := newTestApplication(t)

		// Update alice to have pro tier
		err := app.db.UpdateUserSubscriptionTier(testUsers["alice"].id, "pro")
		if err != nil {
			t.Fatal(err)
		}

		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(app.requireTier("pro")(next)))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})

	t.Run("Allows user with one of multiple allowed tiers to proceed", func(t *testing.T) {
		app := newTestApplication(t)

		// Update alice to have enterprise tier
		err := app.db.UpdateUserSubscriptionTier(testUsers["alice"].id, "enterprise")
		if err != nil {
			t.Fatal(err)
		}

		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(app.requireTier("pro", "enterprise")(next)))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})

	t.Run("Sends 403 response for user without matching tier", func(t *testing.T) {
		app := newTestApplication(t)

		// alice has free tier by default
		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(app.requireTier("pro")(next)))
		assert.Equal(t, res.StatusCode, http.StatusForbidden)
		assert.Equal(t, res.BodyFields["Error"], "You don't have permission to access this resource")
	})

	t.Run("Sends 401 response for unauthenticated user", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, app.authenticate(app.requireTier("pro")(next)))
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
		assert.Equal(t, res.BodyFields["Error"], "You must be authenticated to access this resource")
	})
}

func TestRequireTierOrHigher(t *testing.T) {
	t.Run("Allows user with exact required tier to proceed", func(t *testing.T) {
		app := newTestApplication(t)

		// Update alice to have pro tier
		err := app.db.UpdateUserSubscriptionTier(testUsers["alice"].id, "pro")
		if err != nil {
			t.Fatal(err)
		}

		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(app.requireTierOrHigher("pro")(next)))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})

	t.Run("Allows user with higher tier to proceed", func(t *testing.T) {
		app := newTestApplication(t)

		// Update alice to have enterprise tier
		err := app.db.UpdateUserSubscriptionTier(testUsers["alice"].id, "enterprise")
		if err != nil {
			t.Fatal(err)
		}

		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(app.requireTierOrHigher("pro")(next)))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})

	t.Run("Sends 403 response for user with lower tier", func(t *testing.T) {
		app := newTestApplication(t)

		// alice has free tier by default
		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(app.requireTierOrHigher("pro")(next)))
		assert.Equal(t, res.StatusCode, http.StatusForbidden)
		assert.Equal(t, res.BodyFields["Error"], "You don't have permission to access this resource")
	})

	t.Run("Sends 403 response for user with invalid tier", func(t *testing.T) {
		app := newTestApplication(t)

		// Update alice to have invalid tier
		err := app.db.UpdateUserSubscriptionTier(testUsers["alice"].id, "invalid_tier")
		if err != nil {
			t.Fatal(err)
		}

		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(app.requireTierOrHigher("pro")(next)))
		assert.Equal(t, res.StatusCode, http.StatusForbidden)
		assert.Equal(t, res.BodyFields["Error"], "You don't have permission to access this resource")
	})

	t.Run("Sends 401 response for unauthenticated user", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)

		res := send(t, req, app.authenticate(app.requireTierOrHigher("pro")(next)))
		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
		assert.Equal(t, res.BodyFields["Error"], "You must be authenticated to access this resource")
	})

	t.Run("Free tier can access free tier resources", func(t *testing.T) {
		app := newTestApplication(t)

		// alice has free tier by default
		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(app.requireTierOrHigher("free")(next)))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})

	t.Run("Pro tier can access free tier resources", func(t *testing.T) {
		app := newTestApplication(t)

		// Update alice to have pro tier
		err := app.db.UpdateUserSubscriptionTier(testUsers["alice"].id, "pro")
		if err != nil {
			t.Fatal(err)
		}

		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		req := newTestRequest(t, http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		res := send(t, req, app.authenticate(app.requireTierOrHigher("free")(next)))
		assert.Equal(t, res.StatusCode, http.StatusTeapot)
	})
}
