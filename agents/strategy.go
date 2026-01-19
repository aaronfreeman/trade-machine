package agents

import "trade-machine/models"

// ActionStrategy defines the interface for determining trading actions from scores
type ActionStrategy interface {
	// DetermineAction converts a score and confidence into a recommendation action
	DetermineAction(score float64, confidence float64) models.RecommendationAction
	// Name returns the strategy name for logging/display
	Name() string
}

// DefaultStrategy uses the standard ±25 thresholds
type DefaultStrategy struct {
	BuyThreshold  float64
	SellThreshold float64
}

// NewDefaultStrategy creates a strategy with the standard thresholds
func NewDefaultStrategy() *DefaultStrategy {
	return &DefaultStrategy{
		BuyThreshold:  25,
		SellThreshold: -25,
	}
}

func (s *DefaultStrategy) DetermineAction(score float64, confidence float64) models.RecommendationAction {
	if score > s.BuyThreshold {
		return models.RecommendationActionBuy
	}
	if score < s.SellThreshold {
		return models.RecommendationActionSell
	}
	return models.RecommendationActionHold
}

func (s *DefaultStrategy) Name() string {
	return "default"
}

// ConservativeStrategy requires higher thresholds and minimum confidence
type ConservativeStrategy struct {
	BuyThreshold  float64
	SellThreshold float64
	MinConfidence float64
}

// NewConservativeStrategy creates a conservative strategy with higher thresholds
func NewConservativeStrategy() *ConservativeStrategy {
	return &ConservativeStrategy{
		BuyThreshold:  35,
		SellThreshold: -35,
		MinConfidence: 60,
	}
}

func (s *ConservativeStrategy) DetermineAction(score float64, confidence float64) models.RecommendationAction {
	if confidence < s.MinConfidence {
		return models.RecommendationActionHold
	}
	if score > s.BuyThreshold {
		return models.RecommendationActionBuy
	}
	if score < s.SellThreshold {
		return models.RecommendationActionSell
	}
	return models.RecommendationActionHold
}

func (s *ConservativeStrategy) Name() string {
	return "conservative"
}

// AggressiveStrategy uses lower thresholds for more active trading
type AggressiveStrategy struct {
	BuyThreshold  float64
	SellThreshold float64
}

// NewAggressiveStrategy creates an aggressive strategy with lower thresholds
func NewAggressiveStrategy() *AggressiveStrategy {
	return &AggressiveStrategy{
		BuyThreshold:  15,
		SellThreshold: -15,
	}
}

func (s *AggressiveStrategy) DetermineAction(score float64, confidence float64) models.RecommendationAction {
	if score > s.BuyThreshold {
		return models.RecommendationActionBuy
	}
	if score < s.SellThreshold {
		return models.RecommendationActionSell
	}
	return models.RecommendationActionHold
}

func (s *AggressiveStrategy) Name() string {
	return "aggressive"
}

// CustomStrategy allows fully configurable thresholds
type CustomStrategy struct {
	BuyThreshold  float64
	SellThreshold float64
	MinConfidence float64
	StrategyName  string
}

// NewCustomStrategy creates a strategy with custom thresholds
func NewCustomStrategy(buyThreshold, sellThreshold, minConfidence float64) *CustomStrategy {
	return &CustomStrategy{
		BuyThreshold:  buyThreshold,
		SellThreshold: sellThreshold,
		MinConfidence: minConfidence,
		StrategyName:  "custom",
	}
}

func (s *CustomStrategy) DetermineAction(score float64, confidence float64) models.RecommendationAction {
	if s.MinConfidence > 0 && confidence < s.MinConfidence {
		return models.RecommendationActionHold
	}
	if score > s.BuyThreshold {
		return models.RecommendationActionBuy
	}
	if score < s.SellThreshold {
		return models.RecommendationActionSell
	}
	return models.RecommendationActionHold
}

func (s *CustomStrategy) Name() string {
	return s.StrategyName
}

// StrategyFromName returns a strategy by name
func StrategyFromName(name string) ActionStrategy {
	switch name {
	case "conservative":
		return NewConservativeStrategy()
	case "aggressive":
		return NewAggressiveStrategy()
	default:
		return NewDefaultStrategy()
	}
}
