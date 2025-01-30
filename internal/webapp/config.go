package webapp

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/database"
)

type Config struct {
	Port           int
	Env            string
	Version        string
	DB             database.Config
	TrustedOrigins []string
}

func NewConfig() (*Config, error) {
	var err error
	cfg := &Config{}

	// default config for PROD

	cfg.Version = "0.1.0"
	cfg.TrustedOrigins = []string{"https://pixelarcade.dev"} // TODO: update when ready to deploy

	flag.IntVar(&cfg.Port, "port", 8080, "Server port")
	flag.StringVar(&cfg.Env, "env", "prod", "Environment (dev|prod)")
	flag.StringVar(&cfg.DB.Dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.DB.MaxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.DB.MaxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.DB.MaxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")
	flag.Parse()

	if cfg.Env == "dev" {
		err = godotenv.Load()
		if err != nil {
			return nil, err
		}
		// allow requests from frontend client
		cfg.TrustedOrigins = []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	}

	// Use the env variable for DSN if the flag is not provided
	if cfg.DB.Dsn == "" {
		cfg.DB.Dsn = os.Getenv("PIXELARCADE_DB_DSN")
		if cfg.DB.Dsn == "" {
			return nil, fmt.Errorf("missing required DB DSN in environment or flags")
		}
	}

	return cfg, nil
}
