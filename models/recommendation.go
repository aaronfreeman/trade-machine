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
	DataCompleteness float64              `json:"data_completeness"` // 0-100: percentage of agents that succeeded
	MissingAgents    []MissingAgentInfo   `json:"missing_agents,omitempty"`
	Status           RecommendationStatus `json:"status"`
	ApprovedAt       *time.Time           `json:"approved_at,omitempty"`
	RejectedAt       *time.Time           `json:"rejected_at,omitempty"`
	ExecutedTradeID  *uuid.UUID           `json:"executed_trade_id,omitempty"`
	CreatedAt        time.Time            `json:"created_at"`
}

// MissingAgentInfo captures information about an agent that was unavailable or failed
type MissingAgentInfo struct {
	AgentType AgentType `json:"agent_type"`
	Reason    string    `json:"reason"`
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
