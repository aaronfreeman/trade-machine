package agents

import (
	"context"
	"time"

	"trade-machine/models"
	"trade-machine/services"

	marketdata "github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/shopspring/decimal"
)

type mockBedrockService struct {
	response string
	err      error
}

func (m *mockBedrockService) InvokeWithPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *mockBedrockService) InvokeStructured(ctx context.Context, systemPrompt, userPrompt string, result interface{}) error {
	return m.err
}

func (m *mockBedrockService) Chat(ctx context.Context, systemPrompt string, messages []services.ClaudeMessage) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

type mockAlphaVantageService struct {
	fundamentals *models.Fundamentals
	news         []models.NewsArticle
	quote        *models.Quote
	err          error
}

func (m *mockAlphaVantageService) GetFundamentals(ctx context.Context, symbol string) (*models.Fundamentals, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.fundamentals, nil
}

func (m *mockAlphaVantageService) GetNews(ctx context.Context, symbol string) ([]models.NewsArticle, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.news, nil
}

func (m *mockAlphaVantageService) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.quote, nil
}

type mockNewsAPIService struct {
	articles []models.NewsArticle
	err      error
}

func (m *mockNewsAPIService) GetNews(ctx context.Context, query string, limit int) ([]models.NewsArticle, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.articles, nil
}

func (m *mockNewsAPIService) GetHeadlines(ctx context.Context, query string, limit int) ([]models.NewsArticle, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.articles, nil
}

type mockAlpacaService struct {
	bars []marketdata.Bar
	err  error
}

func (m *mockAlpacaService) GetBars(ctx context.Context, symbol string, start, end time.Time, timeframe marketdata.TimeFrame) ([]marketdata.Bar, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.bars, nil
}

func (m *mockAlpacaService) GetDailyBars(ctx context.Context, symbol string, days int) ([]marketdata.Bar, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.bars, nil
}

func (m *mockAlpacaService) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	return nil, nil
}

func (m *mockAlpacaService) GetLatestTrade(ctx context.Context, symbol string) (*models.Quote, error) {
	return nil, nil
}

func (m *mockAlpacaService) GetAccount(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}

func (m *mockAlpacaService) PlaceOrder(ctx context.Context, symbol string, qty decimal.Decimal, side models.TradeSide, orderType string) (string, error) {
	return "", nil
}

func (m *mockAlpacaService) GetPositions(ctx context.Context) ([]models.Position, error) {
	return nil, nil
}

func (m *mockAlpacaService) GetPosition(ctx context.Context, symbol string) (*models.Position, error) {
	return nil, nil
}
