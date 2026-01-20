// Package main provides a standalone HTTP server for E2E testing.
// This server runs the same routes and handlers as the main Wails app,
// but without the desktop window, making it suitable for Playwright tests.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"trade-machine/config"
	"trade-machine/internal/api"
	"trade-machine/internal/app"
	"trade-machine/internal/settings"
	"trade-machine/observability"
	"trade-machine/repository"
)

func main() {
	// Initialize logger in development mode for tests
	observability.InitLogger(false)
	observability.InitMetrics()

	// Get configuration from environment
	port := os.Getenv("E2E_SERVER_PORT")
	if port == "" {
		port = "9090"
	}

	databaseURL := os.Getenv("E2E_DATABASE_URL")
	if databaseURL == "" {
		observability.Fatal("E2E_DATABASE_URL environment variable is required")
	}

	// Load config with minimal settings for testing
	cfg := &config.Config{
		HTTP: config.HTTPConfig{},
		Agent: config.AgentConfig{
			ConcurrencyLimit: 3,
			TimeoutSeconds:   30,
		},
	}

	ctx := context.Background()

	// Initialize database
	repo, err := repository.NewRepository(ctx, databaseURL)
	if err != nil {
		observability.Fatal("failed to connect to database", "error", err)
	}
	defer repo.Close()

	observability.Info("connected to test database")

	// Initialize app with minimal dependencies (no external API services for e2e)
	application := app.New(cfg, repo, nil, nil)
	application.Startup(ctx)

	// Initialize Settings Store with test directory
	settingsDir := os.Getenv("E2E_SETTINGS_DIR")
	if settingsDir == "" {
		settingsDir, err = os.MkdirTemp("", "trade-machine-e2e-settings-*")
		if err != nil {
			observability.Fatal("failed to create temp settings dir", "error", err)
		}
		defer os.RemoveAll(settingsDir)
	}

	settingsStore, err := settings.NewStore(settingsDir, "e2e-test-passphrase")
	if err != nil {
		observability.Fatal("failed to initialize settings store", "error", err)
	}
	application.SetSettings(settingsStore)
	observability.Info("settings store initialized", "dir", settingsDir)

	// Create HTTP router
	handler := api.NewHandler(application, cfg)
	router := api.NewRouter(handler, cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start server in goroutine
	go func() {
		observability.Info("starting E2E test server", "port", port, "url", fmt.Sprintf("http://localhost:%s", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			observability.Fatal("server error", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	observability.Info("shutting down E2E test server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		observability.Fatal("server forced to shutdown", "error", err)
	}

	application.Shutdown(shutdownCtx)
	observability.Info("E2E test server stopped")
}
