package agents

import (
	"testing"

	"trade-machine/models"
)

func TestScoreToAction(t *testing.T) {
	tests := []struct {
		name  string
		score float64
		want  models.RecommendationAction
	}{
		{"strong buy signal", 75.0, models.RecommendationActionBuy},
		{"moderate buy signal", 30.0, models.RecommendationActionBuy},
		{"threshold buy", 26.0, models.RecommendationActionBuy},
		{"borderline buy", 25.1, models.RecommendationActionBuy}, // > 25 is buy
		{"neutral at 25", 25.0, models.RecommendationActionHold},
		{"neutral zero", 0.0, models.RecommendationActionHold},
		{"neutral at -25", -25.0, models.RecommendationActionHold},
		{"borderline sell", -25.1, models.RecommendationActionSell}, // < -25 is sell
		{"threshold sell", -26.0, models.RecommendationActionSell},
		{"moderate sell signal", -50.0, models.RecommendationActionSell},
		{"strong sell signal", -75.0, models.RecommendationActionSell},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScoreToAction(tt.score)
			if got != tt.want {
				t.Errorf("ScoreToAction(%v) = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

func TestNormalizeScore(t *testing.T) {
	tests := []struct {
		name  string
		score float64
		want  float64
	}{
		{"within range positive", 50.0, 50.0},
		{"within range negative", -50.0, -50.0},
		{"at max", 100.0, 100.0},
		{"at min", -100.0, -100.0},
		{"above max", 150.0, 100.0},
		{"below min", -150.0, -100.0},
		{"zero", 0.0, 0.0},
		{"way above max", 1000.0, 100.0},
		{"way below min", -1000.0, -100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeScore(tt.score)
			if got != tt.want {
				t.Errorf("NormalizeScore(%v) = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

func TestNormalizeConfidence(t *testing.T) {
	tests := []struct {
		name       string
		confidence float64
		want       float64
	}{
		{"within range", 50.0, 50.0},
		{"at max", 100.0, 100.0},
		{"at min", 0.0, 0.0},
		{"above max", 150.0, 100.0},
		{"below min", -50.0, 0.0},
		{"way above max", 1000.0, 100.0},
		{"slightly negative", -0.1, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeConfidence(tt.confidence)
			if got != tt.want {
				t.Errorf("NormalizeConfidence(%v) = %v, want %v", tt.confidence, got, tt.want)
			}
		})
	}
}

func TestAnalysis_Fields(t *testing.T) {
	analysis := &Analysis{
		Symbol:     "AAPL",
		AgentType:  models.AgentTypeFundamental,
		Score:      65.5,
		Confidence: 80.0,
		Reasoning:  "Strong fundamentals with good growth potential",
		Data: map[string]interface{}{
			"pe_ratio":   25.5,
			"market_cap": "2.5T",
		},
	}

	if analysis.Symbol != "AAPL" {
		t.Errorf("Symbol = %v, want 'AAPL'", analysis.Symbol)
	}
	if analysis.AgentType != models.AgentTypeFundamental {
		t.Errorf("AgentType = %v, want AgentTypeFundamental", analysis.AgentType)
	}
	if analysis.Score != 65.5 {
		t.Errorf("Score = %v, want 65.5", analysis.Score)
	}
	if analysis.Confidence != 80.0 {
		t.Errorf("Confidence = %v, want 80.0", analysis.Confidence)
	}
	if analysis.Data["pe_ratio"] != 25.5 {
		t.Errorf("Data[pe_ratio] = %v, want 25.5", analysis.Data["pe_ratio"])
	}
}

func TestRecommendation_Fields(t *testing.T) {
	rec := &Recommendation{
		Symbol:           "TSLA",
		Action:           models.RecommendationActionBuy,
		Confidence:       75.0,
		Reasoning:        "Multiple agents agree on bullish outlook",
		FundamentalScore: 60.0,
		SentimentScore:   70.0,
		TechnicalScore:   80.0,
	}

	if rec.Symbol != "TSLA" {
		t.Errorf("Symbol = %v, want 'TSLA'", rec.Symbol)
	}
	if rec.Action != models.RecommendationActionBuy {
		t.Errorf("Action = %v, want Buy", rec.Action)
	}

	// Calculate weighted average manually to verify
	expectedAvg := (60.0 + 70.0 + 80.0) / 3
	actualAvg := (rec.FundamentalScore + rec.SentimentScore + rec.TechnicalScore) / 3
	if actualAvg != expectedAvg {
		t.Errorf("Average score = %v, want %v", actualAvg, expectedAvg)
	}
}
