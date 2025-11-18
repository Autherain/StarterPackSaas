package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/autherain/test/internal/user"
	"github.com/jmoiron/sqlx"
)

const defaultTimeout = 3 * time.Second

// usersStore implements the user.UsersReadWriter interface.
type usersStore struct {
	db *sqlx.DB
}

var _ user.UsersReadWriter = (*usersStore)(nil)

// New creates a new users store with the given database connection.
func New(db *sqlx.DB) user.UsersReadWriter {
	return &usersStore{db: db}
}

// CreateUser creates a new user.
func (s *usersStore) CreateUser(u *user.User) error {
	// Set defaults if not specified
	if u.Created.IsZero() {
		u.Created = time.Now()
	}
	if u.SubscriptionTier == "" {
		u.SubscriptionTier = "free"
	}

	if err := u.Validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO users (created, email, hashed_password, subscription_tier)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	err := s.db.GetContext(ctx, &u.ID, query, u.Created, u.Email, u.HashedPassword, u.SubscriptionTier)
	return err
}

// ReadUser reads a user by ID.
func (s *usersStore) ReadUser(selector *user.UserSelector) (*user.User, bool, error) {
	if selector.ID != 0 {
		return s.readUserByID(selector.ID)
	}
	if selector.Email != "" {
		return s.ReadUserByEmail(selector.Email)
	}
	return nil, false, errors.New("user selector must have either ID or Email set")
}

// readUserByID is an internal helper to read a user by ID.
func (s *usersStore) readUserByID(id int) (*user.User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var u user.User

	query := `SELECT id, created, email, hashed_password, subscription_tier FROM users WHERE id = $1`

	err := s.db.GetContext(ctx, &u, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}

	return &u, true, err
}

// ReadUserByEmail reads a user by email address.
func (s *usersStore) ReadUserByEmail(email string) (*user.User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	var u user.User

	query := `SELECT id, created, email, hashed_password, subscription_tier FROM users WHERE LOWER(email) = LOWER($1)`

	err := s.db.GetContext(ctx, &u, query, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}

	return &u, true, err
}

// UpdateUser updates a user.
func (s *usersStore) UpdateUser(u *user.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		UPDATE users 
		SET email = $1, hashed_password = $2, subscription_tier = $3
		WHERE id = $4`

	_, err := s.db.ExecContext(ctx, query, u.Email, u.HashedPassword, u.SubscriptionTier, u.ID)
	return err
}

// DeleteUser deletes a user.
func (s *usersStore) DeleteUser(selector *user.UserSelector) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `DELETE FROM users WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, selector.ID)
	return err
}
