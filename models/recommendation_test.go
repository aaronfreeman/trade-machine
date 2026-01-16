package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestNewRecommendation(t *testing.T) {
	rec := NewRecommendation("AAPL", RecommendationActionBuy, "Strong fundamentals and positive momentum")

	if rec.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want 'AAPL'", rec.Symbol)
	}
	if rec.Action != RecommendationActionBuy {
		t.Errorf("Action = %v, want RecommendationActionBuy", rec.Action)
	}
	if rec.Reasoning != "Strong fundamentals and positive momentum" {
		t.Errorf("Reasoning = %v, want expected text", rec.Reasoning)
	}
	if rec.Status != RecommendationStatusPending {
		t.Errorf("Status = %v, want RecommendationStatusPending", rec.Status)
	}
	if rec.ID == [16]byte{} {
		t.Error("ID should not be zero UUID")
	}
	if rec.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestRecommendation_Approve(t *testing.T) {
	rec := NewRecommendation("AAPL", RecommendationActionBuy, "Test")

	if rec.ApprovedAt != nil {
		t.Error("ApprovedAt should be nil before approval")
	}

	rec.Approve()

	if rec.Status != RecommendationStatusApproved {
		t.Errorf("Status = %v, want RecommendationStatusApproved", rec.Status)
	}
	if rec.ApprovedAt == nil {
		t.Error("ApprovedAt should not be nil after approval")
	}
	if rec.RejectedAt != nil {
		t.Error("RejectedAt should still be nil after approval")
	}
}

func TestRecommendation_Reject(t *testing.T) {
	rec := NewRecommendation("AAPL", RecommendationActionSell, "Test")

	if rec.RejectedAt != nil {
		t.Error("RejectedAt should be nil before rejection")
	}

	rec.Reject()

	if rec.Status != RecommendationStatusRejected {
		t.Errorf("Status = %v, want RecommendationStatusRejected", rec.Status)
	}
	if rec.RejectedAt == nil {
		t.Error("RejectedAt should not be nil after rejection")
	}
	if rec.ApprovedAt != nil {
		t.Error("ApprovedAt should still be nil after rejection")
	}
}

func TestRecommendation_MarkExecuted(t *testing.T) {
	rec := NewRecommendation("AAPL", RecommendationActionBuy, "Test")
	tradeID := uuid.New()

	rec.MarkExecuted(tradeID)

	if rec.Status != RecommendationStatusExecuted {
		t.Errorf("Status = %v, want RecommendationStatusExecuted", rec.Status)
	}
	if rec.ExecutedTradeID == nil {
		t.Error("ExecutedTradeID should not be nil after execution")
	}
	if *rec.ExecutedTradeID != tradeID {
		t.Errorf("ExecutedTradeID = %v, want %v", *rec.ExecutedTradeID, tradeID)
	}
}

func TestRecommendationAction_Constants(t *testing.T) {
	actions := map[RecommendationAction]string{
		RecommendationActionBuy:  "buy",
		RecommendationActionSell: "sell",
		RecommendationActionHold: "hold",
	}

	for action, expected := range actions {
		if string(action) != expected {
			t.Errorf("RecommendationAction %v = %v, want '%v'", action, string(action), expected)
		}
	}
}

func TestRecommendationStatus_Constants(t *testing.T) {
	statuses := map[RecommendationStatus]string{
		RecommendationStatusPending:  "pending",
		RecommendationStatusApproved: "approved",
		RecommendationStatusRejected: "rejected",
		RecommendationStatusExecuted: "executed",
	}

	for status, expected := range statuses {
		if string(status) != expected {
			t.Errorf("RecommendationStatus %v = %v, want '%v'", status, string(status), expected)
		}
	}
}

func TestRecommendation_FullWorkflow(t *testing.T) {
	// Create recommendation
	rec := NewRecommendation("TSLA", RecommendationActionBuy, "Bullish setup")
	rec.Quantity = decimal.NewFromInt(10)
	rec.TargetPrice = decimal.NewFromFloat(250.00)
	rec.Confidence = 75.5
	rec.FundamentalScore = 60.0
	rec.SentimentScore = 70.0
	rec.TechnicalScore = 80.0

	// Verify initial state
	if rec.Status != RecommendationStatusPending {
		t.Errorf("Initial status = %v, want pending", rec.Status)
	}

	// Approve
	rec.Approve()
	if rec.Status != RecommendationStatusApproved {
		t.Errorf("After approve status = %v, want approved", rec.Status)
	}

	// Execute
	tradeID := uuid.New()
	rec.MarkExecuted(tradeID)
	if rec.Status != RecommendationStatusExecuted {
		t.Errorf("After execute status = %v, want executed", rec.Status)
	}

	// Verify scores preserved
	if rec.Confidence != 75.5 {
		t.Errorf("Confidence = %v, want 75.5", rec.Confidence)
	}
	if rec.FundamentalScore != 60.0 {
		t.Errorf("FundamentalScore = %v, want 60.0", rec.FundamentalScore)
	}
}

func TestRecommendation_Timestamps(t *testing.T) {
	rec := NewRecommendation("AAPL", RecommendationActionHold, "Wait and see")

	beforeApprove := time.Now()
	time.Sleep(1 * time.Millisecond)
	rec.Approve()
	time.Sleep(1 * time.Millisecond)
	afterApprove := time.Now()

	if rec.ApprovedAt.Before(beforeApprove) || rec.ApprovedAt.After(afterApprove) {
		t.Errorf("ApprovedAt = %v, should be between %v and %v", rec.ApprovedAt, beforeApprove, afterApprove)
	}
}
