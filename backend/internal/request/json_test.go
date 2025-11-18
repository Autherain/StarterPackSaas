package request

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testDecodeJSONTarget struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func TestDecodeJSON(t *testing.T) {
	t.Run("Decode valid JSON successfully", func(t *testing.T) {
		jsonBody := `{"name":"John","age":30,"email":"john@example.com"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		w := httptest.NewRecorder()

		var target testDecodeJSONTarget
		err := DecodeJSON(w, req, &target)
		assert.Nil(t, err)
		assert.Equal(t, "John", target.Name)
		assert.Equal(t, 30, target.Age)
		assert.Equal(t, "john@example.com", target.Email)
	})

	t.Run("Allow unknown fields", func(t *testing.T) {
		jsonBody := `{"name":"John","age":30,"email":"john@example.com","unknown_field":"value"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		w := httptest.NewRecorder()

		var target testDecodeJSONTarget
		err := DecodeJSON(w, req, &target)
		assert.Nil(t, err)
		assert.Equal(t, "John", target.Name)
		assert.Equal(t, 30, target.Age)
		assert.Equal(t, "john@example.com", target.Email)
	})

	t.Run("Return error for empty body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(""))
		w := httptest.NewRecorder()

		var target testDecodeJSONTarget
		err := DecodeJSON(w, req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, "body must not be empty", err.Error())
	})

	t.Run("Return error for JSON that isn't a struct", func(t *testing.T) {
		jsonBody := `"not-a-struct"`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		w := httptest.NewRecorder()

		var target testDecodeJSONTarget
		err := DecodeJSON(w, req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, "body contains incorrect JSON type (at character 14)", err.Error())
	})

	t.Run("Return error for malformed JSON", func(t *testing.T) {
		jsonBody := `{"name":"John","age":"30",}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		w := httptest.NewRecorder()

		var target testDecodeJSONTarget
		err := DecodeJSON(w, req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, "body contains badly-formed JSON (at character 27)", err.Error())
	})

	t.Run("Return error for unexpected EOF", func(t *testing.T) {
		jsonBody := `{"name"`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		w := httptest.NewRecorder()

		var target testDecodeJSONTarget
		err := DecodeJSON(w, req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, "body contains badly-formed JSON", err.Error())
	})

	t.Run("Return error for incorrect JSON type", func(t *testing.T) {
		jsonBody := `{"name":"John","age":"not-a-number","email":"john@example.com"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		w := httptest.NewRecorder()

		var target testDecodeJSONTarget
		err := DecodeJSON(w, req, &target)

		assert.NotNil(t, err)
		assert.Equal(t, `body contains incorrect JSON type for field "age"`, err.Error())
	})

	t.Run("Return error for body larger than limit", func(t *testing.T) {

		largeValue := strings.Repeat("a", 1_048_577)
		jsonBody := `{"name":"` + largeValue + `"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		w := httptest.NewRecorder()

		var target testDecodeJSONTarget
		err := DecodeJSON(w, req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, "body must not be larger than 1048576 bytes", err.Error())
	})

	t.Run("Return error for multiple JSON values", func(t *testing.T) {
		jsonBody := `{"name":"John","age":30}{"name":"Jane","age":25}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		w := httptest.NewRecorder()

		var target testDecodeJSONTarget
		err := DecodeJSON(w, req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, "body must only contain a single JSON value", err.Error())
	})
}

func TestDecodeJSONStrict(t *testing.T) {
	t.Run("Return error for unknown fields", func(t *testing.T) {
		jsonBody := `{"name":"John","age":30,"email":"john@example.com","unknown_field":"value"}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		w := httptest.NewRecorder()

		var target testDecodeJSONTarget
		err := DecodeJSONStrict(w, req, &target)
		assert.NotNil(t, err)
		assert.Equal(t, `body contains unknown key "unknown_field"`, err.Error())
	})
}
