package webapp

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/logger"
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

	app := &Application{
		Config: cfg,
		Logger: logger,
	}

	return app
}

// TODO:
func (app *Application) InitServices(db *sql.DB) {}

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
