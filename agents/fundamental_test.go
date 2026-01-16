package agents

import (
	"testing"

	"trade-machine/models"
)

func TestFundamentalAnalyst_Name(t *testing.T) {
	analyst := &FundamentalAnalyst{}
	if analyst.Name() != "Fundamental Analyst" {
		t.Errorf("Name() = %v, want 'Fundamental Analyst'", analyst.Name())
	}
}

func TestFundamentalAnalyst_Type(t *testing.T) {
	analyst := &FundamentalAnalyst{}
	if analyst.Type() != models.AgentTypeFundamental {
		t.Errorf("Type() = %v, want AgentTypeFundamental", analyst.Type())
	}
}

func TestNewFundamentalAnalyst(t *testing.T) {
	// Test with nil services (should not panic)
	analyst := NewFundamentalAnalyst(nil, nil)
	if analyst == nil {
		t.Error("NewFundamentalAnalyst should not return nil")
	}
}

func TestFundamentalAnalystResponse_Parsing(t *testing.T) {
	// Test that the response struct has correct fields
	resp := FundamentalAnalystResponse{
		Score:      75.5,
		Confidence: 80.0,
		Reasoning:  "Strong financials",
		KeyFactors: []string{"P/E below industry average", "Growing revenue"},
	}

	if resp.Score != 75.5 {
		t.Errorf("Score = %v, want 75.5", resp.Score)
	}
	if resp.Confidence != 80.0 {
		t.Errorf("Confidence = %v, want 80.0", resp.Confidence)
	}
	if len(resp.KeyFactors) != 2 {
		t.Errorf("KeyFactors length = %v, want 2", len(resp.KeyFactors))
	}
}
