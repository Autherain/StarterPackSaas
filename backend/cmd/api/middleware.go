package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/autherain/test/internal/user"

	"github.com/pascaldekloe/jwt"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader != "" {
			headerParts := strings.Split(authorizationHeader, " ")

			if len(headerParts) == 2 && headerParts[0] == "Bearer" {
				token := headerParts[1]

				claims, err := jwt.HMACCheck([]byte(token), []byte(app.config.jwt.secretKey))
				if err != nil {
					app.errorHandler.InvalidAuthenticationToken(w, r)
					return
				}

				if !claims.Valid(time.Now()) {
					app.errorHandler.InvalidAuthenticationToken(w, r)
					return
				}

				if claims.Issuer != app.config.baseURL {
					app.errorHandler.InvalidAuthenticationToken(w, r)
					return
				}

				if !claims.AcceptAudience(app.config.baseURL) {
					app.errorHandler.InvalidAuthenticationToken(w, r)
					return
				}

				userID, err := strconv.Atoi(claims.Subject)
				if err != nil {
					app.errorHandler.ServerError(w, r, err)
					return
				}

				user, found, err := app.store.Users.ReadUser(&user.UserSelector{ID: userID})
				if err != nil {
					app.errorHandler.ServerError(w, r, err)
					return
				}

				if found {
					r = contextSetAuthenticatedUser(r, *user)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, found := contextGetAuthenticatedUser(r)

		if !found {
			app.errorHandler.AuthenticationRequired(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireBasicAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, plaintextPassword, ok := r.BasicAuth()
		if !ok {
			app.errorHandler.BasicAuthenticationRequired(w, r)
			return
		}

		if app.config.basicAuth.username != username {
			app.errorHandler.BasicAuthenticationRequired(w, r)
			return
		}

		err := bcrypt.CompareHashAndPassword([]byte(app.config.basicAuth.hashedPassword), []byte(plaintextPassword))
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			app.errorHandler.BasicAuthenticationRequired(w, r)
			return
		case err != nil:
			app.errorHandler.ServerError(w, r, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireTier(allowedTiers ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, found := contextGetAuthenticatedUser(r)
			if !found {
				app.errorHandler.AuthenticationRequired(w, r)
				return
			}

			hasTier := false
			for _, tier := range allowedTiers {
				if user.SubscriptionTier == tier {
					hasTier = true
					break
				}
			}

			if !hasTier {
				app.errorHandler.Forbidden(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) requireTierOrHigher(minTier string) func(http.Handler) http.Handler {
	tierLevels := map[string]int{
		"free":       1,
		"pro":        2,
		"enterprise": 3,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, found := contextGetAuthenticatedUser(r)
			if !found {
				app.errorHandler.AuthenticationRequired(w, r)
				return
			}

			userLevel := tierLevels[user.SubscriptionTier]
			requiredLevel := tierLevels[minTier]

			if userLevel == 0 {
				app.errorHandler.Forbidden(w, r)
				return
			}

			if userLevel < requiredLevel {
				app.errorHandler.Forbidden(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
