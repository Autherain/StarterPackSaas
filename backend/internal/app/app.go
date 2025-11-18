package app

import (
	"log/slog"
	"sync"

	"github.com/autherain/test/internal/database"
	"github.com/autherain/test/internal/errors"
	"github.com/autherain/test/internal/store"
	"github.com/autherain/test/internal/user/server"
)

type Config struct {
	BaseURL   string
	HTTPPort  int
	BasicAuth struct {
		Username       string
		HashedPassword string
	}
	Cookie struct {
		SecretKey string
	}
	DB struct {
		DSN         string
		Automigrate bool
	}
	JWT struct {
		SecretKey string
	}
}

type Application struct {
	Config       Config
	DB           *database.DB
	Store        *store.Store
	UserServer   *server.Server
	Logger       *slog.Logger
	ErrorHandler *errors.Handler
	WG           sync.WaitGroup
}

func New(cfg Config, db *database.DB, store *store.Store, logger *slog.Logger, errorHandler *errors.Handler) *Application {
	app := &Application{
		Config:       cfg,
		DB:           db,
		Store:        store,
		Logger:       logger,
		ErrorHandler: errorHandler,
	}

	// Initialize user server
	app.UserServer = server.New(
		app.Store.Users,
		app.ErrorHandler.ServerError,
		app.ErrorHandler.FailedValidation,
	)

	return app
}
