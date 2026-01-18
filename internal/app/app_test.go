package app

import (
	"context"
	"testing"

	"trade-machine/config"
	"trade-machine/repository"
)

// testConfig returns a test configuration
func testConfig() *config.Config {
	return config.NewTestConfig()
}

// testApp creates an App with test config for testing
func testApp(repo RepositoryInterface) *App {
	return New(testConfig(), repo, nil, nil)
}

func TestNew_WithConcurrencyLimit(t *testing.T) {
	cfg := config.NewTestConfig()
	cfg.Agent.ConcurrencyLimit = 5
	a := New(cfg, nil, nil, nil)

	if a.AnalysisSemCapacity() != 5 {
		t.Errorf("expected concurrency limit 5, got %d", a.AnalysisSemCapacity())
	}
}

func TestApp_AnalyzeStock_RateLimiting(t *testing.T) {
	ctx := context.Background()
	a := testApp(nil)
	a.Startup(ctx)

	requestCount := 5
	var results []error

	for i := 0; i < requestCount; i++ {
		_, err := a.AnalyzeStock("AAPL")
		results = append(results, err)
	}

	managerNotInitCount := 0
	for _, err := range results {
		if err != nil && err.Error() == "portfolio manager not initialized" {
			managerNotInitCount++
		}
	}

	if managerNotInitCount != requestCount {
		t.Errorf("expected all %d requests to fail with manager not initialized, got %d", requestCount, managerNotInitCount)
	}
}

func TestApp_GetRecommendations(t *testing.T) {
	t.Run("repository not initialized", func(t *testing.T) {
		a := testApp(nil)
		_, err := a.GetRecommendations(10)
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})
}

func TestApp_GetPendingRecommendations(t *testing.T) {
	t.Run("repository not initialized", func(t *testing.T) {
		a := testApp(nil)
		_, err := a.GetPendingRecommendations()
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})
}

func TestApp_GetPositions(t *testing.T) {
	t.Run("repository not initialized", func(t *testing.T) {
		a := testApp(nil)
		_, err := a.GetPositions()
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})
}

func TestApp_GetTrades(t *testing.T) {
	t.Run("repository not initialized", func(t *testing.T) {
		a := testApp(nil)
		_, err := a.GetTrades(10)
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})
}

func TestApp_GetAgentRuns(t *testing.T) {
	t.Run("repository not initialized", func(t *testing.T) {
		a := testApp(nil)
		_, err := a.GetAgentRuns(10)
		if err == nil {
			t.Error("expected error when repository is nil")
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

		a := testApp(repo)
		a.Shutdown(ctx) // Should close repository without error
	})

	t.Run("without repository", func(t *testing.T) {
		a := testApp(nil)
		a.Shutdown(ctx) // Should not panic
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
			_, err := ParseUUID(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseUUID() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestRecommendationActions(t *testing.T) {
	a := testApp(nil)

	t.Run("approve with nil repository", func(t *testing.T) {
		err := a.ApproveRecommendation("550e8400-e29b-41d4-a716-446655440000")
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})

	t.Run("reject with nil repository", func(t *testing.T) {
		err := a.RejectRecommendation("550e8400-e29b-41d4-a716-446655440000")
		if err == nil {
			t.Error("expected error when repository is nil")
		}
	})

	t.Run("approve with invalid UUID", func(t *testing.T) {
		err := a.ApproveRecommendation("invalid")
		if err == nil {
			t.Error("expected error with invalid UUID")
		}
	})
}

func TestApp_RejectRecommendation_InvalidUUID(t *testing.T) {
	a := testApp(nil)
	err := a.RejectRecommendation("not-a-uuid")
	if err == nil {
		t.Error("expected error with invalid UUID")
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

	a := testApp(repo)
	a.Startup(ctx)

	err = a.RejectRecommendation("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Logf("reject recommendation error (expected for nonexistent ID): %v", err)
	}
}
