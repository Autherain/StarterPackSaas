package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/autherain/test/internal/app"
	"github.com/autherain/test/internal/database"
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
	var cfg app.Config

	flag.StringVar(&cfg.BaseURL, "base-url", "http://localhost:9798", "base URL for the application")
	flag.IntVar(&cfg.HTTPPort, "http-port", 9798, "port to listen on for HTTP requests")
	flag.StringVar(&cfg.BasicAuth.Username, "basic-auth-username", "admin", "basic auth username")
	flag.StringVar(&cfg.BasicAuth.HashedPassword, "basic-auth-hashed-password", "$2a$10$jRb2qniNcoCyQM23T59RfeEQUbgdAXfR6S0scynmKfJa5Gj3arGJa", "basic auth password hashed with bcrpyt")
	flag.StringVar(&cfg.Cookie.SecretKey, "cookie-secret-key", "mflpw6hs4mdzads3s5kxjgtoltlgp3sp", "secret key for cookie authentication/encryption")
	flag.StringVar(&cfg.DB.DSN, "db-dsn", "user:pass@localhost:5432/db", "postgreSQL DSN")
	flag.BoolVar(&cfg.DB.Automigrate, "db-automigrate", true, "run migrations on startup")
	flag.StringVar(&cfg.JWT.SecretKey, "jwt-secret-key", "q54isdosxiujnhjwmxrscqohr2tfm2c7", "secret key for JWT authentication")

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	db, err := database.New(cfg.DB.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if cfg.DB.Automigrate {
		err = db.MigrateUp()
		if err != nil {
			return err
		}
	}

	store, err := store.New(store.WithDB(db.DB))
	if err != nil {
		return err
	}

	errorHandler := errors.New(logger)
	application := app.New(cfg, db, store, logger, errorHandler)

	return server.ServeHTTP(application)
}
