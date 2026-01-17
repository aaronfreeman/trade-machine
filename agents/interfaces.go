package agents

import (
	"context"
	"time"

	"trade-machine/models"

	marketdata "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/shopspring/decimal"
)

type BedrockServiceInterface interface {
	InvokeWithPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type AlphaVantageServiceInterface interface {
	GetFundamentals(ctx context.Context, symbol string) (*models.Fundamentals, error)
	GetNews(ctx context.Context, symbol string) ([]models.NewsArticle, error)
	GetQuote(ctx context.Context, symbol string) (*models.Quote, error)
}

type NewsAPIServiceInterface interface {
	GetNews(ctx context.Context, query string, limit int) ([]models.NewsArticle, error)
	GetHeadlines(ctx context.Context, query string, limit int) ([]models.NewsArticle, error)
}

type AlpacaServiceInterface interface {
	GetBars(ctx context.Context, symbol string, start, end time.Time, timeframe marketdata.TimeFrame) ([]marketdata.Bar, error)
	GetDailyBars(ctx context.Context, symbol string, days int) ([]marketdata.Bar, error)
	GetQuote(ctx context.Context, symbol string) (*models.Quote, error)
	GetLatestTrade(ctx context.Context, symbol string) (*models.Quote, error)
	GetAccount(ctx context.Context) (map[string]interface{}, error)
	PlaceOrder(ctx context.Context, symbol string, qty decimal.Decimal, side models.TradeSide, orderType string) (string, error)
	GetPositions(ctx context.Context) ([]models.Position, error)
	GetPosition(ctx context.Context, symbol string) (*models.Position, error)
}
