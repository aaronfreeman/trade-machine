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

// AlpacaServiceInterface defines the interface for trading and market data operations
type AlpacaServiceInterface interface {
	// Market data operations
	GetBars(ctx context.Context, symbol string, start, end time.Time, timeframe marketdata.TimeFrame) ([]marketdata.Bar, error)
	GetDailyBars(ctx context.Context, symbol string, days int) ([]marketdata.Bar, error)
	GetQuote(ctx context.Context, symbol string) (*models.Quote, error)
	GetLatestTrade(ctx context.Context, symbol string) (*models.Quote, error)

	// Account operations
	GetAccount(ctx context.Context) (map[string]interface{}, error)

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
