package webapp

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/navazjm/pixelarcade/internal/webapp/auth"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/logger"
)

type Application struct {
	Config      *Config
	Logger      *slog.Logger
	AuthService *auth.Service
}

func New() *Application {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg, err := NewConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &Application{
		Config: cfg,
		Logger: logger,
	}

	return app
}

func (app *Application) InitServices(db *sql.DB) {
	app.AuthService = auth.NewService(db, app.Logger)
}

// ============================================================================
// Mock App for testing purposes
// ============================================================================

func setupTestApp() *Application {
	return &Application{
		Config: &Config{
			TrustedOrigins: []string{"https://example.com", "https://trusted.com"},
		},
		Logger: logger.NewMock(),
	}
}
