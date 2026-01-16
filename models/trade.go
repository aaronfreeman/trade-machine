package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Trade struct {
	ID            uuid.UUID       `json:"id"`
	Symbol        string          `json:"symbol"`
	Side          TradeSide       `json:"side"`
	Quantity      decimal.Decimal `json:"quantity"`
	Price         decimal.Decimal `json:"price"`
	TotalValue    decimal.Decimal `json:"total_value"`
	Commission    decimal.Decimal `json:"commission"`
	Status        TradeStatus     `json:"status"`
	AlpacaOrderID string          `json:"alpaca_order_id,omitempty"`
	ExecutedAt    *time.Time      `json:"executed_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

type TradeSide string

const (
	TradeSideBuy  TradeSide = "buy"
	TradeSideSell TradeSide = "sell"
)

type TradeStatus string

const (
	TradeStatusPending   TradeStatus = "pending"
	TradeStatusExecuted  TradeStatus = "executed"
	TradeStatusRejected  TradeStatus = "rejected"
	TradeStatusCancelled TradeStatus = "cancelled"
)

func NewTrade(symbol string, side TradeSide, quantity, price decimal.Decimal) *Trade {
	return &Trade{
		ID:         uuid.New(),
		Symbol:     symbol,
		Side:       side,
		Quantity:   quantity,
		Price:      price,
		TotalValue: quantity.Mul(price),
		Commission: decimal.Zero,
		Status:     TradeStatusPending,
		CreatedAt:  time.Now(),
	}
}
