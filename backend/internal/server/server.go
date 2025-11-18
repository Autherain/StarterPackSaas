package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/autherain/test/internal/app"
	"github.com/autherain/test/internal/routes"
)

const (
	DefaultIdleTimeout    = time.Minute
	DefaultReadTimeout    = 5 * time.Second
	DefaultWriteTimeout   = 10 * time.Second
	DefaultShutdownPeriod = 30 * time.Second
)

func ServeHTTP(app *app.Application) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.HTTPPort),
		Handler:      routes.Setup(app),
		ErrorLog:     slog.NewLogLogger(app.Logger.Handler(), slog.LevelWarn),
		IdleTimeout:  DefaultIdleTimeout,
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
	}

	shutdownErrorChan := make(chan error)

	go func() {
		quitChan := make(chan os.Signal, 1)
		signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
		<-quitChan

		ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownPeriod)
		defer cancel()

		shutdownErrorChan <- srv.Shutdown(ctx)
	}()

	app.Logger.Info("starting server", slog.Group("server", "addr", srv.Addr))

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErrorChan
	if err != nil {
		return err
	}

	app.Logger.Info("stopped server", slog.Group("server", "addr", srv.Addr))

	app.WG.Wait()
	return nil
}
