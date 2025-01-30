package webapp

import (
	"flag"
	"os"
	"testing"
	"time"
)

func clearEnvVars() {
	os.Unsetenv("PIXELARCADE_DB_DSN")
}

func resetFlags() {
	// Reset flags to avoid conflicts
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Manually parse flags, ignoring test flags
	os.Args = os.Args[:1]
	flag.CommandLine.Parse(os.Args[1:])
}

func TestNewConfig_WithoutFlags(t *testing.T) {
	clearEnvVars()
	os.Setenv("PIXELARCADE_DB_DSN", "postgres://user:password@localhost/dbname")
	resetFlags()

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if cfg.Version != "0.1.0" {
		t.Errorf("expected version 0.1.0, got %s", cfg.Version)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.Env != "prod" {
		t.Errorf("expected env 'prod', got %s", cfg.Env)
	}
	if cfg.DB.MaxOpenConns != 25 {
		t.Errorf("expected DB max open connections 25, got %d", cfg.DB.MaxOpenConns)
	}
	if cfg.DB.MaxIdleConns != 25 {
		t.Errorf("expected DB max idle connections 25, got %d", cfg.DB.MaxIdleConns)
	}
	if cfg.DB.MaxIdleTime != 15*time.Minute {
		t.Errorf("expected DB max idle time 15 minutes, got %v", cfg.DB.MaxIdleTime)
	}
	if len(cfg.TrustedOrigins) == 0 {
		t.Errorf("expected trusted origins to be populated")
	}
}

func TestNewConfig_WithFlags(t *testing.T) {
	clearEnvVars()
	resetFlags()

	os.Args = []string{
		"cmd/webapp", "-env", "test", "-port", "9090", "-db-dsn", "postgres://u:p@localhost/db",
	}

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.Env != "test" {
		t.Errorf("expected env 'test', got %s", cfg.Env)
	}
	if cfg.DB.Dsn != "postgres://u:p@localhost/db" {
		t.Errorf("expected DB DSN 'postgres://u:p@localhost/db', got %s", cfg.DB.Dsn)
	}
}

func TestNewConfig_MissingDbDsn(t *testing.T) {
	clearEnvVars()
	resetFlags()

	_, err := NewConfig()
	if err == nil {
		t.Fatal("expected error, but got none")
	}

	// Check the error message
	expectedErr := "missing required DB DSN in environment or flags"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

// Should attempt to run godotenv.Load() but fail since .env is located in
// project root and not next to config_test.go file. FYI, when running 'go test'
// it changes the cwd to the package not the root of the project hence why this
// should error
func TestNewConfig_WithFlagEnvDev(t *testing.T) {
	clearEnvVars()
	resetFlags()

	os.Args = []string{
		"cmd/webapp", "-env", "dev",
	}

	_, err := NewConfig()
	if err == nil {
		t.Fatal("expected error, but got none")
	}

	// Check the error message
	expectedErr := "open .env: no such file or directory"
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
}
