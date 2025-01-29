package webapp

import (
	"database/sql"
	"log/slog"
	"os"
)

type Application struct {
	Config *Config
	Logger *slog.Logger
}

func New() *Application {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := NewConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	srv := &Application{
		Config: cfg,
		Logger: logger,
	}

	return srv
}

func (app *Application) InitServices(db *sql.DB) {}
