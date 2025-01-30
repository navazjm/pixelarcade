package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestOpen(t *testing.T) {
	// Get absolute path to the project root's .env file
	rootEnvPath, err := filepath.Abs("../../../../.env") // ./internal/webapp/utils/database/database_test.go
	if err != nil {
		t.Fatalf("Error resolving .env path: %v", err)
	}

	// Load .env file
	err = godotenv.Load(rootEnvPath)
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	cfg := &Config{
		Dsn:          os.Getenv("PIXELARCADE_DB_DSN"),
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		MaxIdleTime:  5 * time.Minute,
	}

	db, err := Open(cfg)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}
}
