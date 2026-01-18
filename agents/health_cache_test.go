package agents

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"trade-machine/models"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/shopspring/decimal"
)

func TestNewHealthCache(t *testing.T) {
	ttl := 30 * time.Second
	cache := NewHealthCache(ttl)

	if cache == nil {
		t.Fatal("NewHealthCache should not return nil")
	}
	if cache.TTL() != ttl {
		t.Errorf("TTL() = %v, want %v", cache.TTL(), ttl)
	}
}

func TestHealthCache_InitialState(t *testing.T) {
	cache := NewHealthCache(30 * time.Second)

	// Initial state should be invalid (no cached value)
	if cache.IsValid() {
		t.Error("New cache should not be valid initially")
	}

	available, valid := cache.Get()
	if valid {
		t.Error("New cache should return valid=false")
	}
	if available {
		t.Error("New cache should return available=false")
	}
}

func TestHealthCache_SetAndGet(t *testing.T) {
	cache := NewHealthCache(30 * time.Second)

	// Set available=true
	cache.Set(true)
	available, valid := cache.Get()
	if !valid {
		t.Error("Cache should be valid after Set")
	}
	if !available {
		t.Error("Cache should return available=true")
	}

	// Set available=false
	cache.Set(false)
	available, valid = cache.Get()
	if !valid {
		t.Error("Cache should still be valid after Set")
	}
	if available {
		t.Error("Cache should return available=false after setting to false")
	}
}

func TestHealthCache_TTLExpiration(t *testing.T) {
	// Use a very short TTL
	cache := NewHealthCache(10 * time.Millisecond)

	cache.Set(true)

	// Should be valid immediately
	available, valid := cache.Get()
	if !valid {
		t.Error("Cache should be valid immediately after Set")
	}
	if !available {
		t.Error("Cache should return true")
	}

	// Wait for TTL to expire
	time.Sleep(15 * time.Millisecond)

	// Should be invalid after TTL
	available, valid = cache.Get()
	if valid {
		t.Error("Cache should be invalid after TTL expires")
	}
}

func TestHealthCache_Invalidate(t *testing.T) {
	cache := NewHealthCache(30 * time.Second)

	cache.Set(true)

	// Verify it's cached
	if !cache.IsValid() {
		t.Error("Cache should be valid after Set")
	}

	// Invalidate
	cache.Invalidate()

	// Should no longer be valid
	if cache.IsValid() {
		t.Error("Cache should be invalid after Invalidate")
	}

	_, valid := cache.Get()
	if valid {
		t.Error("Cache Get should return valid=false after Invalidate")
	}
}

func TestHealthCache_Concurrency(t *testing.T) {
	cache := NewHealthCache(30 * time.Second)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writers
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(val bool) {
			defer wg.Done()
			cache.Set(val)
		}(i%2 == 0)
	}

	// Concurrent readers
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Get()
			cache.IsValid()
		}()
	}

	// Concurrent invalidations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Invalidate()
		}()
	}

	wg.Wait()
	// If no race conditions, test passes
}

func TestHealthCache_ZeroTTL(t *testing.T) {
	// Zero TTL should effectively disable caching
	cache := NewHealthCache(0)

	cache.Set(true)

	// With zero TTL, cache is immediately invalid
	_, valid := cache.Get()
	if valid {
		t.Error("Zero TTL cache should never be valid")
	}
}

func TestHealthCache_IsValid(t *testing.T) {
	cache := NewHealthCache(50 * time.Millisecond)

	if cache.IsValid() {
		t.Error("New cache should not be valid")
	}

	cache.Set(false)
	if !cache.IsValid() {
		t.Error("Cache should be valid after Set")
	}

	time.Sleep(60 * time.Millisecond)
	if cache.IsValid() {
		t.Error("Cache should be invalid after TTL expires")
	}
}

func TestDefaultHealthCacheTTL(t *testing.T) {
	if DefaultHealthCacheTTL != 30*time.Second {
		t.Errorf("DefaultHealthCacheTTL = %v, want 30s", DefaultHealthCacheTTL)
	}
}

// Integration tests for agents using health cache

func TestFundamentalAnalyst_HealthCache_Integration(t *testing.T) {
	callCount := 0
	mockAlphaVantage := &mockAlphaVantageServiceWithCounter{
		fundamentals: &models.Fundamentals{Symbol: "AAPL", PERatio: 25.5},
		callCount:    &callCount,
	}

	// Use short TTL for testing
	analyst := NewFundamentalAnalystWithCacheTTL(nil, mockAlphaVantage, 50*time.Millisecond)
	ctx := context.Background()

	// First call should hit the API
	analyst.IsAvailable(ctx)
	if callCount != 1 {
		t.Errorf("First call should make 1 API call, got %d", callCount)
	}

	// Second call within TTL should use cache
	analyst.IsAvailable(ctx)
	if callCount != 1 {
		t.Errorf("Second call within TTL should not make API call, got %d calls", callCount)
	}

	// Wait for TTL to expire
	time.Sleep(60 * time.Millisecond)

	// Third call should hit the API again
	analyst.IsAvailable(ctx)
	if callCount != 2 {
		t.Errorf("Call after TTL should make API call, got %d calls", callCount)
	}
}

func TestFundamentalAnalyst_InvalidateHealthCache(t *testing.T) {
	callCount := 0
	mockAlphaVantage := &mockAlphaVantageServiceWithCounter{
		fundamentals: &models.Fundamentals{Symbol: "AAPL", PERatio: 25.5},
		callCount:    &callCount,
	}

	analyst := NewFundamentalAnalystWithCacheTTL(nil, mockAlphaVantage, 30*time.Second)
	ctx := context.Background()

	// First call
	analyst.IsAvailable(ctx)
	if callCount != 1 {
		t.Errorf("First call should make 1 API call, got %d", callCount)
	}

	// Invalidate cache
	analyst.InvalidateHealthCache()

	// Next call should hit API again
	analyst.IsAvailable(ctx)
	if callCount != 2 {
		t.Errorf("Call after invalidation should make API call, got %d calls", callCount)
	}
}

func TestFundamentalAnalyst_HealthCache_CachesFailure(t *testing.T) {
	callCount := 0
	mockAlphaVantage := &mockAlphaVantageServiceWithCounter{
		err:       errors.New("service unavailable"),
		callCount: &callCount,
	}

	analyst := NewFundamentalAnalystWithCacheTTL(nil, mockAlphaVantage, 50*time.Millisecond)
	ctx := context.Background()

	// First call should hit the API and get failure
	available := analyst.IsAvailable(ctx)
	if available {
		t.Error("Should return false when service is unavailable")
	}
	if callCount != 1 {
		t.Errorf("First call should make 1 API call, got %d", callCount)
	}

	// Second call should use cached failure
	available = analyst.IsAvailable(ctx)
	if available {
		t.Error("Cached result should still be false")
	}
	if callCount != 1 {
		t.Errorf("Second call should use cache, got %d calls", callCount)
	}
}

func TestTechnicalAnalyst_HealthCache_Integration(t *testing.T) {
	callCount := 0
	mockAlpaca := &mockAlpacaServiceWithCounter{
		callCount: &callCount,
	}

	cfg := testConfig()
	analyst := NewTechnicalAnalystWithCacheTTL(nil, mockAlpaca, cfg, 50*time.Millisecond)
	ctx := context.Background()

	// First call should hit the API
	analyst.IsAvailable(ctx)
	if callCount != 1 {
		t.Errorf("First call should make 1 API call, got %d", callCount)
	}

	// Second call within TTL should use cache
	analyst.IsAvailable(ctx)
	if callCount != 1 {
		t.Errorf("Second call within TTL should not make API call, got %d calls", callCount)
	}

	// Invalidate and check
	analyst.InvalidateHealthCache()
	analyst.IsAvailable(ctx)
	if callCount != 2 {
		t.Errorf("Call after invalidation should make API call, got %d calls", callCount)
	}
}

func TestNewsAnalyst_HealthCache_Integration(t *testing.T) {
	callCount := 0
	mockNewsAPI := &mockNewsAPIServiceWithCounter{
		callCount: &callCount,
	}

	analyst := NewNewsAnalystWithCacheTTL(nil, mockNewsAPI, 50*time.Millisecond)
	ctx := context.Background()

	// First call should hit the API
	analyst.IsAvailable(ctx)
	if callCount != 1 {
		t.Errorf("First call should make 1 API call, got %d", callCount)
	}

	// Second call within TTL should use cache
	analyst.IsAvailable(ctx)
	if callCount != 1 {
		t.Errorf("Second call within TTL should not make API call, got %d calls", callCount)
	}

	// Invalidate and check
	analyst.InvalidateHealthCache()
	analyst.IsAvailable(ctx)
	if callCount != 2 {
		t.Errorf("Call after invalidation should make API call, got %d calls", callCount)
	}
}

// Mock services with call counting

type mockAlphaVantageServiceWithCounter struct {
	fundamentals *models.Fundamentals
	err          error
	callCount    *int
}

func (m *mockAlphaVantageServiceWithCounter) GetFundamentals(ctx context.Context, symbol string) (*models.Fundamentals, error) {
	*m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	return m.fundamentals, nil
}

func (m *mockAlphaVantageServiceWithCounter) GetNews(ctx context.Context, symbol string) ([]models.NewsArticle, error) {
	return nil, nil
}

func (m *mockAlphaVantageServiceWithCounter) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	return nil, nil
}

type mockAlpacaServiceWithCounter struct {
	callCount *int
	err       error
}

func (m *mockAlpacaServiceWithCounter) GetBars(ctx context.Context, symbol string, start, end time.Time, timeframe marketdata.TimeFrame) ([]marketdata.Bar, error) {
	*m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	return []marketdata.Bar{{Close: 100.0}}, nil
}

func (m *mockAlpacaServiceWithCounter) GetDailyBars(ctx context.Context, symbol string, days int) ([]marketdata.Bar, error) {
	return nil, nil
}

func (m *mockAlpacaServiceWithCounter) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	return nil, nil
}

func (m *mockAlpacaServiceWithCounter) GetLatestTrade(ctx context.Context, symbol string) (*models.Quote, error) {
	return nil, nil
}

func (m *mockAlpacaServiceWithCounter) GetAccount(ctx context.Context) (*models.Account, error) {
	return &models.Account{
		ID:             "mock-account",
		Currency:       "USD",
		BuyingPower:    decimal.NewFromInt(100000),
		Cash:           decimal.NewFromInt(50000),
		PortfolioValue: decimal.NewFromInt(100000),
		Equity:         decimal.NewFromInt(100000),
	}, nil
}

func (m *mockAlpacaServiceWithCounter) GetPositions(ctx context.Context) ([]models.Position, error) {
	return nil, nil
}

func (m *mockAlpacaServiceWithCounter) GetPosition(ctx context.Context, symbol string) (*models.Position, error) {
	return nil, nil
}

func (m *mockAlpacaServiceWithCounter) PlaceOrder(ctx context.Context, symbol string, qty decimal.Decimal, side models.TradeSide, orderType string) (string, error) {
	return "", nil
}

type mockNewsAPIServiceWithCounter struct {
	callCount *int
	err       error
}

func (m *mockNewsAPIServiceWithCounter) GetNews(ctx context.Context, query string, limit int) ([]models.NewsArticle, error) {
	*m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	return []models.NewsArticle{{Title: "Test"}}, nil
}

func (m *mockNewsAPIServiceWithCounter) GetHeadlines(ctx context.Context, query string, limit int) ([]models.NewsArticle, error) {
	return nil, nil
}
