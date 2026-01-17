package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"trade-machine/models"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
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

// Mock implementations for Alpaca clients

type mockAlpacaTradeClient struct {
	getAccountFunc    func() (*alpaca.Account, error)
	placeOrderFunc    func(req alpaca.PlaceOrderRequest) (*alpaca.Order, error)
	getPositionsFunc  func() ([]alpaca.Position, error)
	getPositionFunc   func(symbol string) (*alpaca.Position, error)
}

func (m *mockAlpacaTradeClient) GetAccount() (*alpaca.Account, error) {
	return m.getAccountFunc()
}

func (m *mockAlpacaTradeClient) PlaceOrder(req alpaca.PlaceOrderRequest) (*alpaca.Order, error) {
	return m.placeOrderFunc(req)
}

func (m *mockAlpacaTradeClient) GetPositions() ([]alpaca.Position, error) {
	return m.getPositionsFunc()
}

func (m *mockAlpacaTradeClient) GetPosition(symbol string) (*alpaca.Position, error) {
	return m.getPositionFunc(symbol)
}

type mockAlpacaDataClient struct {
	getLatestQuoteFunc func(symbol string, req marketdata.GetLatestQuoteRequest) (*marketdata.Quote, error)
	getLatestTradeFunc func(symbol string, req marketdata.GetLatestTradeRequest) (*marketdata.Trade, error)
	getBarsFunc        func(symbol string, req marketdata.GetBarsRequest) ([]marketdata.Bar, error)
}

func (m *mockAlpacaDataClient) GetLatestQuote(symbol string, req marketdata.GetLatestQuoteRequest) (*marketdata.Quote, error) {
	return m.getLatestQuoteFunc(symbol, req)
}

func (m *mockAlpacaDataClient) GetLatestTrade(symbol string, req marketdata.GetLatestTradeRequest) (*marketdata.Trade, error) {
	return m.getLatestTradeFunc(symbol, req)
}

func (m *mockAlpacaDataClient) GetBars(symbol string, req marketdata.GetBarsRequest) ([]marketdata.Bar, error) {
	return m.getBarsFunc(symbol, req)
}

func newTestAlpacaService(tradeClient alpacaTradeClient, dataClient alpacaDataClient) *AlpacaService {
	return &AlpacaService{
		tradeClient: tradeClient,
		dataClient:  dataClient,
	}
}

func TestPlaceOrder_Success(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockTrade := &mockAlpacaTradeClient{
		placeOrderFunc: func(req alpaca.PlaceOrderRequest) (*alpaca.Order, error) {
			return &alpaca.Order{ID: "order-123"}, nil
		},
	}
	mockData := &mockAlpacaDataClient{}

	service := newTestAlpacaService(mockTrade, mockData)
	ctx := context.Background()

	orderID, err := service.PlaceOrder(ctx, "AAPL", decimal.NewFromInt(10), models.TradeSideBuy, "market")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if orderID != "order-123" {
		t.Errorf("expected order-123, got %s", orderID)
	}
}

func TestPlaceOrder_SellSide(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockTrade := &mockAlpacaTradeClient{
		placeOrderFunc: func(req alpaca.PlaceOrderRequest) (*alpaca.Order, error) {
			if req.Side != alpaca.Sell {
				t.Errorf("expected sell side, got %v", req.Side)
			}
			return &alpaca.Order{ID: "sell-order"}, nil
		},
	}
	mockData := &mockAlpacaDataClient{}

	service := newTestAlpacaService(mockTrade, mockData)
	ctx := context.Background()

	_, err := service.PlaceOrder(ctx, "AAPL", decimal.NewFromInt(5), models.TradeSideSell, "market")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlaceOrder_OrderTypes(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	tests := []struct {
		orderType     string
		expectedType  alpaca.OrderType
	}{
		{"market", alpaca.Market},
		{"limit", alpaca.Limit},
		{"stop", alpaca.Stop},
		{"stop_limit", alpaca.StopLimit},
		{"unknown", alpaca.Market}, // defaults to market
	}

	for _, tt := range tests {
		t.Run(tt.orderType, func(t *testing.T) {
			mockTrade := &mockAlpacaTradeClient{
				placeOrderFunc: func(req alpaca.PlaceOrderRequest) (*alpaca.Order, error) {
					if req.Type != tt.expectedType {
						t.Errorf("expected %v, got %v", tt.expectedType, req.Type)
					}
					return &alpaca.Order{ID: "test"}, nil
				},
			}
			mockData := &mockAlpacaDataClient{}

			service := newTestAlpacaService(mockTrade, mockData)
			ctx := context.Background()

			_, err := service.PlaceOrder(ctx, "AAPL", decimal.NewFromInt(1), models.TradeSideBuy, tt.orderType)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPlaceOrder_Error(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockTrade := &mockAlpacaTradeClient{
		placeOrderFunc: func(req alpaca.PlaceOrderRequest) (*alpaca.Order, error) {
			return nil, errors.New("insufficient funds")
		},
	}
	mockData := &mockAlpacaDataClient{}

	service := newTestAlpacaService(mockTrade, mockData)
	ctx := context.Background()

	_, err := service.PlaceOrder(ctx, "AAPL", decimal.NewFromInt(10), models.TradeSideBuy, "market")
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetPositions_Success(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	currentPrice := decimal.NewFromFloat(150.00)
	unrealizedPL := decimal.NewFromFloat(50.00)

	mockTrade := &mockAlpacaTradeClient{
		getPositionsFunc: func() ([]alpaca.Position, error) {
			return []alpaca.Position{
				{
					Symbol:        "AAPL",
					Qty:           decimal.NewFromInt(10),
					AvgEntryPrice: decimal.NewFromFloat(145.00),
					CurrentPrice:  &currentPrice,
					UnrealizedPL:  &unrealizedPL,
					Side:          "long",
				},
			}, nil
		},
	}
	mockData := &mockAlpacaDataClient{}

	service := newTestAlpacaService(mockTrade, mockData)
	ctx := context.Background()

	positions, err := service.GetPositions(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positions))
	}
	if positions[0].Symbol != "AAPL" {
		t.Errorf("expected AAPL, got %s", positions[0].Symbol)
	}
	if positions[0].Side != models.PositionSideLong {
		t.Errorf("expected long, got %s", positions[0].Side)
	}
}

func TestGetPositions_ShortSide(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockTrade := &mockAlpacaTradeClient{
		getPositionsFunc: func() ([]alpaca.Position, error) {
			return []alpaca.Position{
				{
					Symbol:        "TSLA",
					Qty:           decimal.NewFromInt(5),
					AvgEntryPrice: decimal.NewFromFloat(200.00),
					Side:          "short",
				},
			}, nil
		},
	}
	mockData := &mockAlpacaDataClient{}

	service := newTestAlpacaService(mockTrade, mockData)
	ctx := context.Background()

	positions, err := service.GetPositions(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if positions[0].Side != models.PositionSideShort {
		t.Errorf("expected short, got %s", positions[0].Side)
	}
}

func TestGetPositions_NilFields(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockTrade := &mockAlpacaTradeClient{
		getPositionsFunc: func() ([]alpaca.Position, error) {
			return []alpaca.Position{
				{
					Symbol:        "NVDA",
					Qty:           decimal.NewFromInt(3),
					AvgEntryPrice: decimal.NewFromFloat(500.00),
					CurrentPrice:  nil, // nil pointer
					UnrealizedPL:  nil, // nil pointer
					Side:          "long",
				},
			}, nil
		},
	}
	mockData := &mockAlpacaDataClient{}

	service := newTestAlpacaService(mockTrade, mockData)
	ctx := context.Background()

	positions, err := service.GetPositions(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !positions[0].CurrentPrice.IsZero() {
		t.Errorf("expected zero current price, got %s", positions[0].CurrentPrice)
	}
	if !positions[0].UnrealizedPL.IsZero() {
		t.Errorf("expected zero unrealized P/L, got %s", positions[0].UnrealizedPL)
	}
}

func TestGetPositions_Error(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockTrade := &mockAlpacaTradeClient{
		getPositionsFunc: func() ([]alpaca.Position, error) {
			return nil, errors.New("API error")
		},
	}
	mockData := &mockAlpacaDataClient{}

	service := newTestAlpacaService(mockTrade, mockData)
	ctx := context.Background()

	_, err := service.GetPositions(ctx)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetPosition_Success(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	currentPrice := decimal.NewFromFloat(175.00)
	unrealizedPL := decimal.NewFromFloat(25.00)

	mockTrade := &mockAlpacaTradeClient{
		getPositionFunc: func(symbol string) (*alpaca.Position, error) {
			if symbol != "AAPL" {
				t.Errorf("expected AAPL, got %s", symbol)
			}
			return &alpaca.Position{
				Symbol:        symbol,
				Qty:           decimal.NewFromInt(20),
				AvgEntryPrice: decimal.NewFromFloat(173.75),
				CurrentPrice:  &currentPrice,
				UnrealizedPL:  &unrealizedPL,
				Side:          "long",
			}, nil
		},
	}
	mockData := &mockAlpacaDataClient{}

	service := newTestAlpacaService(mockTrade, mockData)
	ctx := context.Background()

	position, err := service.GetPosition(ctx, "AAPL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if position.Symbol != "AAPL" {
		t.Errorf("expected AAPL, got %s", position.Symbol)
	}
	if position.Quantity.IntPart() != 20 {
		t.Errorf("expected 20 shares, got %d", position.Quantity.IntPart())
	}
}

func TestGetPosition_ShortSide(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockTrade := &mockAlpacaTradeClient{
		getPositionFunc: func(symbol string) (*alpaca.Position, error) {
			return &alpaca.Position{
				Symbol:        symbol,
				Qty:           decimal.NewFromInt(10),
				AvgEntryPrice: decimal.NewFromFloat(100.00),
				Side:          "short",
			}, nil
		},
	}
	mockData := &mockAlpacaDataClient{}

	service := newTestAlpacaService(mockTrade, mockData)
	ctx := context.Background()

	position, err := service.GetPosition(ctx, "GME")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if position.Side != models.PositionSideShort {
		t.Errorf("expected short, got %s", position.Side)
	}
}

func TestGetPosition_NilFields(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockTrade := &mockAlpacaTradeClient{
		getPositionFunc: func(symbol string) (*alpaca.Position, error) {
			return &alpaca.Position{
				Symbol:        symbol,
				Qty:           decimal.NewFromInt(5),
				AvgEntryPrice: decimal.NewFromFloat(50.00),
				CurrentPrice:  nil,
				UnrealizedPL:  nil,
				Side:          "long",
			}, nil
		},
	}
	mockData := &mockAlpacaDataClient{}

	service := newTestAlpacaService(mockTrade, mockData)
	ctx := context.Background()

	position, err := service.GetPosition(ctx, "XYZ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !position.CurrentPrice.IsZero() {
		t.Errorf("expected zero current price")
	}
	if !position.UnrealizedPL.IsZero() {
		t.Errorf("expected zero unrealized P/L")
	}
}

func TestGetPosition_Error(t *testing.T) {
	SetGlobalRegistry(NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig))

	mockTrade := &mockAlpacaTradeClient{
		getPositionFunc: func(symbol string) (*alpaca.Position, error) {
			return nil, errors.New("position not found")
		},
	}
	mockData := &mockAlpacaDataClient{}

	service := newTestAlpacaService(mockTrade, mockData)
	ctx := context.Background()

	_, err := service.GetPosition(ctx, "INVALID")
	if err == nil {
		t.Error("expected error")
	}
}
