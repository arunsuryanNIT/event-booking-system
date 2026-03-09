// Package config handles application configuration loaded from environment variables.
package config

import (
	"errors"
	"os"
)

// Config holds all configuration values for the application.
type Config struct {
	DatabaseURL string // Postgres connection string (required)
	Port        string // HTTP server listen port (default "8080")
	LogOutput   string // Log destination: "stdout", "stderr", or a file path (default "stdout")
	LogLevel    string // Minimum log level: "debug", "info", "warn", "error" (default "info")
}

// Load reads configuration from environment variables and returns a Config.
// DATABASE_URL is required; all other values have sensible defaults.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logOutput := os.Getenv("LOG_OUTPUT")
	if logOutput == "" {
		logOutput = "stdout"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		DatabaseURL: dbURL,
		Port:        port,
		LogOutput:   logOutput,
		LogLevel:    logLevel,
	}, nil
}
