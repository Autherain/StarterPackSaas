package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/autherain/test/internal/app"
	"github.com/autherain/test/internal/database"
	"github.com/autherain/test/internal/environment"
	"github.com/autherain/test/internal/errors"
	"github.com/autherain/test/internal/server"
	"github.com/autherain/test/internal/store"
	"github.com/autherain/test/internal/version"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	showVersion := flag.Bool("version", false, "display version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	env := environment.Parse()

	cfg := app.Config{
		BaseURL:  env.BaseURL,
		HTTPPort: env.HTTPPort,
		BasicAuth: struct {
			Username       string
			HashedPassword string
		}{
			Username:       env.BasicAuthUsername,
			HashedPassword: env.BasicAuthHashedPassword,
		},
		Cookie: struct {
			SecretKey string
		}{
			SecretKey: env.CookieSecretKey,
		},
		DB: struct {
			DSN string
		}{
			DSN: env.DSN(),
		},
		JWT: struct {
			SecretKey string
		}{
			SecretKey: env.JWTSecretKey,
		},
	}

	db, err := database.New(cfg.DB.DSN)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database", "error", err)
		}
	}()

	store, err := store.New(store.WithDB(db.DB))
	if err != nil {
		return err
	}

	application := app.New(cfg, db, store, logger, errors.New(logger))

	return server.ServeHTTP(application)
}
