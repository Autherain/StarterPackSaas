package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/autherain/test/internal/app"
	"github.com/autherain/test/internal/context"
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
			ID:               app.TestUsers["alice"].ID,
			Email:            app.TestUsers["alice"].Email,
			HashedPassword:   app.TestUsers["alice"].HashedPassword,
			SubscriptionTier: app.TestUsers["alice"].SubscriptionTier,
		}

		mockStore.EXPECT().
			ReadUser(&user.UserSelector{ID: app.TestUsers["alice"].ID}).
			Return(aliceUser, true, nil).
			Once()

		appInstance := app.NewTestApplication(t, mockStore)

		jwt, _, err := appInstance.NewAuthenticationToken(app.TestUsers["alice"].ID)
		if err != nil {
			t.Fatal(err)
		}

		var capturedUser user.User
		var capturedFound bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedUser, capturedFound = context.GetAuthenticatedUser(r)
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, Authenticate(appInstance)(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+jwt).
			Expect().
			Status(http.StatusTeapot)

		if !capturedFound {
			t.Error("expected user to be found in context")
		}
		if capturedUser.ID != app.TestUsers["alice"].ID {
			t.Errorf("expected user ID %d, got %d", app.TestUsers["alice"].ID, capturedUser.ID)
		}
		if capturedUser.Email != app.TestUsers["alice"].Email {
			t.Errorf("expected user email %s, got %s", app.TestUsers["alice"].Email, capturedUser.Email)
		}
	})

	t.Run("Does not add user when no authenticated user ID in request JWT", func(t *testing.T) {
		appInstance := app.NewTestApplication(t)

		var capturedFound bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, capturedFound = context.GetAuthenticatedUser(r)
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, Authenticate(appInstance)(next))
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

		appInstance := app.NewTestApplication(t, mockStore)

		jwt, _, err := appInstance.NewAuthenticationToken(999)
		if err != nil {
			t.Fatal(err)
		}

		var capturedFound bool
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, capturedFound = context.GetAuthenticatedUser(r)
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, Authenticate(appInstance)(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+jwt).
			Expect().
			Status(http.StatusTeapot)

		if capturedFound {
			t.Error("expected user not to be found in context")
		}
	})

	t.Run("Returns a 401 response for malformed JWT bearer token", func(t *testing.T) {
		appInstance := app.NewTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, Authenticate(appInstance)(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer bad_jwt").
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("Invalid authentication token")
	})

	t.Run("Returns a 401 response for JWT created with invalid secret key", func(t *testing.T) {
		appInstance := app.NewTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		claims := app.NewTestClaims()

		jwtToken, err := claims.HMACSign(jwt.HS256, []byte("this-is-the-wrong-key"))
		if err != nil {
			t.Fatal(err)
		}

		e := newTestExpect(t, Authenticate(appInstance)(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+string(jwtToken)).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("Invalid authentication token")
	})

	t.Run("Returns a 401 response for JWT created with invalid issuer", func(t *testing.T) {
		appInstance := app.NewTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		claims := app.NewTestClaims()
		claims.Issuer = "https://wrong.example.com"

		jwtToken, err := claims.HMACSign(jwt.HS256, []byte(appInstance.Config.JWT.SecretKey))
		if err != nil {
			t.Fatal(err)
		}

		e := newTestExpect(t, Authenticate(appInstance)(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+string(jwtToken)).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("Invalid authentication token")
	})

	t.Run("Returns a 401 response for JWT created with invalid audience", func(t *testing.T) {
		appInstance := app.NewTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		claims := app.NewTestClaims()
		claims.Audiences = []string{"https://wrong.example.com"}

		jwtToken, err := claims.HMACSign(jwt.HS256, []byte(appInstance.Config.JWT.SecretKey))
		if err != nil {
			t.Fatal(err)
		}

		e := newTestExpect(t, Authenticate(appInstance)(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+string(jwtToken)).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("Invalid authentication token")
	})

	t.Run("Returns a 401 response for expired JWT", func(t *testing.T) {
		appInstance := app.NewTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		claims := app.NewTestClaims()
		claims.Issued = jwt.NewNumericTime(time.Now().Add(-1 * time.Hour))
		claims.NotBefore = jwt.NewNumericTime(time.Now().Add(-1 * time.Hour))
		claims.Expires = jwt.NewNumericTime(time.Now().Add(-1 * time.Second))

		jwtToken, err := claims.HMACSign(jwt.HS256, []byte(appInstance.Config.JWT.SecretKey))
		if err != nil {
			t.Fatal(err)
		}

		e := newTestExpect(t, Authenticate(appInstance)(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+string(jwtToken)).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("Invalid authentication token")
	})

	t.Run("Returns a 401 response for not-yet issued JWT", func(t *testing.T) {
		appInstance := app.NewTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		claims := app.NewTestClaims()
		claims.Issued = jwt.NewNumericTime(time.Now().Add(time.Second))
		claims.NotBefore = jwt.NewNumericTime(time.Now().Add(time.Second))
		claims.Expires = jwt.NewNumericTime(time.Now().Add(time.Hour))

		jwtToken, err := claims.HMACSign(jwt.HS256, []byte(appInstance.Config.JWT.SecretKey))
		if err != nil {
			t.Fatal(err)
		}

		e := newTestExpect(t, Authenticate(appInstance)(next))
		e.GET("/test").
			WithHeader("Authorization", "Bearer "+string(jwtToken)).
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
			ID:               app.TestUsers["alice"].ID,
			Email:            app.TestUsers["alice"].Email,
			HashedPassword:   app.TestUsers["alice"].HashedPassword,
			SubscriptionTier: app.TestUsers["alice"].SubscriptionTier,
		}

		mockStore.EXPECT().
			ReadUser(&user.UserSelector{ID: app.TestUsers["alice"].ID}).
			Return(aliceUser, true, nil).
			Once()

		appInstance := app.NewTestApplication(t, mockStore)

		jwt, _, err := appInstance.NewAuthenticationToken(app.TestUsers["alice"].ID)
		if err != nil {
			t.Fatal(err)
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, Authenticate(appInstance)(RequireAuthenticatedUser(appInstance)(next)))
		e.GET("/restricted").
			WithHeader("Authorization", "Bearer "+jwt).
			Expect().
			Status(http.StatusTeapot)
	})

	t.Run("Sends unauthenticated user a 401 response", func(t *testing.T) {
		appInstance := app.NewTestApplication(t)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, Authenticate(appInstance)(RequireAuthenticatedUser(appInstance)(next)))
		e.GET("/test").
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("Error").IsEqual("You must be authenticated to access this resource")
	})
}

func TestRequireBasicAuthentication(t *testing.T) {
	t.Run("Allows user with valid basic auth credentials to proceed", func(t *testing.T) {
		appInstance := app.NewTestApplication(t)
		authUsername := "admin"
		authPassword := "placeholder*77"
		validHashedPassword := "$2a$04$HLvpR86.wXVT.2KHHkUbFe4/ou3wYGnc9FD7VcKaixofed5enOS.W"

		appInstance.Config.BasicAuth.Username = authUsername
		appInstance.Config.BasicAuth.HashedPassword = validHashedPassword

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		e := newTestExpect(t, RequireBasicAuthentication(appInstance)(next))
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
				appInstance := app.NewTestApplication(t)

				appInstance.Config.BasicAuth.Username = validUsername
				appInstance.Config.BasicAuth.HashedPassword = validHashedPassword

				next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusTeapot)
				})

				e := newTestExpect(t, RequireBasicAuthentication(appInstance)(next))
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
