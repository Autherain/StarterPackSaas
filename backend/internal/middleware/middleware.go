package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/autherain/test/internal/app"
	"github.com/autherain/test/internal/context"
	"github.com/autherain/test/internal/user"

	"github.com/pascaldekloe/jwt"
	"golang.org/x/crypto/bcrypt"
)

func Authenticate(app *app.Application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Authorization")

			authorizationHeader := r.Header.Get("Authorization")

			if authorizationHeader != "" {
				headerParts := strings.Split(authorizationHeader, " ")

				if len(headerParts) == 2 && headerParts[0] == "Bearer" {
					token := headerParts[1]

					claims, err := jwt.HMACCheck([]byte(token), []byte(app.Config.JWT.SecretKey))
					if err != nil {
						app.ErrorHandler.InvalidAuthenticationToken(w, r)
						return
					}

					if !claims.Valid(time.Now()) {
						app.ErrorHandler.InvalidAuthenticationToken(w, r)
						return
					}

					if claims.Issuer != app.Config.BaseURL {
						app.ErrorHandler.InvalidAuthenticationToken(w, r)
						return
					}

					if !claims.AcceptAudience(app.Config.BaseURL) {
						app.ErrorHandler.InvalidAuthenticationToken(w, r)
						return
					}

					userID, err := strconv.Atoi(claims.Subject)
					if err != nil {
						app.ErrorHandler.ServerError(w, r, err)
						return
					}

					u, found, err := app.Store.Users.ReadUser(&user.UserSelector{ID: userID})
					if err != nil {
						app.ErrorHandler.ServerError(w, r, err)
						return
					}

					if found {
						r = context.SetAuthenticatedUser(r, *u)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuthenticatedUser(app *app.Application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, found := context.GetAuthenticatedUser(r)

			if !found {
				app.ErrorHandler.AuthenticationRequired(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireBasicAuthentication(app *app.Application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, plaintextPassword, ok := r.BasicAuth()
			if !ok {
				app.ErrorHandler.BasicAuthenticationRequired(w, r)
				return
			}

			if app.Config.BasicAuth.Username != username {
				app.ErrorHandler.BasicAuthenticationRequired(w, r)
				return
			}

			err := bcrypt.CompareHashAndPassword([]byte(app.Config.BasicAuth.HashedPassword), []byte(plaintextPassword))
			switch {
			case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
				app.ErrorHandler.BasicAuthenticationRequired(w, r)
				return
			case err != nil:
				app.ErrorHandler.ServerError(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireTier(app *app.Application, allowedTiers ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, found := context.GetAuthenticatedUser(r)
			if !found {
				app.ErrorHandler.AuthenticationRequired(w, r)
				return
			}

			hasTier := false
			for _, tier := range allowedTiers {
				if u.SubscriptionTier == tier {
					hasTier = true
					break
				}
			}

			if !hasTier {
				app.ErrorHandler.Forbidden(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireTierOrHigher(app *app.Application, minTier string) func(http.Handler) http.Handler {
	tierLevels := map[string]int{
		"free":       1,
		"pro":        2,
		"enterprise": 3,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, found := context.GetAuthenticatedUser(r)
			if !found {
				app.ErrorHandler.AuthenticationRequired(w, r)
				return
			}

			userLevel := tierLevels[u.SubscriptionTier]
			requiredLevel := tierLevels[minTier]

			if userLevel == 0 {
				app.ErrorHandler.Forbidden(w, r)
				return
			}

			if userLevel < requiredLevel {
				app.ErrorHandler.Forbidden(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
