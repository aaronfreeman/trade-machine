package models

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestNewTrade(t *testing.T) {
	trade := NewTrade("AAPL", TradeSideBuy, decimal.NewFromInt(10), decimal.NewFromFloat(150.00))

	if trade.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want 'AAPL'", trade.Symbol)
	}
	if trade.Side != TradeSideBuy {
		t.Errorf("Side = %v, want TradeSideBuy", trade.Side)
	}
	if !trade.Quantity.Equal(decimal.NewFromInt(10)) {
		t.Errorf("Quantity = %v, want 10", trade.Quantity)
	}
	if !trade.Price.Equal(decimal.NewFromFloat(150.00)) {
		t.Errorf("Price = %v, want 150.00", trade.Price)
	}
	if !trade.TotalValue.Equal(decimal.NewFromFloat(1500.00)) {
		t.Errorf("TotalValue = %v, want 1500.00", trade.TotalValue)
	}
	if trade.Status != TradeStatusPending {
		t.Errorf("Status = %v, want TradeStatusPending", trade.Status)
	}
	if trade.ID == [16]byte{} {
		t.Error("ID should not be zero UUID")
	}
	if trade.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestTradeSide_Constants(t *testing.T) {
	if TradeSideBuy != "buy" {
		t.Errorf("TradeSideBuy = %v, want 'buy'", TradeSideBuy)
	}
	if TradeSideSell != "sell" {
		t.Errorf("TradeSideSell = %v, want 'sell'", TradeSideSell)
	}
}

func TestTradeStatus_Constants(t *testing.T) {
	statuses := map[TradeStatus]string{
		TradeStatusPending:   "pending",
		TradeStatusExecuted:  "executed",
		TradeStatusRejected:  "rejected",
		TradeStatusCancelled: "cancelled",
	}

	for status, expected := range statuses {
		if string(status) != expected {
			t.Errorf("TradeStatus %v = %v, want '%v'", status, string(status), expected)
		}
	}
}

func TestTrade_Fields(t *testing.T) {
	now := time.Now()
	trade := NewTrade("MSFT", TradeSideSell, decimal.NewFromInt(5), decimal.NewFromFloat(300.00))
	trade.AlpacaOrderID = "order-123"
	trade.ExecutedAt = &now
	trade.Status = TradeStatusExecuted

	if trade.AlpacaOrderID != "order-123" {
		t.Errorf("AlpacaOrderID = %v, want 'order-123'", trade.AlpacaOrderID)
	}
	if trade.ExecutedAt == nil || !trade.ExecutedAt.Equal(now) {
		t.Errorf("ExecutedAt = %v, want %v", trade.ExecutedAt, now)
	}
	if trade.Status != TradeStatusExecuted {
		t.Errorf("Status = %v, want TradeStatusExecuted", trade.Status)
	}
}

func TestNewTrade_DifferentValues(t *testing.T) {
	tests := []struct {
		name      string
		symbol    string
		side      TradeSide
		quantity  decimal.Decimal
		price     decimal.Decimal
		wantTotal decimal.Decimal
	}{
		{
			name:      "buy 100 shares at $50",
			symbol:    "GOOGL",
			side:      TradeSideBuy,
			quantity:  decimal.NewFromInt(100),
			price:     decimal.NewFromFloat(50.00),
			wantTotal: decimal.NewFromFloat(5000.00),
		},
		{
			name:      "sell 25 shares at $200",
			symbol:    "TSLA",
			side:      TradeSideSell,
			quantity:  decimal.NewFromInt(25),
			price:     decimal.NewFromFloat(200.00),
			wantTotal: decimal.NewFromFloat(5000.00),
		},
		{
			name:      "fractional shares",
			symbol:    "AMZN",
			side:      TradeSideBuy,
			quantity:  decimal.NewFromFloat(0.5),
			price:     decimal.NewFromFloat(100.00),
			wantTotal: decimal.NewFromFloat(50.00),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trade := NewTrade(tt.symbol, tt.side, tt.quantity, tt.price)
			if !trade.TotalValue.Equal(tt.wantTotal) {
				t.Errorf("TotalValue = %v, want %v", trade.TotalValue, tt.wantTotal)
			}
		})
	}
}
