package store

import (
	"fmt"

	"github.com/autherain/test/internal/user"
	userstore "github.com/autherain/test/internal/user/store"
	"github.com/jmoiron/sqlx"
)

// Store aggregates all domain stores.
type Store struct {
	Users user.UsersReadWriter
	db    *sqlx.DB
}

// Option configures the store.
type Option func(*Store) error

// New creates a new store with the given options.
func New(options ...Option) (*Store, error) {
	s := &Store{}

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, fmt.Errorf("could not create store: %w", err)
		}
	}

	// Initialize domain stores
	if s.db != nil {
		s.Users = userstore.New(s.db)
	}

	return s, nil
}

// WithDB sets the database connection.
func WithDB(db *sqlx.DB) Option {
	return func(s *Store) error {
		if err := db.Ping(); err != nil {
			return fmt.Errorf("could not ping database: %w", err)
		}

		s.db = db

		return nil
	}
}

