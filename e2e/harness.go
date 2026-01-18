// Package e2e provides end-to-end testing infrastructure for trade-machine.
package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"trade-machine/config"
	"trade-machine/e2e/mocks"
	"trade-machine/internal/api"
	"trade-machine/internal/app"
	"trade-machine/repository"
)

// TestHarness provides the infrastructure for running E2E tests.
type TestHarness struct {
	t          *testing.T
	ctx        context.Context
	cancel     context.CancelFunc
	mockServer *mocks.MockServer
	repo       *repository.Repository
	app        *app.App
	router     http.Handler
	config     *config.Config
}

// NewTestHarness creates a new test harness with all dependencies initialized.
func NewTestHarness(t *testing.T) *TestHarness {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	h := &TestHarness{
		t:      t,
		ctx:    ctx,
		cancel: cancel,
	}

	return h
}

// Setup initializes all test dependencies.
func (h *TestHarness) Setup() error {
	// Start mock server for external APIs
	h.mockServer = mocks.NewMockServer()

	// Create test configuration
	h.config = h.createTestConfig()

	// Connect to test database
	dbURL := os.Getenv("E2E_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://trademachine_test:test_password@localhost:5433/trademachine_test?sslmode=disable"
	}

	var err error
	h.repo, err = repository.NewRepository(h.ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Run migrations
	if err := h.runMigrations(dbURL); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create application
	h.app = app.New(h.config, h.repo, nil, nil)
	h.app.Startup(h.ctx)

	// Create router
	handler := api.NewHandler(h.app, h.config)
	h.router = api.NewRouter(handler, h.config)

	return nil
}

// Teardown cleans up all test resources.
func (h *TestHarness) Teardown() {
	if h.cancel != nil {
		h.cancel()
	}

	if h.app != nil {
		h.app.Shutdown(context.Background())
	}

	if h.repo != nil {
		h.cleanupTestData()
		h.repo.Close()
	}

	if h.mockServer != nil {
		h.mockServer.Close()
	}
}

// Context returns the test context.
func (h *TestHarness) Context() context.Context {
	return h.ctx
}

// MockServer returns the mock server for configuring responses.
func (h *TestHarness) MockServer() *mocks.MockServer {
	return h.mockServer
}

// Repository returns the test database repository.
func (h *TestHarness) Repository() *repository.Repository {
	return h.repo
}

// App returns the application instance.
func (h *TestHarness) App() *app.App {
	return h.app
}

// Router returns the HTTP router for making requests.
func (h *TestHarness) Router() http.Handler {
	return h.router
}

// Config returns the test configuration.
func (h *TestHarness) Config() *config.Config {
	return h.config
}

// DoRequest performs an HTTP request and returns the response.
func (h *TestHarness) DoRequest(method, path string, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, stringReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	return w
}

// DoHTMXRequest performs an HTMX request and returns the response.
func (h *TestHarness) DoHTMXRequest(method, path string, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, stringReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("HX-Request", "true")

	w := httptest.NewRecorder()
	h.router.ServeHTTP(w, req)
	return w
}

// ResetDatabase clears all test data from the database.
func (h *TestHarness) ResetDatabase() error {
	return h.cleanupTestData()
}

func (h *TestHarness) createTestConfig() *config.Config {
	mockURL := h.mockServer.URL()

	// Create a test config with mock server URLs
	cfg := config.NewTestConfig()

	// Override external service URLs to point to mock server
	// These would need to be set via environment or config fields
	os.Setenv("ALPACA_BASE_URL", mockURL)
	os.Setenv("ALPHA_VANTAGE_BASE_URL", mockURL)
	os.Setenv("NEWS_API_BASE_URL", mockURL)
	os.Setenv("FMP_BASE_URL", mockURL)
	os.Setenv("BEDROCK_ENDPOINT", mockURL)

	return cfg
}

func (h *TestHarness) runMigrations(dbURL string) error {
	// Find migrations directory
	migrationsDir := findMigrationsDir()
	if migrationsDir == "" {
		return fmt.Errorf("migrations directory not found")
	}

	// Use migrate CLI if available, otherwise skip
	_, err := exec.LookPath("migrate")
	if err != nil {
		h.t.Log("migrate CLI not found, skipping migrations (assuming schema exists)")
		return nil
	}

	cmd := exec.CommandContext(h.ctx, "migrate", "-path", migrationsDir, "-database", dbURL, "up")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Ignore "no change" errors
		if string(output) == "" || contains(string(output), "no change") {
			return nil
		}
		return fmt.Errorf("migration failed: %s: %w", string(output), err)
	}

	return nil
}

func (h *TestHarness) cleanupTestData() error {
	queries := []string{
		"DELETE FROM agent_runs",
		"DELETE FROM trades",
		"DELETE FROM recommendations",
		"DELETE FROM positions",
		"DELETE FROM screener_runs",
		"DELETE FROM market_data_cache",
	}

	for _, q := range queries {
		if _, err := h.repo.Pool().Exec(h.ctx, q); err != nil {
			h.t.Logf("cleanup query failed (may be ok if table doesn't exist): %s: %v", q, err)
		}
	}

	return nil
}

func findMigrationsDir() string {
	// Try common locations
	candidates := []string{
		"migrations",
		"../migrations",
		"../../migrations",
	}

	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}

	return ""
}

func stringReader(s string) *stringReaderImpl {
	return &stringReaderImpl{s: s, i: 0}
}

type stringReaderImpl struct {
	s string
	i int
}

func (r *stringReaderImpl) Read(p []byte) (n int, err error) {
	if r.i >= len(r.s) {
		return 0, fmt.Errorf("EOF")
	}
	n = copy(p, r.s[r.i:])
	r.i += n
	return n, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SkipIfNoDatabase skips the test if the database is not available.
func SkipIfNoDatabase(t *testing.T) {
	t.Helper()

	dbURL := os.Getenv("E2E_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://trademachine_test:test_password@localhost:5433/trademachine_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo, err := repository.NewRepository(ctx, dbURL)
	if err != nil {
		t.Skipf("E2E database not available: %v", err)
	}
	repo.Close()
}

// RequireDockerCompose ensures the docker-compose test environment is running.
func RequireDockerCompose(t *testing.T) {
	t.Helper()

	// Check if we can connect to the test database
	SkipIfNoDatabase(t)
}
