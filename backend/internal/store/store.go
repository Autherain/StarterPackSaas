package store

import (
	"fmt"

	"github.com/autherain/test/internal/user"
	userstore "github.com/autherain/test/internal/user/store"
	"github.com/jmoiron/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Store aggregates all domain stores.
type Store struct {
	Users user.UsersReadWriter
	db    *gorm.DB
}

// Option configures the store.
type Option func(*Store) error

// New creates a new store with the given options.
func New(options ...Option) (*Store, error) {
	store := &Store{}

	store.Users = userstore.New(store)

	for _, option := range options {
		if err := option(store); err != nil {
			return nil, fmt.Errorf("could not create store: %w", err)
		}
	}

	return store, nil
}

// WithDB sets the database connection by converting sqlx.DB to GORM.
func WithDB(db *sqlx.DB) Option {
	return func(s *Store) error {
		if err := db.Ping(); err != nil {
			return fmt.Errorf("could not ping database: %w", err)
		}

		// Convert sqlx.DB to GORM
		gormDB, err := gorm.Open(postgres.New(postgres.Config{
			Conn: db.DB,
		}), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("could not create GORM connection: %w", err)
		}

		s.db = gormDB

		return nil
	}
}

// GetDB returns the database connection.
// This method implements the baseStore interface for domain stores.
func (s *Store) GetDB() *gorm.DB {
	return s.db
}
