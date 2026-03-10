// Package main is the entry point for the event booking API server.
// It loads configuration, connects to PostgreSQL, runs migrations,
// wires all layers (repo → service → handler), and starts the HTTP server.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/arunsuryan/event-booking-system/backend/internal/config"
	"github.com/arunsuryan/event-booking-system/backend/internal/db"
	"github.com/arunsuryan/event-booking-system/backend/internal/handler"
	"github.com/arunsuryan/event-booking-system/backend/internal/logger"
	"github.com/arunsuryan/event-booking-system/backend/internal/middleware"
	"github.com/arunsuryan/event-booking-system/backend/internal/repository"
	"github.com/arunsuryan/event-booking-system/backend/internal/response"
	"github.com/arunsuryan/event-booking-system/backend/internal/service"
	"github.com/gorilla/mux"
)

func main() {
	l := logger.New(logger.Options{})

	cfg, err := config.Load()
	if err != nil {
		l.Error("failed to load config", "error", err)
		os.Exit(1)
	}

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
	if err := db.RunMigrations(database, migrationsDir, l.Info); err != nil {
		l.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	l.Info("migrations complete")

	// --- Dependency wiring ---
	userRepo := repository.NewUserRepo(database)
	eventRepo := repository.NewEventRepo(database)
	bookingRepo := repository.NewBookingRepo(database, l)
	auditRepo := repository.NewAuditRepo(database)

	eventSvc := service.NewEventService(eventRepo, userRepo)
	bookingSvc := service.NewBookingService(bookingRepo, auditRepo)

	eventHandler := handler.NewEventHandler(eventSvc)
	userHandler := handler.NewUserHandler(eventSvc)
	bookingHandler := handler.NewBookingHandler(bookingSvc)
	auditHandler := handler.NewAuditHandler(bookingSvc)

	// --- Router ---
	r := mux.NewRouter()

	r.Use(middleware.CORS)
	r.Use(middleware.Logging(l))

	r.HandleFunc("/api/health", healthCheck).Methods(http.MethodGet)

	r.HandleFunc("/api/users", userHandler.ListUsers).Methods(http.MethodGet)

	r.HandleFunc("/api/events", eventHandler.ListEvents).Methods(http.MethodGet)
	r.HandleFunc("/api/events/{id}", eventHandler.GetEvent).Methods(http.MethodGet)

	r.HandleFunc("/api/events/{id}/book", bookingHandler.BookEvent).Methods(http.MethodPost)
	r.HandleFunc("/api/bookings/{id}/cancel", bookingHandler.CancelBooking).Methods(http.MethodPost)
	r.HandleFunc("/api/users/{id}/bookings", bookingHandler.GetUserBookings).Methods(http.MethodGet)

	r.HandleFunc("/api/audit", auditHandler.GetAuditLogs).Methods(http.MethodGet)

	// --- Server with graceful shutdown ---
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so shutdown signal handling doesn't block.
	go func() {
		l.Info("server started", "port", cfg.Port)
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for SIGINT or SIGTERM, then drain in-flight requests.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		l.Error("server forced to shutdown", "error", err)
	}
	l.Info("server stopped")
}

// healthCheck handles GET /api/health — a simple liveness probe.
func healthCheck(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, nil, "ok")
}

// migrationsPath returns the path to the SQL migrations directory.
// Checks MIGRATIONS_DIR env var first (set in Docker/production), then
// resolves relative to the source file (works with `go run` in development),
// and falls back to "./migrations" for compiled binaries.
func migrationsPath() string {
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir
	}
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		return filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	}
	return "migrations"
}
