package server

import (
	"net/http"
	"time"

	"github.com/autherain/test/internal/password"
	"github.com/autherain/test/internal/request"
	"github.com/autherain/test/internal/response"
	"github.com/autherain/test/internal/user"
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
	var input struct {
		Email     string              `json:"Email"`
		Password  string              `json:"Password"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	_, found, err := s.users.ReadUserByEmail(input.Email)
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	input.Validator.CheckField(input.Email != "", "Email", "Email is required")
	input.Validator.CheckField(validator.Matches(input.Email, validator.RgxEmail), "Email", "Must be a valid email address")
	input.Validator.CheckField(!found, "Email", "Email is already in use")

	input.Validator.CheckField(input.Password != "", "Password", "Password is required")
	input.Validator.CheckField(len(input.Password) >= 8, "Password", "Password is too short")
	input.Validator.CheckField(len(input.Password) <= 72, "Password", "Password is too long")
	input.Validator.CheckField(validator.NotIn(input.Password, password.CommonPasswords...), "Password", "Password is too common")

	if input.Validator.HasErrors() {
		s.validationFailed(w, r, input.Validator)
		return
	}

	hashedPassword, err := password.Hash(input.Password)
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	newUser := &user.User{
		Email:            input.Email,
		HashedPassword:   hashedPassword,
		SubscriptionTier: "free",
	}

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
	var input struct {
		Email     string              `json:"Email"`
		Password  string              `json:"Password"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	u, found, err := s.users.ReadUserByEmail(input.Email)
	if err != nil {
		s.errorHandler(w, r, err)
		return
	}

	input.Validator.CheckField(input.Email != "", "Email", "Email is required")
	input.Validator.CheckField(found, "Email", "Email address could not be found")

	if found {
		passwordMatches, err := password.Matches(input.Password, u.HashedPassword)
		if err != nil {
			s.errorHandler(w, r, err)
			return
		}

		input.Validator.CheckField(input.Password != "", "Password", "Password is required")
		input.Validator.CheckField(passwordMatches, "Password", "Password is incorrect")
	}

	if input.Validator.HasErrors() {
		s.validationFailed(w, r, input.Validator)
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
