package services

import (
	"context"
	"fmt"
	"time"

	"trade-machine/models"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/shopspring/decimal"
)

// AlpacaService handles communication with Alpaca for trading and market data
type AlpacaService struct {
	tradeClient *alpaca.Client
	dataClient  *marketdata.Client
}

// NewAlpacaService creates a new AlpacaService instance
func NewAlpacaService(apiKey, apiSecret, baseURL string) *AlpacaService {
	tradeClient := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   baseURL,
	})

	dataClient := marketdata.NewClient(marketdata.ClientOpts{
		APIKey:    apiKey,
		APISecret: apiSecret,
	})

	return &AlpacaService{
		tradeClient: tradeClient,
		dataClient:  dataClient,
	}
}

// GetAccount returns the current account information
func (s *AlpacaService) GetAccount(ctx context.Context) (*alpaca.Account, error) {
	return s.tradeClient.GetAccount()
}

// GetQuote returns the latest quote for a symbol
func (s *AlpacaService) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	quote, err := s.dataClient.GetLatestQuote(symbol, marketdata.GetLatestQuoteRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get quote for %s: %w", symbol, err)
	}

	return &models.Quote{
		Symbol:    symbol,
		Bid:       decimal.NewFromFloat(quote.BidPrice),
		Ask:       decimal.NewFromFloat(quote.AskPrice),
		BidSize:   int64(quote.BidSize),
		AskSize:   int64(quote.AskSize),
		Timestamp: quote.Timestamp,
	}, nil
}

// GetLatestTrade returns the latest trade for a symbol
func (s *AlpacaService) GetLatestTrade(ctx context.Context, symbol string) (*models.Quote, error) {
	trade, err := s.dataClient.GetLatestTrade(symbol, marketdata.GetLatestTradeRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get trade for %s: %w", symbol, err)
	}

	return &models.Quote{
		Symbol:    symbol,
		Last:      decimal.NewFromFloat(trade.Price),
		Volume:    int64(trade.Size),
		Timestamp: trade.Timestamp,
	}, nil
}

// GetBars returns historical bars for a symbol
func (s *AlpacaService) GetBars(ctx context.Context, symbol string, start, end time.Time, timeframe marketdata.TimeFrame) ([]models.Bar, error) {
	bars, err := s.dataClient.GetBars(symbol, marketdata.GetBarsRequest{
		TimeFrame: timeframe,
		Start:     start,
		End:       end,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bars for %s: %w", symbol, err)
	}

	result := make([]models.Bar, 0, len(bars))
	for _, bar := range bars {
		result = append(result, models.Bar{
			Symbol:    symbol,
			Timestamp: bar.Timestamp,
			Open:      decimal.NewFromFloat(bar.Open),
			High:      decimal.NewFromFloat(bar.High),
			Low:       decimal.NewFromFloat(bar.Low),
			Close:     decimal.NewFromFloat(bar.Close),
			Volume:    int64(bar.Volume),
			VWAP:      decimal.NewFromFloat(bar.VWAP),
		})
	}

	return result, nil
}

// GetDailyBars returns daily bars for the last N days
func (s *AlpacaService) GetDailyBars(ctx context.Context, symbol string, days int) ([]models.Bar, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)

	return s.GetBars(ctx, symbol, start, end, marketdata.OneDay)
}

// PlaceOrder places a trade order
func (s *AlpacaService) PlaceOrder(ctx context.Context, symbol string, quantity decimal.Decimal, side models.TradeSide) (*alpaca.Order, error) {
	qty := quantity

	var alpacaSide alpaca.Side
	if side == models.TradeSideBuy {
		alpacaSide = alpaca.Buy
	} else {
		alpacaSide = alpaca.Sell
	}

	return s.tradeClient.PlaceOrder(alpaca.PlaceOrderRequest{
		Symbol:      symbol,
		Qty:         &qty,
		Side:        alpacaSide,
		Type:        alpaca.Market,
		TimeInForce: alpaca.Day,
	})
}

// GetPositions returns all current positions
func (s *AlpacaService) GetPositions(ctx context.Context) ([]alpaca.Position, error) {
	return s.tradeClient.GetPositions()
}

// GetPosition returns a specific position
func (s *AlpacaService) GetPosition(ctx context.Context, symbol string) (*alpaca.Position, error) {
	return s.tradeClient.GetPosition(symbol)
}
