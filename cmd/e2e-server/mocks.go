package main

import (
	"context"
	"time"

	"trade-machine/models"
	"trade-machine/services"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/shopspring/decimal"
)

// MockFMPService provides mock FMP data for e2e testing
type MockFMPService struct{}

func NewMockFMPService() *MockFMPService {
	return &MockFMPService{}
}

func (m *MockFMPService) Screen(ctx context.Context, criteria services.ScreenCriteria) ([]services.ScreenerResult, error) {
	// Return mock screener results based on fixture data
	return []services.ScreenerResult{
		{
			Symbol:        "JNJ",
			CompanyName:   "Johnson & Johnson",
			MarketCap:     400000000000,
			Sector:        "Healthcare",
			Industry:      "Drug Manufacturers",
			Price:         155.50,
			PERatio:       15.2,
			PBRatio:       5.8,
			EPS:           10.15,
			DividendYield: 2.95,
			Beta:          0.55,
			Volume:        7500000,
			Exchange:      "NYSE",
			Country:       "US",
		},
		{
			Symbol:        "PG",
			CompanyName:   "Procter & Gamble",
			MarketCap:     350000000000,
			Sector:        "Consumer Defensive",
			Industry:      "Household Products",
			Price:         148.75,
			PERatio:       24.5,
			PBRatio:       7.2,
			EPS:           5.90,
			DividendYield: 2.45,
			Beta:          0.42,
			Volume:        6200000,
			Exchange:      "NYSE",
			Country:       "US",
		},
		{
			Symbol:        "KO",
			CompanyName:   "Coca-Cola Company",
			MarketCap:     260000000000,
			Sector:        "Consumer Defensive",
			Industry:      "Beverages",
			Price:         60.25,
			PERatio:       22.8,
			PBRatio:       10.5,
			EPS:           2.48,
			DividendYield: 3.1,
			Beta:          0.58,
			Volume:        12000000,
			Exchange:      "NYSE",
			Country:       "US",
		},
	}, nil
}

func (m *MockFMPService) GetCompanyProfile(ctx context.Context, symbol string) (*services.CompanyProfile, error) {
	profiles := map[string]*services.CompanyProfile{
		"JNJ": {
			Symbol:            "JNJ",
			CompanyName:       "Johnson & Johnson",
			Price:             155.50,
			MarketCap:         400000000000,
			Sector:            "Healthcare",
			Industry:          "Drug Manufacturers",
			Description:       "Johnson & Johnson is a global healthcare company.",
			CEO:               "Joaquin Duato",
			Website:           "https://www.jnj.com",
			Exchange:          "NYSE",
			Country:           "US",
			Beta:              0.55,
			VolAvg:            7500000,
			LastDividend:      4.76,
			IsActivelyTrading: true,
		},
		"PG": {
			Symbol:            "PG",
			CompanyName:       "Procter & Gamble",
			Price:             148.75,
			MarketCap:         350000000000,
			Sector:            "Consumer Defensive",
			Industry:          "Household Products",
			Description:       "Procter & Gamble is a consumer goods corporation.",
			CEO:               "Jon Moeller",
			Website:           "https://www.pg.com",
			Exchange:          "NYSE",
			Country:           "US",
			Beta:              0.42,
			VolAvg:            6200000,
			LastDividend:      3.76,
			IsActivelyTrading: true,
		},
		"KO": {
			Symbol:            "KO",
			CompanyName:       "Coca-Cola Company",
			Price:             60.25,
			MarketCap:         260000000000,
			Sector:            "Consumer Defensive",
			Industry:          "Beverages",
			Description:       "The Coca-Cola Company is a beverage company.",
			CEO:               "James Quincey",
			Website:           "https://www.coca-colacompany.com",
			Exchange:          "NYSE",
			Country:           "US",
			Beta:              0.58,
			VolAvg:            12000000,
			LastDividend:      1.84,
			IsActivelyTrading: true,
		},
	}

	if profile, ok := profiles[symbol]; ok {
		return profile, nil
	}

	// Return a default profile for unknown symbols
	return &services.CompanyProfile{
		Symbol:            symbol,
		CompanyName:       symbol + " Corp",
		Price:             100.00,
		MarketCap:         10000000000,
		Sector:            "Technology",
		Industry:          "Software",
		IsActivelyTrading: true,
	}, nil
}

// MockPortfolioManager provides mock analysis for e2e testing
type MockPortfolioManager struct {
	repo ScreenerRepoInterface
}

// ScreenerRepoInterface is a subset of repository operations needed by MockPortfolioManager
type ScreenerRepoInterface interface {
	CreateRecommendation(ctx context.Context, rec *models.Recommendation) error
}

func NewMockPortfolioManager(repo ScreenerRepoInterface) *MockPortfolioManager {
	return &MockPortfolioManager{repo: repo}
}

func (m *MockPortfolioManager) AnalyzeSymbol(ctx context.Context, symbol string) (*models.Recommendation, error) {
	// Create a mock recommendation with realistic scores
	scores := map[string]struct {
		fundamental float64
		technical   float64
		sentiment   float64
	}{
		"JNJ": {fundamental: 65.0, technical: 45.0, sentiment: 55.0},
		"PG":  {fundamental: 55.0, technical: 40.0, sentiment: 50.0},
		"KO":  {fundamental: 50.0, technical: 35.0, sentiment: 45.0},
	}

	score, ok := scores[symbol]
	if !ok {
		score = struct {
			fundamental float64
			technical   float64
			sentiment   float64
		}{50.0, 40.0, 45.0}
	}

	// Calculate weighted score (40% fundamental, 30% technical, 30% sentiment)
	weightedScore := score.fundamental*0.4 + score.technical*0.3 + score.sentiment*0.3

	action := models.RecommendationActionHold
	if weightedScore > 25 {
		action = models.RecommendationActionBuy
	} else if weightedScore < -25 {
		action = models.RecommendationActionSell
	}

	rec := models.NewRecommendation(symbol, action, "Mock analysis for e2e testing: Strong fundamentals with stable technicals.")
	rec.Quantity = decimal.NewFromInt(10)
	rec.Confidence = 65.0
	rec.FundamentalScore = score.fundamental
	rec.TechnicalScore = score.technical
	rec.SentimentScore = score.sentiment

	// Save to repository if available
	if m.repo != nil {
		if err := m.repo.CreateRecommendation(ctx, rec); err != nil {
			// Log but don't fail - the recommendation is still valid
		}
	}

	return rec, nil
}

// MockAlpacaService provides mock Alpaca data for e2e testing
type MockAlpacaService struct{}

func NewMockAlpacaService() *MockAlpacaService {
	return &MockAlpacaService{}
}

func (m *MockAlpacaService) GetBars(ctx context.Context, symbol string, start, end time.Time, timeframe marketdata.TimeFrame) ([]marketdata.Bar, error) {
	return []marketdata.Bar{}, nil
}

func (m *MockAlpacaService) GetDailyBars(ctx context.Context, symbol string, days int) ([]marketdata.Bar, error) {
	return []marketdata.Bar{}, nil
}

func (m *MockAlpacaService) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	return &models.Quote{
		Symbol: symbol,
		Last:   decimal.NewFromFloat(100.00),
		Bid:    decimal.NewFromFloat(99.95),
		Ask:    decimal.NewFromFloat(100.05),
	}, nil
}

func (m *MockAlpacaService) GetLatestTrade(ctx context.Context, symbol string) (*models.Quote, error) {
	return &models.Quote{
		Symbol: symbol,
		Last:   decimal.NewFromFloat(100.00),
	}, nil
}

func (m *MockAlpacaService) GetAccount(ctx context.Context) (*models.Account, error) {
	return &models.Account{
		ID:           "mock-account-id",
		Currency:     "USD",
		Cash:         decimal.NewFromFloat(100000.00),
		PortfolioValue: decimal.NewFromFloat(100000.00),
		BuyingPower:  decimal.NewFromFloat(100000.00),
	}, nil
}

func (m *MockAlpacaService) PlaceOrder(ctx context.Context, symbol string, qty decimal.Decimal, side models.TradeSide, orderType string) (string, error) {
	return "mock-order-id", nil
}

func (m *MockAlpacaService) GetPositions(ctx context.Context) ([]models.Position, error) {
	return []models.Position{}, nil
}

func (m *MockAlpacaService) GetPosition(ctx context.Context, symbol string) (*models.Position, error) {
	return nil, nil
}
