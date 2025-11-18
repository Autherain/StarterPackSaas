package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsResponseWriter(t *testing.T) {
	t.Run("Track bytes written correctly", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.Write([]byte("test data"))

		assert.Equal(t, http.StatusOK, mw.StatusCode)
		assert.Equal(t, 9, mw.BytesCount)
	})

	t.Run("Track bytes written correctly for multiple writes", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.Write([]byte("test"))
		mw.Write([]byte(" "))
		mw.Write([]byte("data"))

		assert.Equal(t, http.StatusOK, mw.StatusCode)
		assert.Equal(t, 9, mw.BytesCount)
	})

	t.Run("Track status code correctly", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.WriteHeader(http.StatusTeapot)
		mw.Write([]byte("test data"))

		assert.Equal(t, http.StatusTeapot, mw.StatusCode)
		assert.Equal(t, 9, mw.BytesCount)
	})

	t.Run("Ignore status code changes after first write", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.WriteHeader(http.StatusCreated)
		mw.WriteHeader(http.StatusTeapot)

		assert.Equal(t, http.StatusCreated, mw.StatusCode)
	})

	t.Run("Write status code to underlying http.ResponseWriter", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.WriteHeader(http.StatusCreated)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Write headers to underlying http.ResponseWriter", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.Header().Set("Content-Type", "application/json")
		mw.Header().Set("X-Custom", "test-value")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "test-value", w.Header().Get("X-Custom"))
	})

	t.Run("Write body to underlying http.ResponseWriter", func(t *testing.T) {
		w := httptest.NewRecorder()
		mw := NewMetricsResponseWriter(w)

		mw.Write([]byte("test data"))
		assert.Equal(t, "test data", w.Body.String())
	})
}
