package auth

import (
	"database/sql"
	"log/slog"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/logger"
)

type Service struct {
	Models Model
	Logger *slog.Logger
}

func NewService(db *sql.DB, logger *slog.Logger) *Service {
	return &Service{
		Models: Model{DB: db},
		Logger: logger,
	}
}

func newMockService(t *testing.T) (*Service, sqlmock.Sqlmock) {
	// Create mock database
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock DB: %v", err)
	}

	// Create service with mock DB
	service := &Service{
		Models: Model{DB: mockDB},
		Logger: logger.NewMock(),
	}

	return service, mock
}
