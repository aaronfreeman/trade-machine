package agents

import (
	"testing"

	"trade-machine/models"
)

func TestNewsAnalyst_Name(t *testing.T) {
	analyst := &NewsAnalyst{}
	if analyst.Name() != "News Sentiment Analyst" {
		t.Errorf("Name() = %v, want 'News Sentiment Analyst'", analyst.Name())
	}
}

func TestNewsAnalyst_Type(t *testing.T) {
	analyst := &NewsAnalyst{}
	if analyst.Type() != models.AgentTypeNews {
		t.Errorf("Type() = %v, want AgentTypeNews", analyst.Type())
	}
}

func TestNewNewsAnalyst(t *testing.T) {
	// Test with nil services (should not panic)
	analyst := NewNewsAnalyst(nil, nil)
	if analyst == nil {
		t.Error("NewNewsAnalyst should not return nil")
	}
}

func TestNewsAnalystResponse_Parsing(t *testing.T) {
	// Test that the response struct has correct fields
	resp := NewsAnalystResponse{
		Score:           45.0,
		Confidence:      70.0,
		Reasoning:       "Mixed sentiment with slight positive bias",
		KeyThemes:       []string{"Earnings beat", "Market volatility", "Product launch"},
		NotableArticles: []string{"AAPL beats Q4 expectations", "Market reacts to Fed news"},
	}

	if resp.Score != 45.0 {
		t.Errorf("Score = %v, want 45.0", resp.Score)
	}
	if resp.Confidence != 70.0 {
		t.Errorf("Confidence = %v, want 70.0", resp.Confidence)
	}
	if len(resp.KeyThemes) != 3 {
		t.Errorf("KeyThemes length = %v, want 3", len(resp.KeyThemes))
	}
	if len(resp.NotableArticles) != 2 {
		t.Errorf("NotableArticles length = %v, want 2", len(resp.NotableArticles))
	}
}
