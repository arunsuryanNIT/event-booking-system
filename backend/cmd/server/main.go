// Package main is the entry point for the event booking API server.
// It loads configuration, connects to PostgreSQL, runs migrations, and starts the HTTP server.
package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/arunsuryan/event-booking-system/backend/internal/config"
	"github.com/arunsuryan/event-booking-system/backend/internal/db"
	"github.com/arunsuryan/event-booking-system/backend/internal/logger"
)

func main() {
	// Start with a default stdout/info logger so every code path has structured logging.
	l := logger.New(logger.Options{})

	cfg, err := config.Load()
	if err != nil {
		l.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Reconfigure logger with user-specified output and level from config.
	l = logger.New(logger.Options{
		Output: cfg.LogOutput,
		Level:  cfg.LogLevel,
	})

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		l.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()
	l.Info("connected to database")

	migrationsDir := migrationsPath()
	if err = db.RunMigrations(database, migrationsDir, l.Info); err != nil {
		l.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	l.Info("migrations complete")

	l.Info("server starting", "port", cfg.Port)
	// Router and server setup will be added in a later chunk
}

// migrationsPath resolves the migrations directory relative to the binary
// or relative to the source file during development.
func migrationsPath() string {
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		return filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	}
	return "migrations"
}
