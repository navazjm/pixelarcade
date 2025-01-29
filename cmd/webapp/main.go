package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/navazjm/pixelarcade/internal/webapp"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/database"
)

func main() {
	app := webapp.New()

	db, err := database.Open(&app.Config.DB)
	if err != nil {
		app.Logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	app.Logger.Info("database connection pool established")

	app.InitServices(db)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.Port),
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.Logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	app.Logger.Info("starting server", "port", srv.Addr, "env", app.Config.Env)
	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		app.Logger.Error(err.Error())
		os.Exit(1)
	}
	err = <-shutdownError
	if err != nil {
		app.Logger.Error(err.Error())
		os.Exit(1)
	}
	app.Logger.Info("stopped server", "addr", srv.Addr)
}
