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

// alpacaTradeClient defines the interface for Alpaca trading operations (for testing)
type alpacaTradeClient interface {
	GetAccount() (*alpaca.Account, error)
	PlaceOrder(req alpaca.PlaceOrderRequest) (*alpaca.Order, error)
	GetPositions() ([]alpaca.Position, error)
	GetPosition(symbol string) (*alpaca.Position, error)
}

// alpacaDataClient defines the interface for Alpaca market data operations (for testing)
type alpacaDataClient interface {
	GetLatestQuote(symbol string, req marketdata.GetLatestQuoteRequest) (*marketdata.Quote, error)
	GetLatestTrade(symbol string, req marketdata.GetLatestTradeRequest) (*marketdata.Trade, error)
	GetBars(symbol string, req marketdata.GetBarsRequest) ([]marketdata.Bar, error)
}

// AlpacaService handles communication with Alpaca for trading and market data
type AlpacaService struct {
	tradeClient alpacaTradeClient
	dataClient  alpacaDataClient
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
func (s *AlpacaService) GetAccount(ctx context.Context) (map[string]interface{}, error) {
	return WithCircuitBreaker(ctx, BreakerAlpaca, func() (map[string]interface{}, error) {
		account, err := s.tradeClient.GetAccount()
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"id":                      account.ID,
			"account_number":          account.AccountNumber,
			"status":                  account.Status,
			"currency":                account.Currency,
			"buying_power":            account.BuyingPower,
			"cash":                    account.Cash,
			"portfolio_value":         account.PortfolioValue,
			"pattern_day_trader":      account.PatternDayTrader,
			"trading_blocked":         account.TradingBlocked,
			"transfers_blocked":       account.TransfersBlocked,
			"account_blocked":         account.AccountBlocked,
			"created_at":              account.CreatedAt,
			"trade_suspended_by_user": account.TradeSuspendedByUser,
			"multiplier":              account.Multiplier,
			"shorting_enabled":        account.ShortingEnabled,
			"equity":                  account.Equity,
			"last_equity":             account.LastEquity,
			"long_market_value":       account.LongMarketValue,
			"short_market_value":      account.ShortMarketValue,
			"initial_margin":          account.InitialMargin,
			"maintenance_margin":      account.MaintenanceMargin,
			"last_maintenance_margin": account.LastMaintenanceMargin,
			"sma":                     account.SMA,
			"daytrade_count":          account.DaytradeCount,
		}, nil
	})
}

// GetQuote returns the latest quote for a symbol
func (s *AlpacaService) GetQuote(ctx context.Context, symbol string) (*models.Quote, error) {
	return WithCircuitBreaker(ctx, BreakerAlpaca, func() (*models.Quote, error) {
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
	})
}

// GetLatestTrade returns the latest trade for a symbol
func (s *AlpacaService) GetLatestTrade(ctx context.Context, symbol string) (*models.Quote, error) {
	return WithCircuitBreaker(ctx, BreakerAlpaca, func() (*models.Quote, error) {
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
	})
}

// GetBars returns historical bars for a symbol
func (s *AlpacaService) GetBars(ctx context.Context, symbol string, start, end time.Time, timeframe marketdata.TimeFrame) ([]marketdata.Bar, error) {
	return WithCircuitBreaker(ctx, BreakerAlpaca, func() ([]marketdata.Bar, error) {
		bars, err := s.dataClient.GetBars(symbol, marketdata.GetBarsRequest{
			TimeFrame: timeframe,
			Start:     start,
			End:       end,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get bars for %s: %w", symbol, err)
		}

		return bars, nil
	})
}

// GetDailyBars returns daily bars for the last N days
func (s *AlpacaService) GetDailyBars(ctx context.Context, symbol string, days int) ([]marketdata.Bar, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)

	return s.GetBars(ctx, symbol, start, end, marketdata.OneDay)
}

// PlaceOrder places a trade order
func (s *AlpacaService) PlaceOrder(ctx context.Context, symbol string, quantity decimal.Decimal, side models.TradeSide, orderType string) (string, error) {
	return WithCircuitBreaker(ctx, BreakerAlpaca, func() (string, error) {
		qty := quantity

		var alpacaSide alpaca.Side
		if side == models.TradeSideBuy {
			alpacaSide = alpaca.Buy
		} else {
			alpacaSide = alpaca.Sell
		}

		var alpacaOrderType alpaca.OrderType
		switch orderType {
		case "limit":
			alpacaOrderType = alpaca.Limit
		case "stop":
			alpacaOrderType = alpaca.Stop
		case "stop_limit":
			alpacaOrderType = alpaca.StopLimit
		default:
			alpacaOrderType = alpaca.Market
		}

		order, err := s.tradeClient.PlaceOrder(alpaca.PlaceOrderRequest{
			Symbol:      symbol,
			Qty:         &qty,
			Side:        alpacaSide,
			Type:        alpacaOrderType,
			TimeInForce: alpaca.Day,
		})
		if err != nil {
			return "", fmt.Errorf("failed to place order: %w", err)
		}

		return order.ID, nil
	})
}

// GetPositions returns all current positions
func (s *AlpacaService) GetPositions(ctx context.Context) ([]models.Position, error) {
	return WithCircuitBreaker(ctx, BreakerAlpaca, func() ([]models.Position, error) {
		alpacaPositions, err := s.tradeClient.GetPositions()
		if err != nil {
			return nil, fmt.Errorf("failed to get positions: %w", err)
		}

		positions := make([]models.Position, 0, len(alpacaPositions))
		for _, ap := range alpacaPositions {
			side := models.PositionSideLong
			if ap.Side == "short" {
				side = models.PositionSideShort
			}

			currentPrice := decimal.Zero
			if ap.CurrentPrice != nil {
				currentPrice = *ap.CurrentPrice
			}

			unrealizedPL := decimal.Zero
			if ap.UnrealizedPL != nil {
				unrealizedPL = *ap.UnrealizedPL
			}

			positions = append(positions, models.Position{
				Symbol:        ap.Symbol,
				Quantity:      ap.Qty,
				AvgEntryPrice: ap.AvgEntryPrice,
				CurrentPrice:  currentPrice,
				UnrealizedPL:  unrealizedPL,
				Side:          side,
			})
		}

		return positions, nil
	})
}

// GetPosition returns a specific position
func (s *AlpacaService) GetPosition(ctx context.Context, symbol string) (*models.Position, error) {
	return WithCircuitBreaker(ctx, BreakerAlpaca, func() (*models.Position, error) {
		ap, err := s.tradeClient.GetPosition(symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to get position for %s: %w", symbol, err)
		}

		side := models.PositionSideLong
		if ap.Side == "short" {
			side = models.PositionSideShort
		}

		currentPrice := decimal.Zero
		if ap.CurrentPrice != nil {
			currentPrice = *ap.CurrentPrice
		}

		unrealizedPL := decimal.Zero
		if ap.UnrealizedPL != nil {
			unrealizedPL = *ap.UnrealizedPL
		}

		return &models.Position{
			Symbol:        ap.Symbol,
			Quantity:      ap.Qty,
			AvgEntryPrice: ap.AvgEntryPrice,
			CurrentPrice:  currentPrice,
			UnrealizedPL:  unrealizedPL,
			Side:          side,
		}, nil
	})
}
