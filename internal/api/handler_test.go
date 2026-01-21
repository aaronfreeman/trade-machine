package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trade-machine/config"
	"trade-machine/internal/app"
	"trade-machine/internal/settings"
	"trade-machine/repository"
)

// mockSettingsRepository implements settings.RepositoryInterface for testing
type mockSettingsRepository struct {
	apiKeys map[string]*settings.APIKeyModel
}

func newMockSettingsRepository() *mockSettingsRepository {
	return &mockSettingsRepository{
		apiKeys: make(map[string]*settings.APIKeyModel),
	}
}

func (m *mockSettingsRepository) GetAPIKey(ctx context.Context, serviceName string) (*settings.APIKeyModel, error) {
	key, ok := m.apiKeys[serviceName]
	if !ok {
		return nil, errors.New("not found")
	}
	return key, nil
}

func (m *mockSettingsRepository) GetAllAPIKeys(ctx context.Context) ([]settings.APIKeyModel, error) {
	var keys []settings.APIKeyModel
	for _, key := range m.apiKeys {
		keys = append(keys, *key)
	}
	return keys, nil
}

func (m *mockSettingsRepository) UpsertAPIKey(ctx context.Context, apiKey *settings.APIKeyModel) error {
	m.apiKeys[apiKey.ServiceName] = apiKey
	return nil
}

func (m *mockSettingsRepository) DeleteAPIKey(ctx context.Context, serviceName string) error {
	delete(m.apiKeys, serviceName)
	return nil
}

// testConfig returns a test configuration
func testConfig() *config.Config {
	return config.NewTestConfig()
}

// testApp creates an App with test config for testing
func testApp(repo app.RepositoryInterface) *app.App {
	return app.New(testConfig(), repo, nil, nil)
}

// testAppWithSettings creates an App with test config and settings store
func testAppWithSettings(t *testing.T) *app.App {
	t.Helper()
	tmpDir := t.TempDir()
	mockRepo := newMockSettingsRepository()
	store, err := settings.NewStore(tmpDir, "test-passphrase", mockRepo)
	if err != nil {
		t.Fatalf("failed to create settings store: %v", err)
	}
	a := app.New(testConfig(), nil, nil, nil)
	a.SetSettings(store)
	return a
}

// testHandler creates a Handler with test config for testing
func testHandler(application *app.App) *Handler {
	return NewHandler(application, testConfig())
}

// testRouter creates a Chi router with test config for testing
func testRouter(application *app.App) http.Handler {
	cfg := testConfig()
	handler := NewHandler(application, cfg)
	return NewRouter(handler, cfg)
}

func TestHandler_Index(t *testing.T) {
	t.Run("serves templ index at root", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/html") {
			t.Errorf("expected Content-Type text/html, got %s", contentType)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Trade Machine") {
			t.Error("expected body to contain 'Trade Machine'")
		}
	})

	t.Run("serves templ index at /index.html", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("index method not allowed", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", w.Code)
		}
	})
}

func TestHandler_Health(t *testing.T) {
	t.Run("health check without database", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if status, ok := response["status"].(string); !ok || status != "ok" {
			t.Errorf("expected status ok, got %v", response["status"])
		}
	})
}

func TestHandler_AnalyzeStock(t *testing.T) {
	t.Run("missing symbol", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("portfolio manager not initialized", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader(`{"symbol":"AAPL"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_GetRecommendations(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/recommendations", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("with limit parameter", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/recommendations?limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_ApproveRecommendation(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		ctx := context.Background()
		connString := "host=localhost port=5432 user=postgres password=postgres dbname=trademachine_test sslmode=disable"
		repo, err := repository.NewRepository(ctx, connString)
		if err != nil {
			t.Skip("database not available")
		}
		defer repo.Close()

		a := testApp(repo)
		a.Startup(ctx)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/recommendations/invalid-uuid/approve", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_GetPositions(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/positions", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_GetTrades(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/trades", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("with limit parameter", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/trades?limit=25", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_GetAgentRuns(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/agents/runs", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_NotFound(t *testing.T) {
	a := testApp(nil)
	router := testRouter(a)

	req := httptest.NewRequest(http.MethodGet, "/api/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	a := testApp(nil)
	router := testRouter(a)

	req := httptest.NewRequest(http.MethodDelete, "/api/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHandler_ParseLimitParam(t *testing.T) {
	tests := []struct {
		name         string
		queryParam   string
		defaultLimit int
		expected     int
	}{
		{"no parameter", "", 50, 50},
		{"valid limit", "limit=25", 50, 25},
		{"invalid limit", "limit=abc", 50, 50},
		{"negative limit", "limit=-10", 50, 50},
		{"zero limit", "limit=0", 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := testApp(nil)
			handler := testHandler(a)

			url := "/api/test"
			if tt.queryParam != "" {
				url += "?" + tt.queryParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			result := handler.ParseLimitParam(req, tt.defaultLimit)

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestHandler_ValidateSymbol(t *testing.T) {
	a := testApp(nil)
	handler := testHandler(a)

	tests := []struct {
		name      string
		symbol    string
		wantError bool
	}{
		{"valid simple symbol", "AAPL", false},
		{"valid with number", "BRK.B", false},
		{"valid with dash", "BRK-B", false},
		{"valid long symbol", "ABCDEFGHIJ", false},
		{"empty symbol", "", true},
		{"too long", "ABCDEFGHIJK", true},
		{"lowercase", "aapl", true},
		{"special chars", "AAPL!", true},
		{"spaces", "AA PL", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateSymbol(tt.symbol)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateSymbol(%s) error = %v, wantError %v", tt.symbol, err, tt.wantError)
			}
		})
	}
}

func TestHandler_AnalyzeStock_InvalidSymbol(t *testing.T) {
	a := testApp(nil)
	router := testRouter(a)

	tests := []struct {
		name   string
		symbol string
	}{
		{"empty", ""},
		{"too long", "ABCDEFGHIJK"},
		{"special chars", "AAPL!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/analyze",
				strings.NewReader(`{"symbol":"`+tt.symbol+`"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", w.Code)
			}
		})
	}
}

func TestHandler_GetPendingRecommendations(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/recommendations/pending", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_RejectRecommendation(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/recommendations/550e8400-e29b-41d4-a716-446655440000/reject", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("invalid UUID", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/recommendations/invalid-uuid/reject", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_GetPortfolio(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/portfolio", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

// Integration tests with database
func TestIntegration_WithDatabase(t *testing.T) {
	ctx := context.Background()
	connString := "postgres://postgres:postgres@localhost:5432/trademachine_test?sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available")
	}
	defer repo.Close()

	a := testApp(repo)
	a.Startup(ctx)
	router := testRouter(a)

	t.Run("get recommendations with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/recommendations?limit=5", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get pending recommendations with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/recommendations/pending", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get positions with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/positions", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get portfolio with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/portfolio", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get trades with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/trades?limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get agent runs with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/agents/runs?limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("health check with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if status, ok := response["status"].(string); !ok || status != "ok" {
			t.Errorf("expected status ok, got %v", response["status"])
		}

		services := response["services"].(map[string]interface{})
		if dbStatus, ok := services["database"].(string); !ok || dbStatus != "connected" {
			t.Errorf("expected database connected, got %v", services["database"])
		}
	})

	t.Run("approve nonexistent recommendation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/recommendations/550e8400-e29b-41d4-a716-446655440000/approve", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("reject nonexistent recommendation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/recommendations/550e8400-e29b-41d4-a716-446655440000/reject", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestHandler_MethodsNotAllowed(t *testing.T) {
	a := testApp(nil)
	router := testRouter(a)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"health with POST", http.MethodPost, "/api/health"},
		{"portfolio with POST", http.MethodPost, "/api/portfolio"},
		{"positions with POST", http.MethodPost, "/api/positions"},
		{"analyze with GET", http.MethodGet, "/api/analyze"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status 405, got %d", w.Code)
			}
		})
	}
}

func TestHandler_CORSHeaders(t *testing.T) {
	a := testApp(nil)
	router := testRouter(a)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("missing CORS Allow-Origin header")
	}
}

func TestHandler_OptionsRequest(t *testing.T) {
	a := testApp(nil)
	router := testRouter(a)

	req := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for OPTIONS, got %d", w.Code)
	}
}

func TestHandler_GetRecommendations_WithStatus(t *testing.T) {
	ctx := context.Background()
	connString := "postgres://postgres:postgres@localhost:5432/trademachine_test?sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available")
	}
	defer repo.Close()

	a := testApp(repo)
	a.Startup(ctx)
	router := testRouter(a)

	req := httptest.NewRequest(http.MethodGet, "/api/recommendations?status=pending&limit=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandler_AnalyzeStock_InvalidJSON(t *testing.T) {
	a := testApp(nil)
	router := testRouter(a)

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandler_GetAgentRuns_WithType(t *testing.T) {
	ctx := context.Background()
	connString := "postgres://postgres:postgres@localhost:5432/trademachine_test?sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available")
	}
	defer repo.Close()

	a := testApp(repo)
	a.Startup(ctx)
	router := testRouter(a)

	req := httptest.NewRequest(http.MethodGet, "/api/agents/runs?type=fundamental&limit=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandler_RunScreener(t *testing.T) {
	t.Run("screener not configured", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/screener/run", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})
}

func TestHandler_GetLatestScreenerRun(t *testing.T) {
	t.Run("screener not configured", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/screener/latest", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})
}

func TestHandler_GetScreenerRuns(t *testing.T) {
	t.Run("screener not configured", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/screener/runs", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})

	t.Run("screener not configured with limit", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/screener/runs?limit=5", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})
}

func TestHandler_GetScreenerRun(t *testing.T) {
	t.Run("screener not configured", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/screener/runs/550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})
}

func TestHandler_GetTopPicks(t *testing.T) {
	t.Run("screener not configured", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodGet, "/api/screener/picks", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})
}

func TestHandler_UpdateAPIKey(t *testing.T) {
	t.Run("settings not available", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader("service_name=fmp&api_key=test123"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})

	t.Run("missing service name in form data", func(t *testing.T) {
		a := testAppWithSettings(t)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader("api_key=test123"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("missing service name in JSON", func(t *testing.T) {
		a := testAppWithSettings(t)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader(`{"api_key":"test123"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		a := testAppWithSettings(t)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader(`{invalid json`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("valid form data saves API key", func(t *testing.T) {
		a := testAppWithSettings(t)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader("service_name=fmp&api_key=test123"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("valid JSON saves API key", func(t *testing.T) {
		a := testAppWithSettings(t)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader(`{"service_name":"fmp","api_key":"test456"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("preserves existing key when submitting empty form", func(t *testing.T) {
		a := testAppWithSettings(t)
		router := testRouter(a)

		// First, save an API key
		req := httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader("service_name=openai&api_key=sk-test-key-12345"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 when saving key, got %d: %s", w.Code, w.Body.String())
		}

		// Verify the key was saved
		savedKey := a.Settings().GetAPIKey("openai")
		if savedKey == nil || savedKey.APIKey != "sk-test-key-12345" {
			t.Fatalf("expected key to be saved, got %v", savedKey)
		}

		// Now submit the form with an empty API key field (simulating user clicking save without entering anything)
		req = httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader("service_name=openai&api_key="))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 when submitting empty form, got %d: %s", w.Code, w.Body.String())
		}

		// Verify the original key is still preserved
		preservedKey := a.Settings().GetAPIKey("openai")
		if preservedKey == nil {
			t.Fatal("expected key to be preserved, got nil")
		}
		if preservedKey.APIKey != "sk-test-key-12345" {
			t.Errorf("expected key to be preserved as 'sk-test-key-12345', got %q", preservedKey.APIKey)
		}
	})

	t.Run("updates key when new value provided", func(t *testing.T) {
		a := testAppWithSettings(t)
		router := testRouter(a)

		// First, save an API key
		req := httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader("service_name=openai&api_key=sk-old-key"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 when saving key, got %d", w.Code)
		}

		// Now update with a new key
		req = httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader("service_name=openai&api_key=sk-new-key"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 when updating key, got %d", w.Code)
		}

		// Verify the key was updated
		updatedKey := a.Settings().GetAPIKey("openai")
		if updatedKey == nil {
			t.Fatal("expected key to exist, got nil")
		}
		if updatedKey.APIKey != "sk-new-key" {
			t.Errorf("expected key to be updated to 'sk-new-key', got %q", updatedKey.APIKey)
		}
	})

	t.Run("preserves other fields when updating only API key", func(t *testing.T) {
		a := testAppWithSettings(t)
		router := testRouter(a)

		// Save Alpaca config with key, secret, and base URL
		req := httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader("service_name=alpaca&api_key=AKID123&api_secret=secret456&base_url=https://paper-api.alpaca.markets"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 when saving config, got %d", w.Code)
		}

		// Now update only the API key, leaving other fields empty
		req = httptest.NewRequest(http.MethodPost, "/api/settings/api-keys",
			strings.NewReader("service_name=alpaca&api_key=AKID789&api_secret=&base_url="))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 when updating key, got %d", w.Code)
		}

		// Verify the key was updated but other fields were preserved
		config := a.Settings().GetAPIKey("alpaca")
		if config == nil {
			t.Fatal("expected config to exist, got nil")
		}
		if config.APIKey != "AKID789" {
			t.Errorf("expected API key to be updated to 'AKID789', got %q", config.APIKey)
		}
		if config.APISecret != "secret456" {
			t.Errorf("expected API secret to be preserved as 'secret456', got %q", config.APISecret)
		}
		if config.BaseURL != "https://paper-api.alpaca.markets" {
			t.Errorf("expected base URL to be preserved as 'https://paper-api.alpaca.markets', got %q", config.BaseURL)
		}
	})
}

func TestHandler_AnalyzeStock_FormData(t *testing.T) {
	t.Run("form data with symbol", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/analyze",
			strings.NewReader("symbol=AAPL"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should fail with 500 because portfolio manager not initialized, not 400
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("form data without symbol", func(t *testing.T) {
		a := testApp(nil)
		router := testRouter(a)

		req := httptest.NewRequest(http.MethodPost, "/api/analyze",
			strings.NewReader(""))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}
