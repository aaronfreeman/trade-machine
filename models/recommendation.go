package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Recommendation struct {
	ID               uuid.UUID            `json:"id"`
	Symbol           string               `json:"symbol"`
	Action           RecommendationAction `json:"action"`
	Quantity         decimal.Decimal      `json:"quantity"`
	TargetPrice      decimal.Decimal      `json:"target_price"`
	Confidence       float64              `json:"confidence"`
	Reasoning        string               `json:"reasoning"`
	FundamentalScore float64              `json:"fundamental_score"`
	SentimentScore   float64              `json:"sentiment_score"`
	TechnicalScore   float64              `json:"technical_score"`
	Status           RecommendationStatus `json:"status"`
	ApprovedAt       *time.Time           `json:"approved_at,omitempty"`
	RejectedAt       *time.Time           `json:"rejected_at,omitempty"`
	ExecutedTradeID  *uuid.UUID           `json:"executed_trade_id,omitempty"`
	CreatedAt        time.Time            `json:"created_at"`
}

type RecommendationAction string

const (
	RecommendationActionBuy  RecommendationAction = "buy"
	RecommendationActionSell RecommendationAction = "sell"
	RecommendationActionHold RecommendationAction = "hold"
)

type RecommendationStatus string

const (
	RecommendationStatusPending  RecommendationStatus = "pending"
	RecommendationStatusApproved RecommendationStatus = "approved"
	RecommendationStatusRejected RecommendationStatus = "rejected"
	RecommendationStatusExecuted RecommendationStatus = "executed"
)

func NewRecommendation(symbol string, action RecommendationAction, reasoning string) *Recommendation {
	return &Recommendation{
		ID:        uuid.New(),
		Symbol:    symbol,
		Action:    action,
		Reasoning: reasoning,
		Status:    RecommendationStatusPending,
		CreatedAt: time.Now(),
	}
}

func (r *Recommendation) Approve() {
	now := time.Now()
	r.ApprovedAt = &now
	r.Status = RecommendationStatusApproved
}

func (r *Recommendation) Reject() {
	now := time.Now()
	r.RejectedAt = &now
	r.Status = RecommendationStatusRejected
}

func (r *Recommendation) MarkExecuted(tradeID uuid.UUID) {
	r.ExecutedTradeID = &tradeID
	r.Status = RecommendationStatusExecuted
}
