package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"trade-machine/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// getTestDB returns a repository connected to the test database.
// If DATABASE_URL is not set, the test is skipped.
func getTestDB(t *testing.T) *Repository {
	t.Helper()

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo, err := NewRepository(ctx, connString)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	return repo
}

// cleanupPositions removes all test positions
func cleanupPositions(t *testing.T, repo *Repository) {
	t.Helper()
	ctx := context.Background()
	repo.pool.Exec(ctx, "DELETE FROM positions WHERE symbol LIKE 'TEST%'")
}

// cleanupTrades removes all test trades
func cleanupTrades(t *testing.T, repo *Repository) {
	t.Helper()
	ctx := context.Background()
	repo.pool.Exec(ctx, "DELETE FROM trades WHERE symbol LIKE 'TEST%'")
}

// cleanupRecommendations removes all test recommendations
func cleanupRecommendations(t *testing.T, repo *Repository) {
	t.Helper()
	ctx := context.Background()
	repo.pool.Exec(ctx, "DELETE FROM recommendations WHERE symbol LIKE 'TEST%'")
}

// cleanupAgentRuns removes all test agent runs
func cleanupAgentRuns(t *testing.T, repo *Repository) {
	t.Helper()
	ctx := context.Background()
	repo.pool.Exec(ctx, "DELETE FROM agent_runs WHERE symbol LIKE 'TEST%'")
}

// cleanupCache removes all test cache entries
func cleanupCache(t *testing.T, repo *Repository) {
	t.Helper()
	ctx := context.Background()
	repo.pool.Exec(ctx, "DELETE FROM market_data_cache WHERE symbol LIKE 'TEST%'")
}

// =============================================================================
// Position Tests
// =============================================================================

func TestRepository_Positions_CRUD(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupPositions(t, repo)

	ctx := context.Background()

	// Create a position
	pos := &models.Position{
		ID:            uuid.New(),
		Symbol:        "TEST001",
		Quantity:      decimal.NewFromInt(100),
		AvgEntryPrice: decimal.NewFromFloat(150.50),
		CurrentPrice:  decimal.NewFromFloat(155.00),
		UnrealizedPL:  decimal.NewFromFloat(450.00),
		Side:          models.PositionSideLong,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Test Create
	err := repo.CreatePosition(ctx, pos)
	if err != nil {
		t.Fatalf("CreatePosition failed: %v", err)
	}

	// Test GetPosition
	retrieved, err := repo.GetPosition(ctx, pos.ID)
	if err != nil {
		t.Fatalf("GetPosition failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetPosition returned nil")
	}
	if retrieved.Symbol != "TEST001" {
		t.Errorf("expected symbol TEST001, got %s", retrieved.Symbol)
	}
	if !retrieved.Quantity.Equal(decimal.NewFromInt(100)) {
		t.Errorf("expected quantity 100, got %s", retrieved.Quantity)
	}

	// Test GetPositionBySymbol
	bySymbol, err := repo.GetPositionBySymbol(ctx, "TEST001")
	if err != nil {
		t.Fatalf("GetPositionBySymbol failed: %v", err)
	}
	if bySymbol == nil {
		t.Fatal("GetPositionBySymbol returned nil")
	}
	if bySymbol.ID != pos.ID {
		t.Errorf("expected ID %s, got %s", pos.ID, bySymbol.ID)
	}

	// Test Update
	pos.Quantity = decimal.NewFromInt(150)
	pos.CurrentPrice = decimal.NewFromFloat(160.00)
	err = repo.UpdatePosition(ctx, pos)
	if err != nil {
		t.Fatalf("UpdatePosition failed: %v", err)
	}

	updated, err := repo.GetPosition(ctx, pos.ID)
	if err != nil {
		t.Fatalf("GetPosition after update failed: %v", err)
	}
	if !updated.Quantity.Equal(decimal.NewFromInt(150)) {
		t.Errorf("expected updated quantity 150, got %s", updated.Quantity)
	}

	// Test GetPositions
	positions, err := repo.GetPositions(ctx)
	if err != nil {
		t.Fatalf("GetPositions failed: %v", err)
	}
	found := false
	for _, p := range positions {
		if p.ID == pos.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("created position not found in GetPositions")
	}

	// Test Delete
	err = repo.DeletePosition(ctx, pos.ID)
	if err != nil {
		t.Fatalf("DeletePosition failed: %v", err)
	}

	deleted, err := repo.GetPosition(ctx, pos.ID)
	if err != nil {
		t.Fatalf("GetPosition after delete failed: %v", err)
	}
	if deleted != nil {
		t.Error("position should be nil after delete")
	}
}

func TestRepository_GetPosition_NotFound(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()

	ctx := context.Background()
	nonExistentID := uuid.New()

	pos, err := repo.GetPosition(ctx, nonExistentID)
	if err != nil {
		t.Fatalf("GetPosition should not error for non-existent ID: %v", err)
	}
	if pos != nil {
		t.Error("GetPosition should return nil for non-existent ID")
	}
}

func TestRepository_GetPositionBySymbol_NotFound(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	pos, err := repo.GetPositionBySymbol(ctx, "NONEXISTENT")
	if err != nil {
		t.Fatalf("GetPositionBySymbol should not error for non-existent symbol: %v", err)
	}
	if pos != nil {
		t.Error("GetPositionBySymbol should return nil for non-existent symbol")
	}
}

// =============================================================================
// Trade Tests
// =============================================================================

func TestRepository_Trades_CRUD(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupTrades(t, repo)

	ctx := context.Background()

	// Create a trade
	trade := models.NewTrade("TEST002", models.TradeSideBuy, decimal.NewFromInt(50), decimal.NewFromFloat(100.00))

	err := repo.CreateTrade(ctx, trade)
	if err != nil {
		t.Fatalf("CreateTrade failed: %v", err)
	}

	// Test GetTrade
	retrieved, err := repo.GetTrade(ctx, trade.ID)
	if err != nil {
		t.Fatalf("GetTrade failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetTrade returned nil")
	}
	if retrieved.Symbol != "TEST002" {
		t.Errorf("expected symbol TEST002, got %s", retrieved.Symbol)
	}
	if retrieved.Side != models.TradeSideBuy {
		t.Errorf("expected side buy, got %s", retrieved.Side)
	}

	// Test UpdateTradeStatus
	err = repo.UpdateTradeStatus(ctx, trade.ID, models.TradeStatusExecuted)
	if err != nil {
		t.Fatalf("UpdateTradeStatus failed: %v", err)
	}

	updated, err := repo.GetTrade(ctx, trade.ID)
	if err != nil {
		t.Fatalf("GetTrade after update failed: %v", err)
	}
	if updated.Status != models.TradeStatusExecuted {
		t.Errorf("expected status executed, got %s", updated.Status)
	}

	// Test GetTrades
	trades, err := repo.GetTrades(ctx, 10)
	if err != nil {
		t.Fatalf("GetTrades failed: %v", err)
	}
	found := false
	for _, tr := range trades {
		if tr.ID == trade.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("created trade not found in GetTrades")
	}
}

func TestRepository_GetTradesBySymbol(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupTrades(t, repo)

	ctx := context.Background()

	// Create multiple trades
	trade1 := models.NewTrade("TEST003", models.TradeSideBuy, decimal.NewFromInt(10), decimal.NewFromFloat(50.00))
	trade2 := models.NewTrade("TEST003", models.TradeSideSell, decimal.NewFromInt(5), decimal.NewFromFloat(55.00))
	trade3 := models.NewTrade("TEST004", models.TradeSideBuy, decimal.NewFromInt(20), decimal.NewFromFloat(30.00))

	repo.CreateTrade(ctx, trade1)
	repo.CreateTrade(ctx, trade2)
	repo.CreateTrade(ctx, trade3)

	// Get trades for TEST003
	trades, err := repo.GetTradesBySymbol(ctx, "TEST003", 10)
	if err != nil {
		t.Fatalf("GetTradesBySymbol failed: %v", err)
	}

	if len(trades) != 2 {
		t.Errorf("expected 2 trades for TEST003, got %d", len(trades))
	}

	for _, tr := range trades {
		if tr.Symbol != "TEST003" {
			t.Errorf("expected symbol TEST003, got %s", tr.Symbol)
		}
	}
}

func TestRepository_GetTrades_DefaultLimit(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	// Test with zero limit (should default to 50)
	_, err := repo.GetTrades(ctx, 0)
	if err != nil {
		t.Fatalf("GetTrades with zero limit failed: %v", err)
	}

	// Test with negative limit (should default to 50)
	_, err = repo.GetTrades(ctx, -1)
	if err != nil {
		t.Fatalf("GetTrades with negative limit failed: %v", err)
	}
}

// =============================================================================
// Recommendation Tests
// =============================================================================

func TestRepository_Recommendations_CRUD(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupRecommendations(t, repo)

	ctx := context.Background()

	// Create a recommendation
	rec := models.NewRecommendation("TEST005", models.RecommendationActionBuy, "Test recommendation reasoning")
	rec.Quantity = decimal.NewFromInt(25)
	rec.TargetPrice = decimal.NewFromFloat(200.00)
	rec.Confidence = 75.5
	rec.FundamentalScore = 80.0
	rec.SentimentScore = 70.0
	rec.TechnicalScore = 75.0

	err := repo.CreateRecommendation(ctx, rec)
	if err != nil {
		t.Fatalf("CreateRecommendation failed: %v", err)
	}

	// Test GetRecommendation
	retrieved, err := repo.GetRecommendation(ctx, rec.ID)
	if err != nil {
		t.Fatalf("GetRecommendation failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetRecommendation returned nil")
	}
	if retrieved.Symbol != "TEST005" {
		t.Errorf("expected symbol TEST005, got %s", retrieved.Symbol)
	}
	if retrieved.Action != models.RecommendationActionBuy {
		t.Errorf("expected action buy, got %s", retrieved.Action)
	}
	if retrieved.Confidence != 75.5 {
		t.Errorf("expected confidence 75.5, got %f", retrieved.Confidence)
	}

	// Test GetPendingRecommendations
	pending, err := repo.GetPendingRecommendations(ctx)
	if err != nil {
		t.Fatalf("GetPendingRecommendations failed: %v", err)
	}
	found := false
	for _, r := range pending {
		if r.ID == rec.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("created recommendation not found in GetPendingRecommendations")
	}

	// Test ApproveRecommendation
	err = repo.ApproveRecommendation(ctx, rec.ID)
	if err != nil {
		t.Fatalf("ApproveRecommendation failed: %v", err)
	}

	approved, err := repo.GetRecommendation(ctx, rec.ID)
	if err != nil {
		t.Fatalf("GetRecommendation after approve failed: %v", err)
	}
	if approved.Status != models.RecommendationStatusApproved {
		t.Errorf("expected status approved, got %s", approved.Status)
	}
	if approved.ApprovedAt == nil {
		t.Error("ApprovedAt should be set")
	}
}

func TestRepository_RejectRecommendation(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupRecommendations(t, repo)

	ctx := context.Background()

	rec := models.NewRecommendation("TEST006", models.RecommendationActionSell, "Test rejection")
	rec.Quantity = decimal.NewFromInt(10)
	rec.TargetPrice = decimal.NewFromFloat(50.00)
	rec.Confidence = 30.0

	repo.CreateRecommendation(ctx, rec)

	err := repo.RejectRecommendation(ctx, rec.ID)
	if err != nil {
		t.Fatalf("RejectRecommendation failed: %v", err)
	}

	rejected, _ := repo.GetRecommendation(ctx, rec.ID)
	if rejected.Status != models.RecommendationStatusRejected {
		t.Errorf("expected status rejected, got %s", rejected.Status)
	}
	if rejected.RejectedAt == nil {
		t.Error("RejectedAt should be set")
	}
}

func TestRepository_ExecuteRecommendation(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupRecommendations(t, repo)
	defer cleanupTrades(t, repo)

	ctx := context.Background()

	// Create recommendation
	rec := models.NewRecommendation("TEST007", models.RecommendationActionBuy, "Test execution")
	rec.Quantity = decimal.NewFromInt(15)
	rec.TargetPrice = decimal.NewFromFloat(75.00)
	rec.Confidence = 85.0
	repo.CreateRecommendation(ctx, rec)

	// Create trade
	trade := models.NewTrade("TEST007", models.TradeSideBuy, decimal.NewFromInt(15), decimal.NewFromFloat(75.00))
	repo.CreateTrade(ctx, trade)

	// Execute recommendation
	err := repo.ExecuteRecommendation(ctx, rec.ID, trade.ID)
	if err != nil {
		t.Fatalf("ExecuteRecommendation failed: %v", err)
	}

	executed, _ := repo.GetRecommendation(ctx, rec.ID)
	if executed.Status != models.RecommendationStatusExecuted {
		t.Errorf("expected status executed, got %s", executed.Status)
	}
	if executed.ExecutedTradeID == nil || *executed.ExecutedTradeID != trade.ID {
		t.Error("ExecutedTradeID should be set to trade ID")
	}
}

func TestRepository_GetRecommendations_FilterByStatus(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupRecommendations(t, repo)

	ctx := context.Background()

	// Create recommendations with different statuses
	pending := models.NewRecommendation("TEST008", models.RecommendationActionBuy, "Pending")
	pending.Quantity = decimal.NewFromInt(10)
	pending.TargetPrice = decimal.NewFromFloat(100.00)
	pending.Confidence = 50.0

	approved := models.NewRecommendation("TEST009", models.RecommendationActionSell, "Approved")
	approved.Quantity = decimal.NewFromInt(5)
	approved.TargetPrice = decimal.NewFromFloat(90.00)
	approved.Confidence = 60.0

	repo.CreateRecommendation(ctx, pending)
	repo.CreateRecommendation(ctx, approved)
	repo.ApproveRecommendation(ctx, approved.ID)

	// Get only pending
	pendingRecs, err := repo.GetRecommendations(ctx, models.RecommendationStatusPending, 50)
	if err != nil {
		t.Fatalf("GetRecommendations failed: %v", err)
	}

	for _, r := range pendingRecs {
		if r.Status != models.RecommendationStatusPending {
			t.Errorf("expected only pending recommendations, got %s", r.Status)
		}
	}

	// Get all (empty status)
	allRecs, err := repo.GetRecommendations(ctx, "", 50)
	if err != nil {
		t.Fatalf("GetRecommendations (all) failed: %v", err)
	}

	if len(allRecs) < 2 {
		t.Error("expected at least 2 recommendations when filtering by empty status")
	}
}

// =============================================================================
// Agent Run Tests
// =============================================================================

func TestRepository_AgentRuns_CRUD(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupAgentRuns(t, repo)

	ctx := context.Background()

	// Create an agent run
	run := models.NewAgentRun(models.AgentTypeFundamental, "TEST010")
	run.InputData = map[string]interface{}{
		"symbol": "TEST010",
		"period": "1y",
	}

	err := repo.CreateAgentRun(ctx, run)
	if err != nil {
		t.Fatalf("CreateAgentRun failed: %v", err)
	}

	// Test GetAgentRun
	retrieved, err := repo.GetAgentRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetAgentRun failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetAgentRun returned nil")
	}
	if retrieved.AgentType != models.AgentTypeFundamental {
		t.Errorf("expected agent type fundamental, got %s", retrieved.AgentType)
	}
	if retrieved.Symbol != "TEST010" {
		t.Errorf("expected symbol TEST010, got %s", retrieved.Symbol)
	}
	if retrieved.Status != models.AgentRunStatusRunning {
		t.Errorf("expected status running, got %s", retrieved.Status)
	}

	// Complete the run
	run.Complete(map[string]interface{}{
		"score":      75.5,
		"action":     "buy",
		"confidence": 80.0,
	})

	err = repo.UpdateAgentRun(ctx, run)
	if err != nil {
		t.Fatalf("UpdateAgentRun failed: %v", err)
	}

	updated, _ := repo.GetAgentRun(ctx, run.ID)
	if updated.Status != models.AgentRunStatusCompleted {
		t.Errorf("expected status completed, got %s", updated.Status)
	}
	if updated.OutputData == nil {
		t.Error("OutputData should be set after completion")
	}
	if updated.DurationMs <= 0 {
		t.Error("DurationMs should be positive after completion")
	}
}

func TestRepository_AgentRuns_Fail(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupAgentRuns(t, repo)

	ctx := context.Background()

	run := models.NewAgentRun(models.AgentTypeNews, "TEST011")
	if err := repo.CreateAgentRun(ctx, run); err != nil {
		t.Fatalf("CreateAgentRun failed: %v", err)
	}

	// Fail the run
	run.Fail(errors.New("API rate limit exceeded"))
	if err := repo.UpdateAgentRun(ctx, run); err != nil {
		t.Fatalf("UpdateAgentRun failed: %v", err)
	}

	failed, err := repo.GetAgentRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetAgentRun failed: %v", err)
	}
	if failed == nil {
		t.Fatal("expected agent run, got nil")
	}
	if failed.Status != models.AgentRunStatusFailed {
		t.Errorf("expected status failed, got %s", failed.Status)
	}
	if failed.ErrorMessage != "API rate limit exceeded" {
		t.Errorf("expected error message, got %s", failed.ErrorMessage)
	}
}

func TestRepository_GetAgentRuns_FilterByType(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupAgentRuns(t, repo)

	ctx := context.Background()

	// Create runs of different types
	fundamental := models.NewAgentRun(models.AgentTypeFundamental, "TEST012")
	news := models.NewAgentRun(models.AgentTypeNews, "TEST013")
	technical := models.NewAgentRun(models.AgentTypeTechnical, "TEST014")

	repo.CreateAgentRun(ctx, fundamental)
	repo.CreateAgentRun(ctx, news)
	repo.CreateAgentRun(ctx, technical)

	// Get only fundamental runs
	fundamentalRuns, err := repo.GetAgentRuns(ctx, models.AgentTypeFundamental, 50)
	if err != nil {
		t.Fatalf("GetAgentRuns failed: %v", err)
	}

	for _, r := range fundamentalRuns {
		if r.AgentType != models.AgentTypeFundamental {
			t.Errorf("expected only fundamental runs, got %s", r.AgentType)
		}
	}

	// Get all runs
	allRuns, err := repo.GetAgentRuns(ctx, "", 50)
	if err != nil {
		t.Fatalf("GetAgentRuns (all) failed: %v", err)
	}

	if len(allRuns) < 3 {
		t.Error("expected at least 3 runs when filtering by empty type")
	}
}

func TestRepository_GetRecentRunsForSymbol(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupAgentRuns(t, repo)

	ctx := context.Background()

	// Create multiple runs for same symbol
	run1 := models.NewAgentRun(models.AgentTypeFundamental, "TEST015")
	run2 := models.NewAgentRun(models.AgentTypeNews, "TEST015")
	run3 := models.NewAgentRun(models.AgentTypeTechnical, "TEST016") // Different symbol

	repo.CreateAgentRun(ctx, run1)
	repo.CreateAgentRun(ctx, run2)
	repo.CreateAgentRun(ctx, run3)

	runs, err := repo.GetRecentRunsForSymbol(ctx, "TEST015", 10)
	if err != nil {
		t.Fatalf("GetRecentRunsForSymbol failed: %v", err)
	}

	if len(runs) != 2 {
		t.Errorf("expected 2 runs for TEST015, got %d", len(runs))
	}

	for _, r := range runs {
		if r.Symbol != "TEST015" {
			t.Errorf("expected symbol TEST015, got %s", r.Symbol)
		}
	}
}

// =============================================================================
// Cache Tests
// =============================================================================

func TestRepository_Cache_SetAndGet(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupCache(t, repo)

	ctx := context.Background()

	data := map[string]interface{}{
		"price":  150.50,
		"volume": 1000000,
		"change": 2.5,
	}

	// Set cache
	err := repo.SetCachedData(ctx, "TEST017", "quote", data, 1*time.Hour)
	if err != nil {
		t.Fatalf("SetCachedData failed: %v", err)
	}

	// Get cache
	cached, err := repo.GetCachedData(ctx, "TEST017", "quote")
	if err != nil {
		t.Fatalf("GetCachedData failed: %v", err)
	}
	if cached == nil {
		t.Fatal("GetCachedData returned nil")
	}

	if cached["price"] != 150.50 {
		t.Errorf("expected price 150.50, got %v", cached["price"])
	}
}

func TestRepository_Cache_Expiration(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupCache(t, repo)

	ctx := context.Background()

	data := map[string]interface{}{"test": "data"}

	// Set cache with very short TTL
	err := repo.SetCachedData(ctx, "TEST018", "quote", data, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("SetCachedData failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Should return nil for expired data
	cached, err := repo.GetCachedData(ctx, "TEST018", "quote")
	if err != nil {
		t.Fatalf("GetCachedData failed: %v", err)
	}
	if cached != nil {
		t.Error("expected nil for expired cache")
	}
}

func TestRepository_Cache_Upsert(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupCache(t, repo)

	ctx := context.Background()

	// Set initial data
	data1 := map[string]interface{}{"price": 100.00}
	repo.SetCachedData(ctx, "TEST019", "quote", data1, 1*time.Hour)

	// Update with new data
	data2 := map[string]interface{}{"price": 105.00}
	err := repo.SetCachedData(ctx, "TEST019", "quote", data2, 1*time.Hour)
	if err != nil {
		t.Fatalf("SetCachedData (upsert) failed: %v", err)
	}

	// Should get updated data
	cached, _ := repo.GetCachedData(ctx, "TEST019", "quote")
	if cached["price"] != 105.00 {
		t.Errorf("expected updated price 105.00, got %v", cached["price"])
	}
}

func TestRepository_Cache_Invalidate(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupCache(t, repo)

	ctx := context.Background()

	data := map[string]interface{}{"test": "data"}
	repo.SetCachedData(ctx, "TEST020", "quote", data, 1*time.Hour)

	// Invalidate specific cache
	err := repo.InvalidateCache(ctx, "TEST020", "quote")
	if err != nil {
		t.Fatalf("InvalidateCache failed: %v", err)
	}

	cached, _ := repo.GetCachedData(ctx, "TEST020", "quote")
	if cached != nil {
		t.Error("expected nil after invalidation")
	}
}

func TestRepository_Cache_InvalidateAllForSymbol(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupCache(t, repo)

	ctx := context.Background()

	// Set multiple cache entries for same symbol
	repo.SetCachedData(ctx, "TEST021", "quote", map[string]interface{}{"type": "quote"}, 1*time.Hour)
	repo.SetCachedData(ctx, "TEST021", "fundamentals", map[string]interface{}{"type": "fundamentals"}, 1*time.Hour)
	repo.SetCachedData(ctx, "TEST021", "news", map[string]interface{}{"type": "news"}, 1*time.Hour)

	// Invalidate all for symbol
	err := repo.InvalidateAllCacheForSymbol(ctx, "TEST021")
	if err != nil {
		t.Fatalf("InvalidateAllCacheForSymbol failed: %v", err)
	}

	// All should be nil
	quote, _ := repo.GetCachedData(ctx, "TEST021", "quote")
	fundamentals, _ := repo.GetCachedData(ctx, "TEST021", "fundamentals")
	news, _ := repo.GetCachedData(ctx, "TEST021", "news")

	if quote != nil || fundamentals != nil || news != nil {
		t.Error("expected all cache entries to be invalidated")
	}
}

func TestRepository_Cache_CleanExpired(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()
	defer cleanupCache(t, repo)

	ctx := context.Background()

	// Set cache with very short TTL
	data := map[string]interface{}{"test": "expired"}
	repo.SetCachedData(ctx, "TEST022", "quote", data, 1*time.Millisecond)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Clean expired
	deleted, err := repo.CleanExpiredCache(ctx)
	if err != nil {
		t.Fatalf("CleanExpiredCache failed: %v", err)
	}

	if deleted < 1 {
		t.Error("expected at least 1 expired entry to be cleaned")
	}
}

func TestRepository_Cache_NotFound(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()

	ctx := context.Background()

	cached, err := repo.GetCachedData(ctx, "NONEXISTENT", "quote")
	if err != nil {
		t.Fatalf("GetCachedData should not error for non-existent: %v", err)
	}
	if cached != nil {
		t.Error("expected nil for non-existent cache")
	}
}

// =============================================================================
// Repository Connection Tests
// =============================================================================

func TestNewRepository_InvalidConnection(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := NewRepository(ctx, "postgres://invalid:invalid@localhost:9999/nonexistent")
	if err == nil {
		t.Error("expected error for invalid connection string")
	}
}

func TestRepository_Health(t *testing.T) {
	repo := getTestDB(t)
	defer repo.Close()

	ctx := context.Background()
	err := repo.Health(ctx)
	if err != nil {
		t.Errorf("Health() should return nil for valid connection: %v", err)
	}
}
