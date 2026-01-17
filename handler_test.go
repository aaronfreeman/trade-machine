package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trade-machine/config"
	"trade-machine/repository"
)

// testConfig returns a test configuration
func testConfig() *config.Config {
	return config.NewTestConfig()
}

// testApp creates an App with test config for testing
func testApp(repo *repository.Repository) *App {
	return NewApp(testConfig(), repo, nil, nil)
}

// testHandler creates an APIHandler with test config for testing
func testHandler(app *App) *APIHandler {
	return NewAPIHandler(app, testConfig())
}

func TestAPIHandler_Index(t *testing.T) {
	t.Run("serves templ index at root", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

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
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("index method not allowed", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", w.Code)
		}
	})
}

func TestAPIHandler_Health(t *testing.T) {
	tests := []struct {
		name           string
		repo           *repository.Repository
		expectedStatus string
	}{
		{
			name:           "health check without database",
			repo:           nil,
			expectedStatus: "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := testApp(tt.repo)
			handler := testHandler(app)

			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if status, ok := response["status"].(string); !ok || status != tt.expectedStatus {
				t.Errorf("expected status %s, got %v", tt.expectedStatus, response["status"])
			}
		})
	}
}

func TestAPIHandler_AnalyzeStock(t *testing.T) {
	t.Run("missing symbol", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("portfolio manager not initialized", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader(`{"symbol":"AAPL"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestAPIHandler_GetRecommendations(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodGet, "/api/recommendations", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("with limit parameter", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodGet, "/api/recommendations?limit=10", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestAPIHandler_ApproveRecommendation(t *testing.T) {
	t.Run("invalid UUID", func(t *testing.T) {
		ctx := context.Background()
		connString := "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable"
		repo, err := repository.NewRepository(ctx, connString)
		if err != nil {
			t.Skip("database not available")
		}
		defer repo.Close()

		app := testApp(repo)
		app.startup(ctx)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodPost, "/api/recommendations/invalid-uuid/approve", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestAPIHandler_GetPositions(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodGet, "/api/positions", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestAPIHandler_GetTrades(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodGet, "/api/trades", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("with limit parameter", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodGet, "/api/trades?limit=25", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestAPIHandler_GetAgentRuns(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodGet, "/api/agents/runs", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestAPIHandler_NotFound(t *testing.T) {
	app := testApp(nil)
	handler := testHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestAPIHandler_MethodNotAllowed(t *testing.T) {
	app := testApp(nil)
	handler := testHandler(app)

	req := httptest.NewRequest(http.MethodDelete, "/api/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestAPIHandler_ParseLimitParam(t *testing.T) {
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
			app := testApp(nil)
			handler := testHandler(app)

			url := "/api/test"
			if tt.queryParam != "" {
				url += "?" + tt.queryParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			result := handler.parseLimitParam(req, tt.defaultLimit)

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestApp_AnalyzeStock_RateLimiting(t *testing.T) {
	ctx := context.Background()
	app := testApp(nil)
	app.startup(ctx)

	requestCount := 5
	var results []error

	for i := 0; i < requestCount; i++ {
		_, err := app.AnalyzeStock("AAPL")
		results = append(results, err)
	}

	rateLimitedCount := 0
	managerNotInitCount := 0
	for _, err := range results {
		if err != nil {
			if strings.Contains(err.Error(), "queue full") {
				rateLimitedCount++
			} else if strings.Contains(err.Error(), "not initialized") {
				managerNotInitCount++
			}
		}
	}

	if managerNotInitCount != requestCount {
		t.Errorf("expected all %d requests to fail with manager not initialized, got %d", requestCount, managerNotInitCount)
	}
}

func TestApp_GetRecommendations(t *testing.T) {
	t.Run("repository not initialized", func(t *testing.T) {
		app := testApp(nil)
		_, err := app.GetRecommendations(10)
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})
}

func TestApp_GetPositions(t *testing.T) {
	t.Run("repository not initialized", func(t *testing.T) {
		app := testApp(nil)
		_, err := app.GetPositions()
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})
}

func TestApp_GetTrades(t *testing.T) {
	t.Run("repository not initialized", func(t *testing.T) {
		app := testApp(nil)
		_, err := app.GetTrades(10)
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})
}

func TestParseUUID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "valid UUID",
			input:     "550e8400-e29b-41d4-a716-446655440000",
			wantError: false,
		},
		{
			name:      "invalid UUID format",
			input:     "invalid-uuid",
			wantError: true,
		},
		{
			name:      "empty string",
			input:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseUUID(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("parseUUID() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestRecommendationActions(t *testing.T) {
	app := testApp(nil)

	t.Run("approve with nil repository", func(t *testing.T) {
		err := app.ApproveRecommendation("550e8400-e29b-41d4-a716-446655440000")
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})

	t.Run("reject with nil repository", func(t *testing.T) {
		err := app.RejectRecommendation("550e8400-e29b-41d4-a716-446655440000")
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})

	t.Run("approve with invalid UUID", func(t *testing.T) {
		err := app.ApproveRecommendation("invalid")
		if err == nil {
			t.Error("expected error with invalid UUID")
		}
	})
}

func TestAPIHandler_ValidateSymbol(t *testing.T) {
	app := testApp(nil)
	handler := testHandler(app)

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
			err := handler.validateSymbol(tt.symbol)
			if (err != nil) != tt.wantError {
				t.Errorf("validateSymbol(%s) error = %v, wantError %v", tt.symbol, err, tt.wantError)
			}
		})
	}
}

func TestAPIHandler_AnalyzeStock_InvalidSymbol(t *testing.T) {
	app := testApp(nil)
	handler := testHandler(app)

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

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status 400, got %d", w.Code)
			}
		})
	}
}


func TestApp_GetPendingRecommendations(t *testing.T) {
	t.Run("repository not initialized", func(t *testing.T) {
		app := testApp(nil)
		_, err := app.GetPendingRecommendations()
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})
}

func TestApp_GetAgentRuns(t *testing.T) {
	t.Run("repository not initialized", func(t *testing.T) {
		app := testApp(nil)
		_, err := app.GetAgentRuns(10)
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})
}

func TestAPIHandler_GetPendingRecommendations(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodGet, "/api/recommendations/pending", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestAPIHandler_RejectRecommendation(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodPost, "/api/recommendations/550e8400-e29b-41d4-a716-446655440000/reject", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("invalid UUID", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodPost, "/api/recommendations/invalid-uuid/reject", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}


func TestAPIHandler_GetPortfolio(t *testing.T) {
	t.Run("database not initialized", func(t *testing.T) {
		app := testApp(nil)
		handler := testHandler(app)

		req := httptest.NewRequest(http.MethodGet, "/api/portfolio", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestApp_Shutdown(t *testing.T) {
	ctx := context.Background()

	t.Run("with repository", func(t *testing.T) {
		connString := "postgres://trademachine:trademachine_dev@localhost:5432/trademachine?sslmode=disable"
		repo, err := repository.NewRepository(ctx, connString)
		if err != nil {
			t.Skip("database not available")
		}

		app := testApp(repo)
		app.shutdown(ctx) // Should close repository without error
	})

	t.Run("without repository", func(t *testing.T) {
		app := testApp(nil)
		app.shutdown(ctx) // Should not panic
	})
}

func TestApp_RejectRecommendation_InvalidUUID(t *testing.T) {
	app := testApp(nil)
	err := app.RejectRecommendation("not-a-uuid")
	if err == nil {
		t.Error("expected error with invalid UUID")
	}
}

// Integration tests with database
func TestIntegration_WithDatabase(t *testing.T) {
	ctx := context.Background()
	connString := "postgres://trademachine:trademachine_dev@localhost:5432/trademachine?sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available")
	}
	defer repo.Close()

	app := testApp(repo)
	app.startup(ctx)
	handler := testHandler(app)

	t.Run("get recommendations with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/recommendations?limit=5", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get pending recommendations with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/recommendations/pending", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get positions with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/positions", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get portfolio with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/portfolio", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get trades with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/trades?limit=10", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("get agent runs with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/agents/runs?limit=10", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("health check with database", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

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

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("reject nonexistent recommendation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/recommendations/550e8400-e29b-41d4-a716-446655440000/reject", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

func TestAPIHandler_MethodsNotAllowed(t *testing.T) {
	app := testApp(nil)
	handler := testHandler(app)

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

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status 405, got %d", w.Code)
			}
		})
	}
}

func TestAPIHandler_CORSHeaders(t *testing.T) {
	app := testApp(nil)
	handler := testHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("missing CORS Allow-Origin header")
	}
}

func TestAPIHandler_OptionsRequest(t *testing.T) {
	app := testApp(nil)
	handler := testHandler(app)

	req := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for OPTIONS, got %d", w.Code)
	}
}

func TestApp_NewApp_WithConcurrencyLimit(t *testing.T) {
	cfg := config.NewTestConfig()
	cfg.Agent.ConcurrencyLimit = 5
	app := NewApp(cfg, nil, nil, nil)

	if cap(app.analysisSem) != 5 {
		t.Errorf("expected concurrency limit 5, got %d", cap(app.analysisSem))
	}
}

func TestAPIHandler_GetRecommendations_WithStatus(t *testing.T) {
	ctx := context.Background()
	connString := "postgres://trademachine:trademachine_dev@localhost:5432/trademachine?sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available")
	}
	defer repo.Close()

	app := testApp(repo)
	app.startup(ctx)
	handler := testHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/recommendations?status=pending&limit=10", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAPIHandler_AnalyzeStock_InvalidJSON(t *testing.T) {
	app := testApp(nil)
	handler := testHandler(app)

	req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAPIHandler_GetAgentRuns_WithType(t *testing.T) {
	ctx := context.Background()
	connString := "postgres://trademachine:trademachine_dev@localhost:5432/trademachine?sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available")
	}
	defer repo.Close()

	app := testApp(repo)
	app.startup(ctx)
	handler := testHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/agents/runs?type=fundamental&limit=10", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestApp_RejectRecommendation_WithDatabase(t *testing.T) {
	ctx := context.Background()
	connString := "postgres://trademachine:trademachine_dev@localhost:5432/trademachine?sslmode=disable"
	repo, err := repository.NewRepository(ctx, connString)
	if err != nil {
		t.Skip("database not available")
	}
	defer repo.Close()

	app := testApp(repo)
	app.startup(ctx)

	err = app.RejectRecommendation("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Logf("reject recommendation error (expected for nonexistent ID): %v", err)
	}
}
