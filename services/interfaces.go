package services

import (
	"context"
	"time"

	"trade-machine/models"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/shopspring/decimal"
)

// BedrockService defines the interface for AI/LLM operations via AWS Bedrock
type BedrockServiceInterface interface {
	InvokeWithPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	InvokeStructured(ctx context.Context, systemPrompt, userPrompt string, result interface{}) error
	Chat(ctx context.Context, systemPrompt string, messages []ClaudeMessage) (string, error)
}

// AlphaVantageServiceInterface defines the interface for fundamental data operations
type AlphaVantageServiceInterface interface {
	GetFundamentals(ctx context.Context, symbol string) (*models.Fundamentals, error)
	GetNews(ctx context.Context, symbol string) ([]models.NewsArticle, error)
	GetQuote(ctx context.Context, symbol string) (*models.Quote, error)
}

// NewsAPIServiceInterface defines the interface for news data operations
type NewsAPIServiceInterface interface {
	GetNews(ctx context.Context, query string, limit int) ([]models.NewsArticle, error)
	GetHeadlines(ctx context.Context, query string, limit int) ([]models.NewsArticle, error)
}

// FMPServiceInterface defines the interface for Financial Modeling Prep operations
type FMPServiceInterface interface {
	// Screen searches for stocks matching the given criteria
	Screen(ctx context.Context, criteria ScreenCriteria) ([]ScreenerResult, error)
	// GetCompanyProfile returns enriched company profile data
	GetCompanyProfile(ctx context.Context, symbol string) (*CompanyProfile, error)
}

// ScreenCriteria defines filtering criteria for stock screening
type ScreenCriteria struct {
	MarketCapMin     int64   // Minimum market cap (e.g., 1_000_000_000 for $1B)
	MarketCapMax     int64   // Maximum market cap (0 = no limit)
	PERatioMax       float64 // Maximum P/E ratio (e.g., 15)
	PBRatioMax       float64 // Maximum P/B ratio (e.g., 1.5)
	EPSMin           float64 // Minimum EPS (e.g., 0 for positive earnings)
	DividendYieldMin float64 // Minimum dividend yield (optional)
	Sector           string  // Sector filter (optional)
	Limit            int     // Maximum results to return
}

// ScreenerResult represents a single stock from screener results
type ScreenerResult struct {
	Symbol        string  `json:"symbol"`
	CompanyName   string  `json:"companyName"`
	MarketCap     int64   `json:"marketCap"`
	Sector        string  `json:"sector"`
	Industry      string  `json:"industry"`
	Price         float64 `json:"price"`
	PERatio       float64 `json:"peRatio"`
	PBRatio       float64 `json:"pbRatio"`
	EPS           float64 `json:"eps"`
	DividendYield float64 `json:"dividendYield"`
	Beta          float64 `json:"beta"`
	Volume        int64   `json:"volume"`
	Exchange      string  `json:"exchange"`
	Country       string  `json:"country"`
}

// CompanyProfile represents enriched company profile data from FMP
type CompanyProfile struct {
	Symbol            string  `json:"symbol"`
	CompanyName       string  `json:"companyName"`
	Price             float64 `json:"price"`
	MarketCap         int64   `json:"mktCap"`
	Sector            string  `json:"sector"`
	Industry          string  `json:"industry"`
	Description       string  `json:"description"`
	CEO               string  `json:"ceo"`
	Website           string  `json:"website"`
	Exchange          string  `json:"exchange"`
	Country           string  `json:"country"`
	Beta              float64 `json:"beta"`
	VolAvg            int64   `json:"volAvg"`
	LastDividend      float64 `json:"lastDiv"`
	Range52Week       string  `json:"range"`
	Changes           float64 `json:"changes"`
	DCF               float64 `json:"dcf"`
	IPODate           string  `json:"ipoDate"`
	IsActivelyTrading bool    `json:"isActivelyTrading"`
}

// AlpacaServiceInterface defines the interface for trading and market data operations
type AlpacaServiceInterface interface {
	// Market data operations
	GetBars(ctx context.Context, symbol string, start, end time.Time, timeframe marketdata.TimeFrame) ([]marketdata.Bar, error)
	GetDailyBars(ctx context.Context, symbol string, days int) ([]marketdata.Bar, error)
	GetQuote(ctx context.Context, symbol string) (*models.Quote, error)
	GetLatestTrade(ctx context.Context, symbol string) (*models.Quote, error)

	// Account operations
	GetAccount(ctx context.Context) (*models.Account, error)

	// Trading operations
	PlaceOrder(ctx context.Context, symbol string, qty decimal.Decimal, side models.TradeSide, orderType string) (string, error)

	// Position operations
	GetPositions(ctx context.Context) ([]models.Position, error)
	GetPosition(ctx context.Context, symbol string) (*models.Position, error)
}

// Compile-time interface verification
var _ BedrockServiceInterface = (*BedrockService)(nil)
var _ AlphaVantageServiceInterface = (*AlphaVantageService)(nil)
var _ NewsAPIServiceInterface = (*NewsAPIService)(nil)
var _ AlpacaServiceInterface = (*AlpacaService)(nil)
