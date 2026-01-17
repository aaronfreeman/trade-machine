package services

import (
	"context"
	"testing"
	"time"

	"trade-machine/models"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/shopspring/decimal"
)

func TestNewAlpacaService(t *testing.T) {
	service := NewAlpacaService("test-key", "test-secret", "https://paper-api.alpaca.markets")
	if service == nil {
		t.Error("NewAlpacaService should not return nil")
	}
	if service.tradeClient == nil {
		t.Error("tradeClient should not be nil")
	}
	if service.dataClient == nil {
		t.Error("dataClient should not be nil")
	}
}

func TestNewAlpacaService_EmptyCredentials(t *testing.T) {
	// Should still create service (will fail on actual API calls)
	service := NewAlpacaService("", "", "")
	if service == nil {
		t.Error("NewAlpacaService should not return nil even with empty credentials")
	}
}

func TestNewAlpacaService_VariousURLs(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
	}{
		{"Paper Trading", "https://paper-api.alpaca.markets"},
		{"Live Trading", "https://api.alpaca.markets"},
		{"Custom URL", "https://custom.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewAlpacaService("key", "secret", tt.baseURL)
			if service == nil {
				t.Error("NewAlpacaService should not return nil")
			}
			if service.tradeClient == nil {
				t.Error("tradeClient should not be nil")
			}
			if service.dataClient == nil {
				t.Error("dataClient should not be nil")
			}
		})
	}
}

func TestTradeSide_Conversion(t *testing.T) {
	// Test that our models.TradeSide matches expected values
	if models.TradeSideBuy != "buy" {
		t.Errorf("TradeSideBuy = %v, want 'buy'", models.TradeSideBuy)
	}
	if models.TradeSideSell != "sell" {
		t.Errorf("TradeSideSell = %v, want 'sell'", models.TradeSideSell)
	}
}

func TestGetAccount_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewAlpacaService("", "", "")
	ctx := context.Background()

	_, err := service.GetAccount(ctx)
	// We expect an error since we're using invalid credentials
	if err == nil {
		t.Error("GetAccount should return error with invalid credentials")
	}
}

func TestGetQuote_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewAlpacaService("", "", "")
	ctx := context.Background()

	_, err := service.GetQuote(ctx, "AAPL")
	// We expect an error since we're using invalid credentials
	if err == nil {
		t.Error("GetQuote should return error with invalid credentials")
	}
}

func TestGetLatestTrade_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewAlpacaService("", "", "")
	ctx := context.Background()

	_, err := service.GetLatestTrade(ctx, "AAPL")
	// We expect an error since we're using invalid credentials
	if err == nil {
		t.Error("GetLatestTrade should return error with invalid credentials")
	}
}

func TestGetBars_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewAlpacaService("", "", "")
	ctx := context.Background()

	start := time.Now().AddDate(0, 0, -7)
	end := time.Now()

	_, err := service.GetBars(ctx, "AAPL", start, end, marketdata.OneDay)
	// We expect an error since we're using invalid credentials
	if err == nil {
		t.Error("GetBars should return error with invalid credentials")
	}
}

func TestGetDailyBars_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewAlpacaService("", "", "")
	ctx := context.Background()

	_, err := service.GetDailyBars(ctx, "AAPL", 30)
	// We expect an error since we're using invalid credentials
	if err == nil {
		t.Error("GetDailyBars should return error with invalid credentials")
	}
}

func TestPlaceOrder_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewAlpacaService("", "", "")
	ctx := context.Background()

	tests := []struct {
		name      string
		symbol    string
		quantity  decimal.Decimal
		side      models.TradeSide
		orderType string
	}{
		{"Buy Market", "AAPL", decimal.NewFromInt(1), models.TradeSideBuy, "market"},
		{"Sell Market", "AAPL", decimal.NewFromInt(1), models.TradeSideSell, "market"},
		{"Buy Limit", "AAPL", decimal.NewFromInt(1), models.TradeSideBuy, "limit"},
		{"Buy Stop", "AAPL", decimal.NewFromInt(1), models.TradeSideBuy, "stop"},
		{"Buy Stop Limit", "AAPL", decimal.NewFromInt(1), models.TradeSideBuy, "stop_limit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.PlaceOrder(ctx, tt.symbol, tt.quantity, tt.side, tt.orderType)
			// We expect an error since we're using invalid credentials
			if err == nil {
				t.Error("PlaceOrder should return error with invalid credentials")
			}
		})
	}
}

func TestGetPositions_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewAlpacaService("", "", "")
	ctx := context.Background()

	_, err := service.GetPositions(ctx)
	// We expect an error since we're using invalid credentials
	if err == nil {
		t.Error("GetPositions should return error with invalid credentials")
	}
}

func TestGetPosition_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	service := NewAlpacaService("", "", "")
	ctx := context.Background()

	_, err := service.GetPosition(ctx, "AAPL")
	// We expect an error since we're using invalid credentials
	if err == nil {
		t.Error("GetPosition should return error with invalid credentials")
	}
}

func TestPositionSide_Values(t *testing.T) {
	// Test position side constants
	if models.PositionSideLong != "long" {
		t.Errorf("PositionSideLong = %v, want 'long'", models.PositionSideLong)
	}
	if models.PositionSideShort != "short" {
		t.Errorf("PositionSideShort = %v, want 'short'", models.PositionSideShort)
	}
}
