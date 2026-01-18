package screener

import (
	"context"
	"errors"
	"testing"
	"time"

	"trade-machine/config"
	"trade-machine/models"
	"trade-machine/services"

	"github.com/google/uuid"
)

// MockFMPService implements FMPServiceInterface for testing
type MockFMPService struct {
	ScreenFunc          func(ctx context.Context, criteria services.ScreenCriteria) ([]services.ScreenerResult, error)
	GetCompanyProfileFunc func(ctx context.Context, symbol string) (*services.CompanyProfile, error)
}

func (m *MockFMPService) Screen(ctx context.Context, criteria services.ScreenCriteria) ([]services.ScreenerResult, error) {
	if m.ScreenFunc != nil {
		return m.ScreenFunc(ctx, criteria)
	}
	return nil, nil
}

func (m *MockFMPService) GetCompanyProfile(ctx context.Context, symbol string) (*services.CompanyProfile, error) {
	if m.GetCompanyProfileFunc != nil {
		return m.GetCompanyProfileFunc(ctx, symbol)
	}
	return nil, nil
}

// MockAnalysisProvider implements AnalysisProvider for testing
type MockAnalysisProvider struct {
	AnalyzeSymbolFunc func(ctx context.Context, symbol string) (*models.Recommendation, error)
}

func (m *MockAnalysisProvider) AnalyzeSymbol(ctx context.Context, symbol string) (*models.Recommendation, error) {
	if m.AnalyzeSymbolFunc != nil {
		return m.AnalyzeSymbolFunc(ctx, symbol)
	}
	return nil, nil
}

// MockScreenerRepository implements ScreenerRepository for testing
type MockScreenerRepository struct {
	CreateScreenerRunFunc    func(ctx context.Context, run *models.ScreenerRun) error
	UpdateScreenerRunFunc    func(ctx context.Context, run *models.ScreenerRun) error
	GetScreenerRunFunc       func(ctx context.Context, id uuid.UUID) (*models.ScreenerRun, error)
	GetLatestScreenerRunFunc func(ctx context.Context) (*models.ScreenerRun, error)
	GetScreenerRunHistoryFunc func(ctx context.Context, limit int) ([]models.ScreenerRun, error)
	CreateRecommendationFunc func(ctx context.Context, rec *models.Recommendation) error
}

func (m *MockScreenerRepository) CreateScreenerRun(ctx context.Context, run *models.ScreenerRun) error {
	if m.CreateScreenerRunFunc != nil {
		return m.CreateScreenerRunFunc(ctx, run)
	}
	return nil
}

func (m *MockScreenerRepository) UpdateScreenerRun(ctx context.Context, run *models.ScreenerRun) error {
	if m.UpdateScreenerRunFunc != nil {
		return m.UpdateScreenerRunFunc(ctx, run)
	}
	return nil
}

func (m *MockScreenerRepository) GetScreenerRun(ctx context.Context, id uuid.UUID) (*models.ScreenerRun, error) {
	if m.GetScreenerRunFunc != nil {
		return m.GetScreenerRunFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockScreenerRepository) GetLatestScreenerRun(ctx context.Context) (*models.ScreenerRun, error) {
	if m.GetLatestScreenerRunFunc != nil {
		return m.GetLatestScreenerRunFunc(ctx)
	}
	return nil, nil
}

func (m *MockScreenerRepository) GetScreenerRunHistory(ctx context.Context, limit int) ([]models.ScreenerRun, error) {
	if m.GetScreenerRunHistoryFunc != nil {
		return m.GetScreenerRunHistoryFunc(ctx, limit)
	}
	return nil, nil
}

func (m *MockScreenerRepository) CreateRecommendation(ctx context.Context, rec *models.Recommendation) error {
	if m.CreateRecommendationFunc != nil {
		return m.CreateRecommendationFunc(ctx, rec)
	}
	return nil
}

func TestNewValueScreener(t *testing.T) {
	fmp := &MockFMPService{}
	analysis := &MockAnalysisProvider{}
	repo := &MockScreenerRepository{}
	cfg := &config.ScreenerConfig{
		MarketCapMin:       1_000_000_000,
		PERatioMax:         15.0,
		PBRatioMax:         1.5,
		PreFilterLimit:     15,
		TopPicksCount:      3,
		AnalysisTimeoutSec: 120,
		MaxConcurrent:      5,
	}

	screener := NewValueScreener(fmp, analysis, repo, cfg)

	if screener == nil {
		t.Error("NewValueScreener should not return nil")
	}
	if screener.fmpService != fmp {
		t.Error("fmpService not set correctly")
	}
	if screener.analysisProvider != analysis {
		t.Error("analysisProvider not set correctly")
	}
	if screener.repo != repo {
		t.Error("repo not set correctly")
	}
	if screener.cfg != cfg {
		t.Error("cfg not set correctly")
	}
}

func TestValueScreener_RunScreen_Success(t *testing.T) {
	fmp := &MockFMPService{
		ScreenFunc: func(ctx context.Context, criteria services.ScreenCriteria) ([]services.ScreenerResult, error) {
			return []services.ScreenerResult{
				{Symbol: "JNJ", CompanyName: "Johnson & Johnson", PERatio: 10, PBRatio: 1.0, DividendYield: 3.0},
				{Symbol: "PG", CompanyName: "Procter & Gamble", PERatio: 12, PBRatio: 1.2, DividendYield: 2.5},
				{Symbol: "KO", CompanyName: "Coca-Cola", PERatio: 14, PBRatio: 1.4, DividendYield: 2.0},
			}, nil
		},
	}

	analysis := &MockAnalysisProvider{
		AnalyzeSymbolFunc: func(ctx context.Context, symbol string) (*models.Recommendation, error) {
			rec := models.NewRecommendation(symbol, models.RecommendationActionBuy, "Good value stock")
			rec.FundamentalScore = 70
			rec.SentimentScore = 65
			rec.TechnicalScore = 60
			rec.Confidence = 80
			return rec, nil
		},
	}

	var createdRun *models.ScreenerRun
	var updatedRun *models.ScreenerRun

	repo := &MockScreenerRepository{
		CreateScreenerRunFunc: func(ctx context.Context, run *models.ScreenerRun) error {
			createdRun = run
			return nil
		},
		UpdateScreenerRunFunc: func(ctx context.Context, run *models.ScreenerRun) error {
			updatedRun = run
			return nil
		},
		CreateRecommendationFunc: func(ctx context.Context, rec *models.Recommendation) error {
			return nil
		},
	}

	cfg := &config.ScreenerConfig{
		MarketCapMin:       1_000_000_000,
		PERatioMax:         15.0,
		PBRatioMax:         1.5,
		PreFilterLimit:     15,
		TopPicksCount:      3,
		AnalysisTimeoutSec: 120,
		MaxConcurrent:      5,
	}

	screener := NewValueScreener(fmp, analysis, repo, cfg)
	ctx := context.Background()

	run, err := screener.RunScreen(ctx)

	if err != nil {
		t.Fatalf("RunScreen failed: %v", err)
	}
	if run == nil {
		t.Fatal("RunScreen returned nil run")
	}
	if createdRun == nil {
		t.Error("CreateScreenerRun was not called")
	}
	if updatedRun == nil {
		t.Error("UpdateScreenerRun was not called")
	}
	if !run.IsCompleted() {
		t.Errorf("Run status should be completed, got %s", run.Status)
	}
	if len(run.Candidates) != 3 {
		t.Errorf("Should have 3 candidates, got %d", len(run.Candidates))
	}
	if run.DurationMs < 0 {
		t.Error("DurationMs should not be negative")
	}
}

func TestValueScreener_RunScreen_FMPError(t *testing.T) {
	fmp := &MockFMPService{
		ScreenFunc: func(ctx context.Context, criteria services.ScreenCriteria) ([]services.ScreenerResult, error) {
			return nil, errors.New("API rate limit exceeded")
		},
	}

	var updatedRun *models.ScreenerRun
	repo := &MockScreenerRepository{
		CreateScreenerRunFunc: func(ctx context.Context, run *models.ScreenerRun) error {
			return nil
		},
		UpdateScreenerRunFunc: func(ctx context.Context, run *models.ScreenerRun) error {
			updatedRun = run
			return nil
		},
	}

	cfg := &config.ScreenerConfig{
		PreFilterLimit:     15,
		TopPicksCount:      3,
		AnalysisTimeoutSec: 120,
		MaxConcurrent:      5,
	}

	screener := NewValueScreener(fmp, &MockAnalysisProvider{}, repo, cfg)
	ctx := context.Background()

	run, err := screener.RunScreen(ctx)

	if err == nil {
		t.Error("RunScreen should return error when FMP fails")
	}
	if run == nil {
		t.Fatal("RunScreen should return run even on error")
	}
	if !run.IsFailed() {
		t.Errorf("Run should be failed, got %s", run.Status)
	}
	if run.Error == "" {
		t.Error("Run error should be set")
	}
	if updatedRun == nil {
		t.Error("UpdateScreenerRun should be called even on failure")
	}
}

func TestValueScreener_RunScreen_CreateRunError(t *testing.T) {
	repo := &MockScreenerRepository{
		CreateScreenerRunFunc: func(ctx context.Context, run *models.ScreenerRun) error {
			return errors.New("database error")
		},
	}

	cfg := &config.ScreenerConfig{
		PreFilterLimit:     15,
		TopPicksCount:      3,
		AnalysisTimeoutSec: 120,
		MaxConcurrent:      5,
	}

	screener := NewValueScreener(&MockFMPService{}, &MockAnalysisProvider{}, repo, cfg)
	ctx := context.Background()

	_, err := screener.RunScreen(ctx)

	if err == nil {
		t.Error("RunScreen should return error when CreateScreenerRun fails")
	}
}

func TestValueScreener_RunScreen_PartialAnalysisFailure(t *testing.T) {
	fmp := &MockFMPService{
		ScreenFunc: func(ctx context.Context, criteria services.ScreenCriteria) ([]services.ScreenerResult, error) {
			return []services.ScreenerResult{
				{Symbol: "GOOD", PERatio: 10},
				{Symbol: "FAIL", PERatio: 12},
			}, nil
		},
	}

	analysis := &MockAnalysisProvider{
		AnalyzeSymbolFunc: func(ctx context.Context, symbol string) (*models.Recommendation, error) {
			if symbol == "FAIL" {
				return nil, errors.New("analysis failed")
			}
			rec := models.NewRecommendation(symbol, models.RecommendationActionBuy, "Good stock")
			rec.FundamentalScore = 70
			rec.Confidence = 80
			return rec, nil
		},
	}

	repo := &MockScreenerRepository{
		CreateScreenerRunFunc:    func(ctx context.Context, run *models.ScreenerRun) error { return nil },
		UpdateScreenerRunFunc:    func(ctx context.Context, run *models.ScreenerRun) error { return nil },
		CreateRecommendationFunc: func(ctx context.Context, rec *models.Recommendation) error { return nil },
	}

	cfg := &config.ScreenerConfig{
		PreFilterLimit:     15,
		TopPicksCount:      3,
		AnalysisTimeoutSec: 120,
		MaxConcurrent:      5,
	}

	screener := NewValueScreener(fmp, analysis, repo, cfg)
	ctx := context.Background()

	run, err := screener.RunScreen(ctx)

	if err != nil {
		t.Fatalf("RunScreen should succeed with partial failures: %v", err)
	}
	if !run.IsCompleted() {
		t.Error("Run should still complete with partial failures")
	}

	// Count analyzed candidates
	analyzedCount := 0
	for _, c := range run.Candidates {
		if c.Analyzed {
			analyzedCount++
		}
	}
	if analyzedCount != 1 {
		t.Errorf("Should have 1 analyzed candidate, got %d", analyzedCount)
	}
}

func TestValueScreener_GetLatestPicks(t *testing.T) {
	score1, conf1 := 80.0, 90.0
	score2, conf2 := 70.0, 85.0

	repo := &MockScreenerRepository{
		GetLatestScreenerRunFunc: func(ctx context.Context) (*models.ScreenerRun, error) {
			run := &models.ScreenerRun{
				ID:     uuid.New(),
				Status: models.ScreenerRunStatusCompleted,
				Candidates: []models.ScreenerCandidate{
					{Symbol: "JNJ", Score: &score1, Confidence: &conf1, Analyzed: true},
					{Symbol: "PG", Score: &score2, Confidence: &conf2, Analyzed: true},
				},
			}
			return run, nil
		},
	}

	cfg := &config.ScreenerConfig{
		TopPicksCount: 3,
	}

	screener := NewValueScreener(&MockFMPService{}, &MockAnalysisProvider{}, repo, cfg)
	ctx := context.Background()

	picks, err := screener.GetLatestPicks(ctx)

	if err != nil {
		t.Fatalf("GetLatestPicks failed: %v", err)
	}
	if len(picks) != 2 {
		t.Errorf("Should return 2 picks, got %d", len(picks))
	}
	if picks[0].Symbol != "JNJ" {
		t.Errorf("First pick should be JNJ, got %s", picks[0].Symbol)
	}
}

func TestValueScreener_GetLatestPicks_NoRun(t *testing.T) {
	repo := &MockScreenerRepository{
		GetLatestScreenerRunFunc: func(ctx context.Context) (*models.ScreenerRun, error) {
			return nil, nil
		},
	}

	cfg := &config.ScreenerConfig{TopPicksCount: 3}
	screener := NewValueScreener(&MockFMPService{}, &MockAnalysisProvider{}, repo, cfg)
	ctx := context.Background()

	picks, err := screener.GetLatestPicks(ctx)

	if err != nil {
		t.Fatalf("GetLatestPicks failed: %v", err)
	}
	if picks != nil {
		t.Errorf("Should return nil for no run, got %v", picks)
	}
}

func TestValueScreener_GetLatestPicks_RunningRun(t *testing.T) {
	repo := &MockScreenerRepository{
		GetLatestScreenerRunFunc: func(ctx context.Context) (*models.ScreenerRun, error) {
			run := &models.ScreenerRun{
				ID:     uuid.New(),
				Status: models.ScreenerRunStatusRunning,
			}
			return run, nil
		},
	}

	cfg := &config.ScreenerConfig{TopPicksCount: 3}
	screener := NewValueScreener(&MockFMPService{}, &MockAnalysisProvider{}, repo, cfg)
	ctx := context.Background()

	picks, err := screener.GetLatestPicks(ctx)

	if err != nil {
		t.Fatalf("GetLatestPicks failed: %v", err)
	}
	if picks != nil {
		t.Errorf("Should return nil for running run, got %v", picks)
	}
}

func TestValueScreener_GetLatestRun(t *testing.T) {
	expectedRun := &models.ScreenerRun{
		ID:     uuid.New(),
		Status: models.ScreenerRunStatusCompleted,
	}

	repo := &MockScreenerRepository{
		GetLatestScreenerRunFunc: func(ctx context.Context) (*models.ScreenerRun, error) {
			return expectedRun, nil
		},
	}

	cfg := &config.ScreenerConfig{}
	screener := NewValueScreener(&MockFMPService{}, &MockAnalysisProvider{}, repo, cfg)
	ctx := context.Background()

	run, err := screener.GetLatestRun(ctx)

	if err != nil {
		t.Fatalf("GetLatestRun failed: %v", err)
	}
	if run != expectedRun {
		t.Error("GetLatestRun returned unexpected run")
	}
}

func TestValueScreener_GetRunHistory(t *testing.T) {
	expectedRuns := []models.ScreenerRun{
		{ID: uuid.New(), Status: models.ScreenerRunStatusCompleted},
		{ID: uuid.New(), Status: models.ScreenerRunStatusCompleted},
	}

	repo := &MockScreenerRepository{
		GetScreenerRunHistoryFunc: func(ctx context.Context, limit int) ([]models.ScreenerRun, error) {
			if limit != 10 {
				t.Errorf("Expected limit 10, got %d", limit)
			}
			return expectedRuns, nil
		},
	}

	cfg := &config.ScreenerConfig{}
	screener := NewValueScreener(&MockFMPService{}, &MockAnalysisProvider{}, repo, cfg)
	ctx := context.Background()

	runs, err := screener.GetRunHistory(ctx, 10)

	if err != nil {
		t.Fatalf("GetRunHistory failed: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("Should return 2 runs, got %d", len(runs))
	}
}

func TestValueScreener_GetRun(t *testing.T) {
	runID := uuid.New()
	expectedRun := &models.ScreenerRun{
		ID:     runID,
		Status: models.ScreenerRunStatusCompleted,
	}

	repo := &MockScreenerRepository{
		GetScreenerRunFunc: func(ctx context.Context, id uuid.UUID) (*models.ScreenerRun, error) {
			if id != runID {
				t.Errorf("Expected ID %v, got %v", runID, id)
			}
			return expectedRun, nil
		},
	}

	cfg := &config.ScreenerConfig{}
	screener := NewValueScreener(&MockFMPService{}, &MockAnalysisProvider{}, repo, cfg)
	ctx := context.Background()

	run, err := screener.GetRun(ctx, runID)

	if err != nil {
		t.Fatalf("GetRun failed: %v", err)
	}
	if run != expectedRun {
		t.Error("GetRun returned unexpected run")
	}
}

func TestValueScreener_RunScreen_Timeout(t *testing.T) {
	fmp := &MockFMPService{
		ScreenFunc: func(ctx context.Context, criteria services.ScreenCriteria) ([]services.ScreenerResult, error) {
			return []services.ScreenerResult{
				{Symbol: "SLOW"},
			}, nil
		},
	}

	analysis := &MockAnalysisProvider{
		AnalyzeSymbolFunc: func(ctx context.Context, symbol string) (*models.Recommendation, error) {
			// Simulate slow analysis
			time.Sleep(200 * time.Millisecond)
			return nil, ctx.Err()
		},
	}

	repo := &MockScreenerRepository{
		CreateScreenerRunFunc: func(ctx context.Context, run *models.ScreenerRun) error { return nil },
		UpdateScreenerRunFunc: func(ctx context.Context, run *models.ScreenerRun) error { return nil },
	}

	cfg := &config.ScreenerConfig{
		PreFilterLimit:     15,
		TopPicksCount:      3,
		AnalysisTimeoutSec: 1, // Very short timeout
		MaxConcurrent:      5,
	}

	screener := NewValueScreener(fmp, analysis, repo, cfg)
	ctx := context.Background()

	run, err := screener.RunScreen(ctx)

	// Should complete even with timeout (analysis failures are handled gracefully)
	if err != nil {
		t.Fatalf("RunScreen should not return error on analysis timeout: %v", err)
	}
	if run == nil {
		t.Fatal("Run should not be nil")
	}
}

func TestValueScreener_RunScreen_ConcurrencyLimit(t *testing.T) {
	var concurrentCalls int
	var maxConcurrent int
	var mu = make(chan struct{}, 1)
	mu <- struct{}{}

	fmp := &MockFMPService{
		ScreenFunc: func(ctx context.Context, criteria services.ScreenCriteria) ([]services.ScreenerResult, error) {
			return []services.ScreenerResult{
				{Symbol: "A"}, {Symbol: "B"}, {Symbol: "C"},
				{Symbol: "D"}, {Symbol: "E"}, {Symbol: "F"},
			}, nil
		},
	}

	analysis := &MockAnalysisProvider{
		AnalyzeSymbolFunc: func(ctx context.Context, symbol string) (*models.Recommendation, error) {
			<-mu
			concurrentCalls++
			if concurrentCalls > maxConcurrent {
				maxConcurrent = concurrentCalls
			}
			mu <- struct{}{}

			time.Sleep(10 * time.Millisecond)

			<-mu
			concurrentCalls--
			mu <- struct{}{}

			rec := models.NewRecommendation(symbol, models.RecommendationActionBuy, "Test")
			rec.FundamentalScore = 70
			rec.Confidence = 80
			return rec, nil
		},
	}

	repo := &MockScreenerRepository{
		CreateScreenerRunFunc:    func(ctx context.Context, run *models.ScreenerRun) error { return nil },
		UpdateScreenerRunFunc:    func(ctx context.Context, run *models.ScreenerRun) error { return nil },
		CreateRecommendationFunc: func(ctx context.Context, rec *models.Recommendation) error { return nil },
	}

	cfg := &config.ScreenerConfig{
		PreFilterLimit:     15,
		TopPicksCount:      3,
		AnalysisTimeoutSec: 120,
		MaxConcurrent:      2, // Only allow 2 concurrent
	}

	screener := NewValueScreener(fmp, analysis, repo, cfg)
	ctx := context.Background()

	_, err := screener.RunScreen(ctx)

	if err != nil {
		t.Fatalf("RunScreen failed: %v", err)
	}
	if maxConcurrent > 2 {
		t.Errorf("Max concurrent should be <= 2, got %d", maxConcurrent)
	}
}
