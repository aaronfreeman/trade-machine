package app

import (
	"context"
	"testing"

	"trade-machine/config"
	"trade-machine/models"
	"trade-machine/repository"
	"trade-machine/services"

	"github.com/google/uuid"
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
		connString := "postgres://postgres:postgres@localhost:5432/trademachine_test?sslmode=disable"
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
	connString := "postgres://postgres:postgres@localhost:5432/trademachine_test?sslmode=disable"
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

func TestApp_SetScreener(t *testing.T) {
	a := testApp(nil)

	// Initially no screener
	if a.Screener() != nil {
		t.Error("expected screener to be nil initially")
	}

	// Set screener
	mockScreener := &mockScreener{}
	a.SetScreener(mockScreener)

	if a.Screener() == nil {
		t.Error("expected screener to be set")
	}
}

func TestApp_RunScreener_NotInitialized(t *testing.T) {
	ctx := context.Background()
	a := testApp(nil)
	a.Startup(ctx)

	_, err := a.RunScreener()
	if err == nil {
		t.Error("expected error when screener is nil")
	}
	if err.Error() != "screener not initialized" {
		t.Errorf("expected 'screener not initialized' error, got: %v", err)
	}
}

func TestApp_GetLatestScreenerRun_NotInitialized(t *testing.T) {
	ctx := context.Background()
	a := testApp(nil)
	a.Startup(ctx)

	_, err := a.GetLatestScreenerRun()
	if err == nil {
		t.Error("expected error when screener is nil")
	}
}

func TestApp_GetScreenerRunHistory_NotInitialized(t *testing.T) {
	ctx := context.Background()
	a := testApp(nil)
	a.Startup(ctx)

	_, err := a.GetScreenerRunHistory(10)
	if err == nil {
		t.Error("expected error when screener is nil")
	}
}

func TestApp_GetScreenerRun_NotInitialized(t *testing.T) {
	ctx := context.Background()
	a := testApp(nil)
	a.Startup(ctx)

	_, err := a.GetScreenerRun("550e8400-e29b-41d4-a716-446655440000")
	if err == nil {
		t.Error("expected error when screener is nil")
	}
}

func TestApp_GetScreenerRun_InvalidUUID(t *testing.T) {
	ctx := context.Background()
	a := testApp(nil)
	a.Startup(ctx)

	mockScreener := &mockScreener{}
	a.SetScreener(mockScreener)

	_, err := a.GetScreenerRun("invalid-uuid")
	if err == nil {
		t.Error("expected error with invalid UUID")
	}
}

func TestApp_GetTopPicks_NotInitialized(t *testing.T) {
	ctx := context.Background()
	a := testApp(nil)
	a.Startup(ctx)

	_, err := a.GetTopPicks()
	if err == nil {
		t.Error("expected error when screener is nil")
	}
}

// mockScreener implements ScreenerInterface for testing
type mockScreener struct {
	runScreenCalled      bool
	getLatestRunCalled   bool
	getRunHistoryCalled  bool
	getRunCalled         bool
	getLatestPicksCalled bool
}

func (m *mockScreener) RunScreen(ctx context.Context) (*models.ScreenerRun, error) {
	m.runScreenCalled = true
	return &models.ScreenerRun{}, nil
}

func (m *mockScreener) GetLatestPicks(ctx context.Context) ([]models.ScreenerCandidate, error) {
	m.getLatestPicksCalled = true
	return nil, nil
}

func (m *mockScreener) GetLatestRun(ctx context.Context) (*models.ScreenerRun, error) {
	m.getLatestRunCalled = true
	return nil, nil
}

func (m *mockScreener) GetRunHistory(ctx context.Context, limit int) ([]models.ScreenerRun, error) {
	m.getRunHistoryCalled = true
	return nil, nil
}

func (m *mockScreener) GetRun(ctx context.Context, id uuid.UUID) (*models.ScreenerRun, error) {
	m.getRunCalled = true
	return nil, nil
}

// mockPortfolioManager implements PortfolioManagerInterface for testing
type mockPortfolioManager struct{}

func (m *mockPortfolioManager) AnalyzeSymbol(ctx context.Context, symbol string) (*models.Recommendation, error) {
	return nil, nil
}

// mockScreenerRepo implements ScreenerRepositoryInterface for testing
type mockScreenerRepo struct{}

func (m *mockScreenerRepo) CreateScreenerRun(ctx context.Context, run *models.ScreenerRun) error {
	return nil
}

func (m *mockScreenerRepo) UpdateScreenerRun(ctx context.Context, run *models.ScreenerRun) error {
	return nil
}

func (m *mockScreenerRepo) GetScreenerRun(ctx context.Context, id uuid.UUID) (*models.ScreenerRun, error) {
	return nil, nil
}

func (m *mockScreenerRepo) GetLatestScreenerRun(ctx context.Context) (*models.ScreenerRun, error) {
	return nil, nil
}

func (m *mockScreenerRepo) GetScreenerRunHistory(ctx context.Context, limit int) ([]models.ScreenerRun, error) {
	return nil, nil
}

func (m *mockScreenerRepo) CreateRecommendation(ctx context.Context, rec *models.Recommendation) error {
	return nil
}

func TestApp_SetScreenerFactory(t *testing.T) {
	cfg := testConfig()
	a := New(cfg, nil, &mockPortfolioManager{}, nil)

	// Initially no factory or repo
	if a.screenerFactory != nil {
		t.Error("expected screenerFactory to be nil initially")
	}
	if a.screenerRepo != nil {
		t.Error("expected screenerRepo to be nil initially")
	}

	// Set factory and repo
	mockRepo := &mockScreenerRepo{}
	factory := func(fmpService services.FMPServiceInterface, pm PortfolioManagerInterface, r ScreenerRepositoryInterface, cfg *config.ScreenerConfig) ScreenerInterface {
		return &mockScreener{}
	}
	a.SetScreenerFactory(factory, mockRepo)

	if a.screenerFactory == nil {
		t.Error("expected screenerFactory to be set")
	}
	if a.screenerRepo == nil {
		t.Error("expected screenerRepo to be set")
	}
}

func TestApp_InitializeScreenerWithFMPKey(t *testing.T) {
	t.Run("empty API key", func(t *testing.T) {
		a := testApp(nil)
		err := a.InitializeScreenerWithFMPKey("")
		if err == nil {
			t.Error("expected error with empty API key")
		}
		if err.Error() != "FMP API key is required" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("no factory configured", func(t *testing.T) {
		a := testApp(nil)
		err := a.InitializeScreenerWithFMPKey("test-api-key")
		if err == nil {
			t.Error("expected error when factory not configured")
		}
		if err.Error() != "screener factory not configured" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("no repository configured", func(t *testing.T) {
		cfg := testConfig()
		a := New(cfg, nil, &mockPortfolioManager{}, nil)
		factory := func(fmpService services.FMPServiceInterface, pm PortfolioManagerInterface, r ScreenerRepositoryInterface, cfg *config.ScreenerConfig) ScreenerInterface {
			return &mockScreener{}
		}
		a.screenerFactory = factory
		// Don't set screenerRepo

		err := a.InitializeScreenerWithFMPKey("test-api-key")
		if err == nil {
			t.Error("expected error when repo not configured")
		}
		if err.Error() != "screener repository not available" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("no portfolio manager", func(t *testing.T) {
		cfg := testConfig()
		a := New(cfg, nil, nil, nil) // no portfolio manager
		factory := func(fmpService services.FMPServiceInterface, pm PortfolioManagerInterface, r ScreenerRepositoryInterface, cfg *config.ScreenerConfig) ScreenerInterface {
			return &mockScreener{}
		}
		a.screenerFactory = factory
		a.screenerRepo = &mockScreenerRepo{}

		err := a.InitializeScreenerWithFMPKey("test-api-key")
		if err == nil {
			t.Error("expected error when portfolio manager not available")
		}
		if err.Error() != "portfolio manager not available" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		cfg := testConfig()
		a := New(cfg, nil, &mockPortfolioManager{}, nil)

		factoryCalled := false
		factory := func(fmpService services.FMPServiceInterface, pm PortfolioManagerInterface, r ScreenerRepositoryInterface, cfg *config.ScreenerConfig) ScreenerInterface {
			factoryCalled = true
			return &mockScreener{}
		}
		a.SetScreenerFactory(factory, &mockScreenerRepo{})

		// Initially no screener
		if a.Screener() != nil {
			t.Error("expected screener to be nil before initialization")
		}

		err := a.InitializeScreenerWithFMPKey("test-api-key")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !factoryCalled {
			t.Error("expected factory to be called")
		}

		if a.Screener() == nil {
			t.Error("expected screener to be set after initialization")
		}
	})
}

func TestApp_ScreenerStatus(t *testing.T) {
	t.Run("no dependencies", func(t *testing.T) {
		cfg := testConfig()
		a := New(cfg, nil, nil, nil)

		status := a.ScreenerStatus()

		if status.Available {
			t.Error("expected Available to be false")
		}
		if status.HasDatabase {
			t.Error("expected HasDatabase to be false")
		}
		if status.HasPortfolio {
			t.Error("expected HasPortfolio to be false")
		}
		if len(status.MissingServices) != 3 {
			t.Errorf("expected 3 missing services, got %d: %v", len(status.MissingServices), status.MissingServices)
		}
	})

	t.Run("with portfolio manager only", func(t *testing.T) {
		cfg := testConfig()
		a := New(cfg, nil, &mockPortfolioManager{}, nil)

		status := a.ScreenerStatus()

		if status.Available {
			t.Error("expected Available to be false")
		}
		if status.HasDatabase {
			t.Error("expected HasDatabase to be false")
		}
		if !status.HasPortfolio {
			t.Error("expected HasPortfolio to be true")
		}
	})

	t.Run("with screener factory configured", func(t *testing.T) {
		cfg := testConfig()
		a := New(cfg, nil, &mockPortfolioManager{}, nil)

		factory := func(fmpService services.FMPServiceInterface, pm PortfolioManagerInterface, r ScreenerRepositoryInterface, cfg *config.ScreenerConfig) ScreenerInterface {
			return &mockScreener{}
		}
		a.SetScreenerFactory(factory, &mockScreenerRepo{})

		status := a.ScreenerStatus()

		if status.Available {
			t.Error("expected Available to be false (screener not yet created)")
		}
		if !status.HasFMPKey {
			t.Error("expected HasFMPKey to be true when factory is set")
		}
	})

	t.Run("with screener available", func(t *testing.T) {
		cfg := testConfig()
		a := New(cfg, nil, &mockPortfolioManager{}, nil)
		a.SetScreener(&mockScreener{})

		status := a.ScreenerStatus()

		if !status.Available {
			t.Error("expected Available to be true")
		}
		if !status.HasFMPKey {
			t.Error("expected HasFMPKey to be true when screener is available")
		}
	})
}

func TestApp_SetUseMockServices(t *testing.T) {
	t.Run("mock services prevents reinitialization", func(t *testing.T) {
		cfg := testConfig()
		a := New(cfg, nil, &mockPortfolioManager{}, nil)

		// Set up factory and mock screener
		factoryCalled := false
		factory := func(fmpService services.FMPServiceInterface, pm PortfolioManagerInterface, r ScreenerRepositoryInterface, cfg *config.ScreenerConfig) ScreenerInterface {
			factoryCalled = true
			return &mockScreener{}
		}
		a.SetScreenerFactory(factory, &mockScreenerRepo{})
		a.SetScreener(&mockScreener{})

		// Enable mock services
		a.SetUseMockServices(true)

		// Try to reinitialize - should be skipped
		err := a.InitializeScreenerWithFMPKey("test-api-key")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Factory should NOT have been called
		if factoryCalled {
			t.Error("expected factory to NOT be called when mock services enabled")
		}
	})

	t.Run("without mock services allows reinitialization", func(t *testing.T) {
		cfg := testConfig()
		a := New(cfg, nil, &mockPortfolioManager{}, nil)

		factoryCalled := false
		factory := func(fmpService services.FMPServiceInterface, pm PortfolioManagerInterface, r ScreenerRepositoryInterface, cfg *config.ScreenerConfig) ScreenerInterface {
			factoryCalled = true
			return &mockScreener{}
		}
		a.SetScreenerFactory(factory, &mockScreenerRepo{})
		a.SetScreener(&mockScreener{})

		// Mock services NOT enabled (default)

		err := a.InitializeScreenerWithFMPKey("test-api-key")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Factory SHOULD have been called
		if !factoryCalled {
			t.Error("expected factory to be called when mock services not enabled")
		}
	})
}
