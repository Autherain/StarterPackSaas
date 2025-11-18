package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/autherain/test/internal/user"
	"github.com/autherain/test/internal/user/mocks"
	"github.com/gavv/httpexpect/v2"
	"github.com/pascaldekloe/jwt"
)

func newTestExpect(t *testing.T, handler http.Handler) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
		},
		Reporter: httpexpect.NewAssertReporter(t),
	})
}

func TestAuthenticate(t *testing.T) {
	t.Run("Adds valid authenticated user to request context", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)
		
		aliceUser := &user.User{
			ID:               testUsers["alice"].id,
			Email:            testUsers["alice"].email,
			HashedPassword:   testUsers["alice"].hashedPassword,
			SubscriptionTier: testUsers["alice"].subscriptionTier,
		}
		
		mockStore.EXPECT().
			ReadUser(&user.UserSelector{ID: testUsers["alice"].id}).
			Return(aliceUser, true, nil).
			Once()

		app := newTestApplication(t, mockStore)

		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		var capturedUser user.User
		var capturedFound bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedUser, capturedFound = contextGetAuthenticatedUser(r)
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, app.authenticate(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+jwt).
			Expect().
			Status(http.StatusTeapot)

		if !capturedFound {
			t.Error("expected user to be found in context")
		}
		if capturedUser.ID != testUsers["alice"].id {
			t.Errorf("expected user ID %d, got %d", testUsers["alice"].id, capturedUser.ID)
		}
		if capturedUser.Email != testUsers["alice"].email {
			t.Errorf("expected user email %s, got %s", testUsers["alice"].email, capturedUser.Email)
		}
	})

	t.Run("Does not add user when no authenticated user ID in request JWT", func(t *testing.T) {
		app := newTestApplication(t)

		var capturedFound bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, capturedFound = contextGetAuthenticatedUser(r)
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, app.authenticate(next))
		e.GET("/test").
			Expect().
			Status(http.StatusTeapot)

		if capturedFound {
			t.Error("expected user not to be found in context")
		}
	})

	t.Run("Does not add user when user ID not found in database", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)
		
		// Mock store returns user not found
		mockStore.EXPECT().
			ReadUser(&user.UserSelector{ID: 999}).
			Return(nil, false, nil).
			Once()

		app := newTestApplication(t, mockStore)

		jwt, _, err := app.newAuthenticationToken(999)
		if err != nil {
			t.Fatal(err)
		}

		var capturedFound bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, capturedFound = contextGetAuthenticatedUser(r)
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, app.authenticate(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+jwt).
			Expect().
			Status(http.StatusTeapot)

		if capturedFound {
			t.Error("expected user not to be found in context")
		}
	})

	t.Run("Returns a 401 response for malformed JWT bearer token", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, app.authenticate(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer bad_jwt").
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("Invalid authentication token")
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

		e := newTestExpect(t, app.authenticate(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+string(jwt)).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("Invalid authentication token")
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

		e := newTestExpect(t, app.authenticate(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+string(jwt)).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("Invalid authentication token")
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

		e := newTestExpect(t, app.authenticate(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+string(jwt)).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("Invalid authentication token")
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

		e := newTestExpect(t, app.authenticate(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+string(jwt)).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("Invalid authentication token")
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

		e := newTestExpect(t, app.authenticate(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+string(jwt)).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("Invalid authentication token")
	})
}

func TestRequireAuthenticatedUser(t *testing.T) {
	t.Run("Allows authenticated user to proceed", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)
		
		aliceUser := &user.User{
			ID:               testUsers["alice"].id,
			Email:            testUsers["alice"].email,
			HashedPassword:   testUsers["alice"].hashedPassword,
			SubscriptionTier: testUsers["alice"].subscriptionTier,
		}
		
		mockStore.EXPECT().
			ReadUser(&user.UserSelector{ID: testUsers["alice"].id}).
			Return(aliceUser, true, nil).
			Once()

		app := newTestApplication(t, mockStore)

		jwt, _, err := app.newAuthenticationToken(testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, app.authenticate(app.requireAuthenticatedUser(next)))
		e.GET("/restricted").
			WithHeader("Authorization", "Bearer "+jwt).
			Expect().
			Status(http.StatusTeapot)
	})

	t.Run("Sends unauthenticated user a 401 response", func(t *testing.T) {
		app := newTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, app.authenticate(app.requireAuthenticatedUser(next)))
		e.GET("/test").
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("You must be authenticated to access this resource")
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

		e := newTestExpect(t, app.requireBasicAuthentication(next))
		e.GET("/test").
			WithBasicAuth(authUsername, authPassword).
			Expect().
			Status(http.StatusTeapot)
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

				e := newTestExpect(t, app.requireBasicAuthentication(next))
				req := e.GET("/test")

				if tt.setAuth {
					req = req.WithBasicAuth(tt.authUsername, tt.authPassword)
				}

				resp := req.Expect().
					Status(http.StatusUnauthorized)
				
				resp.Header("WWW-Authenticate").IsEqual(`Basic realm="restricted", charset="UTF-8"`)
				resp.JSON().Object().
					Value("Error").IsEqual("You must be authenticated to access this resource")
			})
		}
	})
}
