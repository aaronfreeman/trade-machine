//go:build e2e
// +build e2e

package scenarios

import (
	"encoding/json"
	"net/http"
	"testing"

	"trade-machine/e2e"
	"trade-machine/models"
)

// Note: Tests that require a fully initialized portfolio manager (analysis workflow)
// are skipped until we have full service mocking. These tests verify API endpoints
// and database operations work correctly.

func TestAnalysisWorkflow_PortfolioManagerNotInitialized(t *testing.T) {
	e2e.RequireDockerCompose(t)

	harness := e2e.NewTestHarness(t)
	if err := harness.Setup(); err != nil {
		t.Fatalf("failed to setup test harness: %v", err)
	}
	defer harness.Teardown()

	t.Run("analyze returns error when portfolio manager not initialized", func(t *testing.T) {
		// Without a portfolio manager, analysis should return 500
		resp := harness.DoRequest(http.MethodPost, "/api/analyze", `{"symbol":"AAPL"}`)

		if resp.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500 when portfolio manager not initialized, got %d", resp.Code)
		}

		var errResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		if errResp["error"] != "portfolio manager not initialized" {
			t.Errorf("unexpected error message: %s", errResp["error"])
		}
	})
}

func TestAnalysisWorkflow_InvalidSymbol(t *testing.T) {
	e2e.RequireDockerCompose(t)

	harness := e2e.NewTestHarness(t)
	if err := harness.Setup(); err != nil {
		t.Fatalf("failed to setup test harness: %v", err)
	}
	defer harness.Teardown()

	tests := []struct {
		name   string
		body   string
		status int
	}{
		{"empty symbol", `{"symbol":""}`, http.StatusBadRequest},
		{"missing symbol", `{}`, http.StatusBadRequest},
		{"invalid characters", `{"symbol":"AAPL!"}`, http.StatusBadRequest},
		{"too long", `{"symbol":"ABCDEFGHIJK"}`, http.StatusBadRequest},
		{"invalid JSON", `{invalid}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := harness.DoRequest(http.MethodPost, "/api/analyze", tt.body)

			if resp.Code != tt.status {
				t.Errorf("expected status %d, got %d: %s", tt.status, resp.Code, resp.Body.String())
			}
		})
	}
}

func TestAnalysisWorkflow_GetRecommendations(t *testing.T) {
	e2e.RequireDockerCompose(t)

	harness := e2e.NewTestHarness(t)
	if err := harness.Setup(); err != nil {
		t.Fatalf("failed to setup test harness: %v", err)
	}
	defer harness.Teardown()

	// Note: Full analysis workflow tests require a fully initialized portfolio manager
	// with mocked external services. These tests verify the recommendations API endpoints
	// work correctly with empty results (no recommendations in database).

	t.Run("list recommendations returns empty when none exist", func(t *testing.T) {
		resp := harness.DoRequest(http.MethodGet, "/api/recommendations?limit=10", "")

		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
		}

		var recommendations []models.Recommendation
		if err := json.NewDecoder(resp.Body).Decode(&recommendations); err != nil {
			t.Fatalf("failed to decode recommendations: %v", err)
		}

		if len(recommendations) != 0 {
			t.Errorf("expected 0 recommendations in fresh database, got %d", len(recommendations))
		}
	})

	t.Run("get pending recommendations returns empty when none exist", func(t *testing.T) {
		resp := harness.DoRequest(http.MethodGet, "/api/recommendations/pending", "")

		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
		}

		var pending []models.Recommendation
		if err := json.NewDecoder(resp.Body).Decode(&pending); err != nil {
			t.Fatalf("failed to decode pending recommendations: %v", err)
		}

		if len(pending) != 0 {
			t.Errorf("expected 0 pending recommendations in fresh database, got %d", len(pending))
		}
	})
}
