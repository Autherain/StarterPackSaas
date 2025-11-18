package errors

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/autherain/test/internal/validator"
	"github.com/stretchr/testify/assert"
)

func TestReportServerError(t *testing.T) {
	t.Run("Logs error with correct details", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		handler := New(logger)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		handler.ReportServerError(req, errors.New("this is a test error"))
		assert.True(t, strings.Contains(buf.String(), "level=ERROR"))
		assert.True(t, strings.Contains(buf.String(), `msg="this is a test error"`))
		assert.True(t, strings.Contains(buf.String(), "request.method=GET"))
		assert.True(t, strings.Contains(buf.String(), "request.url=/test"))
	})
}

func TestServerError(t *testing.T) {
	t.Run("Logs error and sends a 500 response without exposing error details", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		handler := New(logger)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServerError(rec, req, errors.New("this is a test error"))

		assert.Equal(t, rec.Code, http.StatusInternalServerError)
		assert.Equal(t, rec.Header().Get("Content-Type"), "application/json")
		assert.True(t, strings.Contains(rec.Body.String(), "The server encountered a problem and could not process your request"))

		assert.True(t, strings.Contains(buf.String(), "level=ERROR"))
		assert.True(t, strings.Contains(buf.String(), `msg="this is a test error"`))
		assert.True(t, strings.Contains(buf.String(), "request.method=GET"))
		assert.True(t, strings.Contains(buf.String(), "request.url=/test"))
	})
}

func TestNotFound(t *testing.T) {
	t.Run("Sends a 404 response and error message", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		handler := New(logger)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.NotFound(rec, req)

		assert.Equal(t, rec.Code, http.StatusNotFound)
		assert.True(t, strings.Contains(rec.Body.String(), "The requested resource could not be found"))
	})
}

func TestMethodNotAllowed(t *testing.T) {
	t.Run("Sends a 405 response and error message", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		handler := New(logger)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.MethodNotAllowed(rec, req)

		assert.Equal(t, rec.Code, http.StatusMethodNotAllowed)
		assert.True(t, strings.Contains(rec.Body.String(), "The GET method is not supported for this resource"))
	})
}

func TestBadRequest(t *testing.T) {
	t.Run("Sends a 400 response including the error message", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		handler := New(logger)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.BadRequest(rec, req, errors.New("this is a baaaad request"))

		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.True(t, strings.Contains(rec.Body.String(), "This is a baaaad request"))
	})
}

func TestFailedValidation(t *testing.T) {
	t.Run("Sends a 422 response including the validation failures", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		handler := New(logger)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		var v validator.Validator
		v.AddError("This is an validation failure message")
		handler.FailedValidation(rec, req, v)

		assert.Equal(t, rec.Code, http.StatusUnprocessableEntity)
		assert.True(t, strings.Contains(rec.Body.String(), "This is an validation failure message"))
	})
}

func TestInvalidAuthenticationToken(t *testing.T) {
	t.Run("Sends a 401 response including a WWW-Authenticate header", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		handler := New(logger)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.InvalidAuthenticationToken(rec, req)

		assert.Equal(t, rec.Code, http.StatusUnauthorized)
		assert.Equal(t, rec.Header().Get("WWW-Authenticate"), "Bearer")
		assert.True(t, strings.Contains(rec.Body.String(), "Invalid authentication token"))
	})
}

func TestAuthenticationRequired(t *testing.T) {
	t.Run("Sends a 401 response including a WWW-Authenticate header", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		handler := New(logger)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.AuthenticationRequired(rec, req)

		assert.Equal(t, rec.Code, http.StatusUnauthorized)
		assert.Equal(t, rec.Header().Get("WWW-Authenticate"), "Bearer")
		assert.True(t, strings.Contains(rec.Body.String(), "You must be authenticated to access this resource"))
	})
}

func TestBasicAuthenticationRequired(t *testing.T) {
	t.Run("Sends a 401 response including a WWW-Authenticate header", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		handler := New(logger)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.BasicAuthenticationRequired(rec, req)

		assert.Equal(t, rec.Code, http.StatusUnauthorized)
		assert.Equal(t, rec.Header().Get("WWW-Authenticate"), `Basic realm="restricted", charset="UTF-8"`)
		assert.True(t, strings.Contains(rec.Body.String(), "You must be authenticated to access this resource"))
	})
}

func TestForbidden(t *testing.T) {
	t.Run("Sends a 403 response and error message", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
		handler := New(logger)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.Forbidden(rec, req)

		assert.Equal(t, rec.Code, http.StatusForbidden)
		assert.True(t, strings.Contains(rec.Body.String(), "You don't have permission to access this resource"))
	})
}
