package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/autherain/test/internal/user"
	"github.com/autherain/test/internal/user/mocks"
	"github.com/autherain/test/internal/user/server"
	"github.com/autherain/test/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test helpers for error and validation handlers
var testErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
}

var testValidationHandler = func(w http.ResponseWriter, r *http.Request, v validator.Validator) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(map[string]any{"Errors": v.Errors})
}

func createJSONRequest(t *testing.T, data map[string]any) *http.Request {
	t.Helper()

	body, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestHandleCreateUser(t *testing.T) {
	t.Run("Successfully creates a new user", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		// User doesn't exist yet
		mockStore.EXPECT().
			ReadUserByEmail("newuser@example.com").
			Return(nil, false, nil).
			Once()

		// CreateUser should be called with valid data
		mockStore.EXPECT().
			CreateUser(mock.MatchedBy(func(u *user.User) bool {
				return u.Email == "newuser@example.com" &&
					u.SubscriptionTier == "free" &&
					u.HashedPassword != "" &&
					u.HashedPassword != "validPassword123!"
			})).
			Return(nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "newuser@example.com",
			"Password": "validPassword123!",
		})

		srv.HandleCreateUser(w, r)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Rejects empty email", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		// Empty email still triggers DB check
		mockStore.EXPECT().
			ReadUserByEmail("").
			Return(nil, false, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "",
			"Password": "validPassword123!",
		})

		srv.HandleCreateUser(w, r)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Rejects invalid email format", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		mockStore.EXPECT().
			ReadUserByEmail("invalid-email").
			Return(nil, false, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "invalid-email",
			"Password": "validPassword123!",
		})

		srv.HandleCreateUser(w, r)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Rejects duplicate email", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		existingUser := &user.User{
			ID:    1,
			Email: "existing@example.com",
		}

		// User already exists
		mockStore.EXPECT().
			ReadUserByEmail("existing@example.com").
			Return(existingUser, true, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "existing@example.com",
			"Password": "validPassword123!",
		})

		srv.HandleCreateUser(w, r)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		// Just verify it returned a validation error, don't inspect the body structure
		assert.True(t, w.Body.Len() > 0)
	})

	t.Run("Rejects password that is too short", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		mockStore.EXPECT().
			ReadUserByEmail("test@example.com").
			Return(nil, false, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "test@example.com",
			"Password": "short",
		})

		srv.HandleCreateUser(w, r)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Rejects password that is too long", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		longPassword := "iRbMr5Av5T1DINST1l2pGBBUtW4Qn628N4lN6tFNjW8Ea4fuYiI84j2KH8tKQrF3INkqbKwZh"

		mockStore.EXPECT().
			ReadUserByEmail("test@example.com").
			Return(nil, false, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "test@example.com",
			"Password": longPassword,
		})

		srv.HandleCreateUser(w, r)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Rejects common password", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		mockStore.EXPECT().
			ReadUserByEmail("test@example.com").
			Return(nil, false, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "test@example.com",
			"Password": "password", // Very common password
		})

		srv.HandleCreateUser(w, r)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestHandleCreateAuthenticationToken(t *testing.T) {
	// Mock token generation function
	mockTokenGen := func(userID int) (string, time.Time, error) {
		expiry := time.Now().Add(24 * time.Hour)
		return "mock_jwt_token_123", expiry, nil
	}

	t.Run("Successfully creates authentication token for valid credentials", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		// Password for "testPass123!" hashed with bcrypt cost 4
		testUser := &user.User{
			ID:               123,
			Email:            "alice@example.com",
			HashedPassword:   "$2a$04$27fHaQw5jwiMKYoxhLek4uyj9zp29lxtmLWGuC0MR6tuispXJn9US",
			SubscriptionTier: "free",
		}

		mockStore.EXPECT().
			ReadUserByEmail("alice@example.com").
			Return(testUser, true, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "alice@example.com",
			"Password": "testPass123!",
		})

		srv.HandleCreateAuthenticationToken(w, r, mockTokenGen)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		json.NewDecoder(w.Body).Decode(&response)

		assert.Equal(t, "mock_jwt_token_123", response["AuthenticationToken"])
		assert.Regexp(t, `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, response["AuthenticationTokenExpiry"])
	})

	t.Run("Rejects non-existent email", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		// User not found
		mockStore.EXPECT().
			ReadUserByEmail("nonexistent@example.com").
			Return(nil, false, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "nonexistent@example.com",
			"Password": "somePassword123!",
		})

		srv.HandleCreateAuthenticationToken(w, r, mockTokenGen)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Rejects empty email", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		mockStore.EXPECT().
			ReadUserByEmail("").
			Return(nil, false, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "",
			"Password": "testPass123!",
		})

		srv.HandleCreateAuthenticationToken(w, r, mockTokenGen)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Rejects empty password", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		testUser := &user.User{
			ID:               123,
			Email:            "alice@example.com",
			HashedPassword:   "$2a$04$27fHaQw5jwiMKYoxhLek4uyj9zp29lxtmLWGuC0MR6tuispXJn9US",
			SubscriptionTier: "free",
		}

		mockStore.EXPECT().
			ReadUserByEmail("alice@example.com").
			Return(testUser, true, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "alice@example.com",
			"Password": "",
		})

		srv.HandleCreateAuthenticationToken(w, r, mockTokenGen)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Rejects incorrect password", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		testUser := &user.User{
			ID:               123,
			Email:            "alice@example.com",
			HashedPassword:   "$2a$04$27fHaQw5jwiMKYoxhLek4uyj9zp29lxtmLWGuC0MR6tuispXJn9US",
			SubscriptionTier: "free",
		}

		mockStore.EXPECT().
			ReadUserByEmail("alice@example.com").
			Return(testUser, true, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "alice@example.com",
			"Password": "wrongPassword123!",
		})

		srv.HandleCreateAuthenticationToken(w, r, mockTokenGen)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})

	t.Run("Validates password before checking if it matches", func(t *testing.T) {
		mockStore := mocks.NewMockUsersReadWriter(t)

		// User doesn't exist
		mockStore.EXPECT().
			ReadUserByEmail("test@example.com").
			Return(nil, false, nil).
			Once()

		srv := server.New(mockStore, testErrorHandler, testValidationHandler)

		w := httptest.NewRecorder()
		r := createJSONRequest(t, map[string]any{
			"Email":    "test@example.com",
			"Password": "", // Empty password
		})

		srv.HandleCreateAuthenticationToken(w, r, mockTokenGen)

		// Should fail validation before attempting password check
		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}
