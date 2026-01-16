package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestPosition_CalculateUnrealizedPL(t *testing.T) {
	tests := []struct {
		name     string
		position Position
		want     decimal.Decimal
	}{
		{
			name: "long position with profit",
			position: Position{
				Quantity:      decimal.NewFromInt(10),
				AvgEntryPrice: decimal.NewFromFloat(100.00),
				CurrentPrice:  decimal.NewFromFloat(110.00),
				Side:          PositionSideLong,
			},
			want: decimal.NewFromFloat(100.00), // (110-100) * 10 = 100
		},
		{
			name: "long position with loss",
			position: Position{
				Quantity:      decimal.NewFromInt(10),
				AvgEntryPrice: decimal.NewFromFloat(100.00),
				CurrentPrice:  decimal.NewFromFloat(90.00),
				Side:          PositionSideLong,
			},
			want: decimal.NewFromFloat(-100.00), // (90-100) * 10 = -100
		},
		{
			name: "short position with profit",
			position: Position{
				Quantity:      decimal.NewFromInt(10),
				AvgEntryPrice: decimal.NewFromFloat(100.00),
				CurrentPrice:  decimal.NewFromFloat(90.00),
				Side:          PositionSideShort,
			},
			want: decimal.NewFromFloat(100.00), // (100-90) * 10 = 100
		},
		{
			name: "short position with loss",
			position: Position{
				Quantity:      decimal.NewFromInt(10),
				AvgEntryPrice: decimal.NewFromFloat(100.00),
				CurrentPrice:  decimal.NewFromFloat(110.00),
				Side:          PositionSideShort,
			},
			want: decimal.NewFromFloat(-100.00), // (100-110) * 10 = -100
		},
		{
			name: "zero current price returns zero",
			position: Position{
				Quantity:      decimal.NewFromInt(10),
				AvgEntryPrice: decimal.NewFromFloat(100.00),
				CurrentPrice:  decimal.Zero,
				Side:          PositionSideLong,
			},
			want: decimal.Zero,
		},
		{
			name: "breakeven position",
			position: Position{
				Quantity:      decimal.NewFromInt(10),
				AvgEntryPrice: decimal.NewFromFloat(100.00),
				CurrentPrice:  decimal.NewFromFloat(100.00),
				Side:          PositionSideLong,
			},
			want: decimal.Zero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.position.CalculateUnrealizedPL()
			if !got.Equal(tt.want) {
				t.Errorf("CalculateUnrealizedPL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPositionSide_Constants(t *testing.T) {
	if PositionSideLong != "long" {
		t.Errorf("PositionSideLong = %v, want 'long'", PositionSideLong)
	}
	if PositionSideShort != "short" {
		t.Errorf("PositionSideShort = %v, want 'short'", PositionSideShort)
	}
}

func TestPosition_Fields(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	pos := Position{
		ID:            id,
		Symbol:        "AAPL",
		Quantity:      decimal.NewFromInt(100),
		AvgEntryPrice: decimal.NewFromFloat(150.50),
		CurrentPrice:  decimal.NewFromFloat(155.00),
		UnrealizedPL:  decimal.NewFromFloat(450.00),
		Side:          PositionSideLong,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if pos.ID != id {
		t.Errorf("ID = %v, want %v", pos.ID, id)
	}
	if pos.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want 'AAPL'", pos.Symbol)
	}
	if !pos.Quantity.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Quantity = %v, want 100", pos.Quantity)
	}
}
