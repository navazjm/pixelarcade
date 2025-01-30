package webapp

import (
	"context"
	"database/sql"
	"fmt"
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

	app := &Application{
		Config: cfg,
		Logger: logger,
	}

	return app
}

// TODO:
func (app *Application) InitServices(db *sql.DB) {}

// ============================================================================
// Mock App and slog.Logger for testing purposes
// ============================================================================

func setupTestApp() *Application {
	return &Application{
		Config: &Config{
			TrustedOrigins: []string{"https://example.com", "https://trusted.com"},
		},
		Logger: NewMockLogger(),
	}
}

// MockHandler is a mock implementation of slog.Handler for testing.
type MockHandler struct {
	LogOutput string
}

func (h *MockHandler) HandleLog(record slog.Record) error {
	h.LogOutput = fmt.Sprintf("Level: %s, Message: %s", record.Level, record.Message)
	return nil
}

func (m *MockHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (m *MockHandler) Handle(ctx context.Context, r slog.Record) error {
	return nil
}
func (m *MockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MockHandler{}
}
func (m *MockHandler) WithGroup(name string) slog.Handler {
	return &MockHandler{}
}

func NewMockLogger() *slog.Logger {
	handler := &MockHandler{}
	return slog.New(handler) // return a new logger using the mock handler
}
