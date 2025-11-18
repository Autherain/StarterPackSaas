package user

import (
	"errors"
	"time"
)

// User represents a user in the system.
type User struct {
	ID               int
	Created          time.Time
	Email            string
	HashedPassword   string
	SubscriptionTier string
}

// Validate validates the user data.
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email must be set")
	}

	if u.HashedPassword == "" {
		return errors.New("hashed password must be set")
	}

	if u.SubscriptionTier == "" {
		return errors.New("subscription tier must be set")
	}

	// Validate subscription tier is one of the allowed values
	validTiers := map[string]bool{
		"free":       true,
		"pro":        true,
		"enterprise": true,
	}
	if !validTiers[u.SubscriptionTier] {
		return errors.New("invalid subscription tier")
	}

	return nil
}

// UserSelector is used to select a single user.
type UserSelector struct {
	ID    int
	Email string
}

// UsersReader defines methods for reading users.
type UsersReader interface {
	ReadUser(selector *UserSelector) (*User, bool, error)
}

// UsersWriter defines methods for writing users.
type UsersWriter interface {
	CreateUser(user *User) error
	UpdateUser(user *User) error
	DeleteUser(selector *UserSelector) error
}

// UsersReadWriter combines reader and writer interfaces.
type UsersReadWriter interface {
	UsersReader
	UsersWriter
}
