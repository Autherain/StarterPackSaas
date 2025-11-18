package server

import (
	"net/http"
	"time"

	"github.com/autherain/test/internal/password"
	"github.com/autherain/test/internal/request"
	"github.com/autherain/test/internal/response"
	"github.com/autherain/test/internal/user"
	"github.com/autherain/test/internal/user/params"
	"github.com/autherain/test/internal/validator"
)

// Server handles HTTP requests for user operations.
type Server struct {
	users            user.UsersReadWriter
	errorHandler     func(w http.ResponseWriter, r *http.Request, err error)
	validationFailed func(w http.ResponseWriter, r *http.Request, v validator.Validator)
}

// New creates a new user server.
func New(users user.UsersReadWriter, errorHandler func(w http.ResponseWriter, r *http.Request, err error), validationFailed func(w http.ResponseWriter, r *http.Request, v validator.Validator)) *Server {
	return &Server{
		users:            users,
		errorHandler:     errorHandler,
		validationFailed: validationFailed,
	}
}

// HandleCreateUser handles POST /users requests.
func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input params.CreateUserParams

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	var v validator.Validator

	_, found, err := s.users.ReadUser(&user.UserSelector{Email: input.Email})
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	v.CheckField(input.Email != "", "Email", "Email is required")
	v.CheckField(validator.Matches(input.Email, validator.RgxEmail), "Email", "Must be a valid email address")
	v.CheckField(!found, "Email", "Email is already in use")

	v.CheckField(input.Password != "", "Password", "Password is required")
	v.CheckField(len(input.Password) >= 8, "Password", "Password is too short")
	v.CheckField(len(input.Password) <= 72, "Password", "Password is too long")
	v.CheckField(validator.NotIn(input.Password, password.CommonPasswords...), "Password", "Password is too common")

	if v.HasErrors() {
		s.validationFailed(w, r, v)
		return
	}

	hashedPassword, err := password.Hash(input.Password)
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	newUser := input.Map(hashedPassword)

	err = s.users.CreateUser(newUser)
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AuthenticationTokenResponse represents the response for authentication token creation.
type AuthenticationTokenResponse struct {
	AuthenticationToken       string
	AuthenticationTokenExpiry string
	User                      *user.User
}

// HandleCreateAuthenticationToken handles POST /authentication-tokens requests.
func (s *Server) HandleCreateAuthenticationToken(w http.ResponseWriter, r *http.Request, createToken func(userID int) (string, time.Time, error)) {
	var input params.CreateAuthenticationTokenParams

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	var v validator.Validator

	u, found, err := s.users.ReadUser(&user.UserSelector{Email: input.Email})
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	v.CheckField(input.Email != "", "Email", "Email is required")
	v.CheckField(found, "Email", "Email address could not be found")

	if found {
		passwordMatches, err := password.Matches(input.Password, u.HashedPassword)
		if err != nil {
			s.errorHandler(w, r, err)
			return
		}

		v.CheckField(input.Password != "", "Password", "Password is required")
		v.CheckField(passwordMatches, "Password", "Password is incorrect")
	}

	if v.HasErrors() {
		s.validationFailed(w, r, v)
		return
	}

	jwt, expiry, err := createToken(u.ID)
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	data := map[string]string{
		"AuthenticationToken":       jwt,
		"AuthenticationTokenExpiry": expiry.UTC().Format(time.RFC3339),
	}

	err = response.JSON(w, http.StatusOK, data)
	if err != nil {
		s.errorHandler(w, r, err)
	}
}
