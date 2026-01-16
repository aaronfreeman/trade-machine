package services

import (
	"testing"

	"trade-machine/models"
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

func TestTradeSide_Conversion(t *testing.T) {
	// Test that our models.TradeSide matches expected values
	if models.TradeSideBuy != "buy" {
		t.Errorf("TradeSideBuy = %v, want 'buy'", models.TradeSideBuy)
	}
	if models.TradeSideSell != "sell" {
		t.Errorf("TradeSideSell = %v, want 'sell'", models.TradeSideSell)
	}
}
