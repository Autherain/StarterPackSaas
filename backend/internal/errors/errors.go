package errors

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/autherain/test/internal/response"
	"github.com/autherain/test/internal/validator"
)

// Handler provides methods for handling HTTP errors with logging.
type Handler struct {
	logger *slog.Logger
}

// New creates a new error Handler with the provided logger.
func New(logger *slog.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// ReportServerError logs server errors with request details and stack trace.
func (h *Handler) ReportServerError(r *http.Request, err error) {
	var (
		message = err.Error()
		method  = r.Method
		url     = r.URL.String()
		trace   = string(debug.Stack())
	)

	requestAttrs := slog.Group("request", "method", method, "url", url)
	h.logger.Error(message, requestAttrs, "trace", trace)
}

// ErrorMessage writes a JSON error response with the given status code and message.
func (h *Handler) ErrorMessage(w http.ResponseWriter, r *http.Request, status int, message string, headers http.Header) {
	message = strings.ToUpper(message[:1]) + message[1:]

	err := response.JSONWithHeaders(w, status, map[string]string{"Error": message}, headers)
	if err != nil {
		h.ReportServerError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ServerError logs the error and sends a 500 Internal Server Error response.
func (h *Handler) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	h.ReportServerError(r, err)

	message := "The server encountered a problem and could not process your request"
	h.ErrorMessage(w, r, http.StatusInternalServerError, message, nil)
}

// NotFound sends a 404 Not Found response.
func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	h.ErrorMessage(w, r, http.StatusNotFound, message, nil)
}

// MethodNotAllowed sends a 405 Method Not Allowed response.
func (h *Handler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	h.ErrorMessage(w, r, http.StatusMethodNotAllowed, message, nil)
}

// BadRequest sends a 400 Bad Request response with the error message.
func (h *Handler) BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	h.ErrorMessage(w, r, http.StatusBadRequest, err.Error(), nil)
}

// FailedValidation sends a 422 Unprocessable Entity response with validation errors.
func (h *Handler) FailedValidation(w http.ResponseWriter, r *http.Request, v validator.Validator) {
	err := response.JSON(w, http.StatusUnprocessableEntity, v)
	if err != nil {
		h.ServerError(w, r, err)
	}
}

// InvalidAuthenticationToken sends a 401 Unauthorized response for invalid tokens.
func (h *Handler) InvalidAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", "Bearer")

	h.ErrorMessage(w, r, http.StatusUnauthorized, "Invalid authentication token", headers)
}

// AuthenticationRequired sends a 401 Unauthorized response when authentication is required.
func (h *Handler) AuthenticationRequired(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", "Bearer")

	h.ErrorMessage(w, r, http.StatusUnauthorized, "You must be authenticated to access this resource", headers)
}

// BasicAuthenticationRequired sends a 401 Unauthorized response for basic auth.
func (h *Handler) BasicAuthenticationRequired(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	message := "You must be authenticated to access this resource"
	h.ErrorMessage(w, r, http.StatusUnauthorized, message, headers)
}

// Forbidden sends a 403 Forbidden response.
func (h *Handler) Forbidden(w http.ResponseWriter, r *http.Request) {
	message := "You don't have permission to access this resource"
	h.ErrorMessage(w, r, http.StatusForbidden, message, nil)
}
