package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Position struct {
	ID            uuid.UUID       `json:"id"`
	Symbol        string          `json:"symbol"`
	Quantity      decimal.Decimal `json:"quantity"`
	AvgEntryPrice decimal.Decimal `json:"avg_entry_price"`
	CurrentPrice  decimal.Decimal `json:"current_price"`
	UnrealizedPL  decimal.Decimal `json:"unrealized_pl"`
	Side          PositionSide    `json:"side"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type PositionSide string

const (
	PositionSideLong  PositionSide = "long"
	PositionSideShort PositionSide = "short"
)

func (p *Position) CalculateUnrealizedPL() decimal.Decimal {
	if p.CurrentPrice.IsZero() {
		return decimal.Zero
	}
	priceDiff := p.CurrentPrice.Sub(p.AvgEntryPrice)
	if p.Side == PositionSideShort {
		priceDiff = priceDiff.Neg()
	}
	return priceDiff.Mul(p.Quantity)
}
