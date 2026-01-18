package agents

import (
	"testing"

	"trade-machine/models"
)

func TestDefaultStrategy_DetermineAction(t *testing.T) {
	strategy := NewDefaultStrategy()

	tests := []struct {
		name       string
		score      float64
		confidence float64
		expected   models.RecommendationAction
	}{
		{"strong buy", 50.0, 80.0, models.RecommendationActionBuy},
		{"threshold buy", 26.0, 80.0, models.RecommendationActionBuy},
		{"exactly at buy threshold", 25.0, 80.0, models.RecommendationActionHold},
		{"strong sell", -50.0, 80.0, models.RecommendationActionSell},
		{"threshold sell", -26.0, 80.0, models.RecommendationActionSell},
		{"exactly at sell threshold", -25.0, 80.0, models.RecommendationActionHold},
		{"neutral positive", 20.0, 80.0, models.RecommendationActionHold},
		{"neutral negative", -20.0, 80.0, models.RecommendationActionHold},
		{"zero score", 0.0, 80.0, models.RecommendationActionHold},
		{"low confidence doesn't affect default", 50.0, 10.0, models.RecommendationActionBuy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.DetermineAction(tt.score, tt.confidence)
			if result != tt.expected {
				t.Errorf("DetermineAction(%f, %f) = %s, want %s", tt.score, tt.confidence, result, tt.expected)
			}
		})
	}
}

func TestDefaultStrategy_Name(t *testing.T) {
	strategy := NewDefaultStrategy()
	if strategy.Name() != "default" {
		t.Errorf("Name() = %s, want default", strategy.Name())
	}
}

func TestConservativeStrategy_DetermineAction(t *testing.T) {
	strategy := NewConservativeStrategy()

	tests := []struct {
		name       string
		score      float64
		confidence float64
		expected   models.RecommendationAction
	}{
		{"strong buy high confidence", 50.0, 80.0, models.RecommendationActionBuy},
		{"threshold buy high confidence", 36.0, 70.0, models.RecommendationActionBuy},
		{"above default but below conservative", 30.0, 80.0, models.RecommendationActionHold},
		{"strong sell high confidence", -50.0, 80.0, models.RecommendationActionSell},
		{"threshold sell high confidence", -36.0, 70.0, models.RecommendationActionSell},
		{"above default but below conservative sell", -30.0, 80.0, models.RecommendationActionHold},
		{"buy score but low confidence", 50.0, 50.0, models.RecommendationActionHold},
		{"sell score but low confidence", -50.0, 50.0, models.RecommendationActionHold},
		{"exactly at confidence threshold", 50.0, 60.0, models.RecommendationActionBuy},
		{"just below confidence threshold", 50.0, 59.9, models.RecommendationActionHold},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.DetermineAction(tt.score, tt.confidence)
			if result != tt.expected {
				t.Errorf("DetermineAction(%f, %f) = %s, want %s", tt.score, tt.confidence, result, tt.expected)
			}
		})
	}
}

func TestConservativeStrategy_Name(t *testing.T) {
	strategy := NewConservativeStrategy()
	if strategy.Name() != "conservative" {
		t.Errorf("Name() = %s, want conservative", strategy.Name())
	}
}

func TestAggressiveStrategy_DetermineAction(t *testing.T) {
	strategy := NewAggressiveStrategy()

	tests := []struct {
		name       string
		score      float64
		confidence float64
		expected   models.RecommendationAction
	}{
		{"strong buy", 50.0, 80.0, models.RecommendationActionBuy},
		{"low threshold buy", 16.0, 80.0, models.RecommendationActionBuy},
		{"exactly at threshold", 15.0, 80.0, models.RecommendationActionHold},
		{"strong sell", -50.0, 80.0, models.RecommendationActionSell},
		{"low threshold sell", -16.0, 80.0, models.RecommendationActionSell},
		{"exactly at sell threshold", -15.0, 80.0, models.RecommendationActionHold},
		{"narrow hold range", 10.0, 80.0, models.RecommendationActionHold},
		{"low confidence doesn't affect", 20.0, 10.0, models.RecommendationActionBuy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.DetermineAction(tt.score, tt.confidence)
			if result != tt.expected {
				t.Errorf("DetermineAction(%f, %f) = %s, want %s", tt.score, tt.confidence, result, tt.expected)
			}
		})
	}
}

func TestAggressiveStrategy_Name(t *testing.T) {
	strategy := NewAggressiveStrategy()
	if strategy.Name() != "aggressive" {
		t.Errorf("Name() = %s, want aggressive", strategy.Name())
	}
}

func TestCustomStrategy_DetermineAction(t *testing.T) {
	strategy := NewCustomStrategy(30, -30, 50)

	tests := []struct {
		name       string
		score      float64
		confidence float64
		expected   models.RecommendationAction
	}{
		{"above custom buy threshold", 35.0, 60.0, models.RecommendationActionBuy},
		{"at custom buy threshold", 30.0, 60.0, models.RecommendationActionHold},
		{"below custom sell threshold", -35.0, 60.0, models.RecommendationActionSell},
		{"at custom sell threshold", -30.0, 60.0, models.RecommendationActionHold},
		{"low confidence holds", 50.0, 40.0, models.RecommendationActionHold},
		{"at min confidence threshold", 35.0, 50.0, models.RecommendationActionBuy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.DetermineAction(tt.score, tt.confidence)
			if result != tt.expected {
				t.Errorf("DetermineAction(%f, %f) = %s, want %s", tt.score, tt.confidence, result, tt.expected)
			}
		})
	}
}

func TestCustomStrategy_NoMinConfidence(t *testing.T) {
	strategy := NewCustomStrategy(20, -20, 0) // No min confidence

	result := strategy.DetermineAction(25.0, 10.0)
	if result != models.RecommendationActionBuy {
		t.Errorf("Expected BUY with no min confidence, got %s", result)
	}
}

func TestCustomStrategy_Name(t *testing.T) {
	strategy := NewCustomStrategy(25, -25, 0)
	if strategy.Name() != "custom" {
		t.Errorf("Name() = %s, want custom", strategy.Name())
	}
}

func TestStrategyFromName(t *testing.T) {
	tests := []struct {
		name         string
		strategyName string
		expectedType string
	}{
		{"default strategy", "default", "default"},
		{"conservative strategy", "conservative", "conservative"},
		{"aggressive strategy", "aggressive", "aggressive"},
		{"unknown defaults to default", "unknown", "default"},
		{"empty defaults to default", "", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := StrategyFromName(tt.strategyName)
			if strategy.Name() != tt.expectedType {
				t.Errorf("StrategyFromName(%s).Name() = %s, want %s", tt.strategyName, strategy.Name(), tt.expectedType)
			}
		})
	}
}

func TestStrategyInterface(t *testing.T) {
	// Verify all strategies implement the interface
	strategies := []ActionStrategy{
		NewDefaultStrategy(),
		NewConservativeStrategy(),
		NewAggressiveStrategy(),
		NewCustomStrategy(25, -25, 0),
	}

	for _, s := range strategies {
		// Should not panic
		_ = s.Name()
		_ = s.DetermineAction(0, 50)
	}
}
