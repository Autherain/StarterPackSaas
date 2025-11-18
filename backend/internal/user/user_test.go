package user

import (
	"testing"
	"time"
)

func TestUserValidate(t *testing.T) {
	t.Run("Valid user passes validation", func(t *testing.T) {
		u := &User{
			ID:               1,
			Created:          time.Now(),
			Email:            "test@example.com",
			HashedPassword:   "$2a$12$testhashedpassword",
			SubscriptionTier: "free",
		}

		err := u.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Missing email fails validation", func(t *testing.T) {
		u := &User{
			ID:               1,
			Created:          time.Now(),
			Email:            "",
			HashedPassword:   "$2a$12$testhashedpassword",
			SubscriptionTier: "free",
		}

		err := u.Validate()
		if err == nil {
			t.Error("expected error for missing email")
		}
		if err.Error() != "email must be set" {
			t.Errorf("expected 'email must be set', got %v", err)
		}
	})

	t.Run("Missing hashed password fails validation", func(t *testing.T) {
		u := &User{
			ID:               1,
			Created:          time.Now(),
			Email:            "test@example.com",
			HashedPassword:   "",
			SubscriptionTier: "free",
		}

		err := u.Validate()
		if err == nil {
			t.Error("expected error for missing hashed password")
		}
		if err.Error() != "hashed password must be set" {
			t.Errorf("expected 'hashed password must be set', got %v", err)
		}
	})

	t.Run("Missing subscription tier fails validation", func(t *testing.T) {
		u := &User{
			ID:               1,
			Created:          time.Now(),
			Email:            "test@example.com",
			HashedPassword:   "$2a$12$testhashedpassword",
			SubscriptionTier: "",
		}

		err := u.Validate()
		if err == nil {
			t.Error("expected error for missing subscription tier")
		}
		if err.Error() != "subscription tier must be set" {
			t.Errorf("expected 'subscription tier must be set', got %v", err)
		}
	})

	t.Run("Invalid subscription tier fails validation", func(t *testing.T) {
		u := &User{
			ID:               1,
			Created:          time.Now(),
			Email:            "test@example.com",
			HashedPassword:   "$2a$12$testhashedpassword",
			SubscriptionTier: "invalid_tier",
		}

		err := u.Validate()
		if err == nil {
			t.Error("expected error for invalid subscription tier")
		}
		if err.Error() != "invalid subscription tier" {
			t.Errorf("expected 'invalid subscription tier', got %v", err)
		}
	})

	t.Run("Valid pro tier passes validation", func(t *testing.T) {
		u := &User{
			ID:               1,
			Created:          time.Now(),
			Email:            "test@example.com",
			HashedPassword:   "$2a$12$testhashedpassword",
			SubscriptionTier: "pro",
		}

		err := u.Validate()
		if err != nil {
			t.Errorf("expected no error for pro tier, got %v", err)
		}
	})

	t.Run("Valid enterprise tier passes validation", func(t *testing.T) {
		u := &User{
			ID:               1,
			Created:          time.Now(),
			Email:            "test@example.com",
			HashedPassword:   "$2a$12$testhashedpassword",
			SubscriptionTier: "enterprise",
		}

		err := u.Validate()
		if err != nil {
			t.Errorf("expected no error for enterprise tier, got %v", err)
		}
	})
}
