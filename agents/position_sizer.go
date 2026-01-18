package agents

import (
	"context"

	"trade-machine/models"

	"github.com/shopspring/decimal"
)

// PositionSizer calculates the appropriate position size for a trade
type PositionSizer interface {
	// CalculateQuantity determines the number of shares to trade based on
	// account information, current price, action, and confidence level
	CalculateQuantity(
		ctx context.Context,
		account *models.Account,
		currentPrice decimal.Decimal,
		action models.RecommendationAction,
		confidence float64,
		existingPosition *models.Position,
	) (decimal.Decimal, error)
}

// PositionSizingConfig holds configuration for position sizing
type PositionSizingConfig struct {
	// MaxPositionPercent is the maximum percentage of portfolio value for a single position (0-1)
	MaxPositionPercent float64

	// RiskPercent is the base percentage of portfolio to risk per trade (0-1)
	RiskPercent float64

	// MinShares is the minimum number of shares to recommend
	MinShares int64

	// MaxShares is the maximum number of shares for a single position (0 = unlimited)
	MaxShares int64

	// UseConfidenceScaling whether to scale position size by confidence
	UseConfidenceScaling bool
}

// DefaultPositionSizingConfig returns sensible defaults for position sizing
func DefaultPositionSizingConfig() PositionSizingConfig {
	return PositionSizingConfig{
		MaxPositionPercent:   0.10, // Max 10% of portfolio in single position
		RiskPercent:          0.02, // Risk 2% of portfolio per trade
		MinShares:            1,
		MaxShares:            0, // Unlimited
		UseConfidenceScaling: true,
	}
}

// DefaultPositionSizer implements position sizing based on portfolio percentage
type DefaultPositionSizer struct {
	config PositionSizingConfig
}

// NewDefaultPositionSizer creates a new DefaultPositionSizer
func NewDefaultPositionSizer(config PositionSizingConfig) *DefaultPositionSizer {
	return &DefaultPositionSizer{
		config: config,
	}
}

// CalculateQuantity determines position size based on:
// - Portfolio value and buying power
// - Maximum position size as percentage of portfolio
// - Confidence level (optionally scales the position)
// - Existing position in the symbol
func (ps *DefaultPositionSizer) CalculateQuantity(
	ctx context.Context,
	account *models.Account,
	currentPrice decimal.Decimal,
	action models.RecommendationAction,
	confidence float64,
	existingPosition *models.Position,
) (decimal.Decimal, error) {
	// HOLD actions don't need quantity
	if action == models.RecommendationActionHold {
		return decimal.Zero, nil
	}

	// Can't calculate without price
	if currentPrice.IsZero() || currentPrice.IsNegative() {
		return decimal.NewFromInt(ps.config.MinShares), nil
	}

	// Handle SELL - sell existing position or minimum
	if action == models.RecommendationActionSell {
		if existingPosition != nil && existingPosition.Quantity.GreaterThan(decimal.Zero) {
			return existingPosition.Quantity, nil
		}
		return decimal.NewFromInt(ps.config.MinShares), nil
	}

	// Handle BUY - calculate based on portfolio
	portfolioValue := account.PortfolioValue
	if portfolioValue.IsZero() || portfolioValue.IsNegative() {
		portfolioValue = account.Equity
	}
	if portfolioValue.IsZero() || portfolioValue.IsNegative() {
		return decimal.NewFromInt(ps.config.MinShares), nil
	}

	// Calculate max position value based on portfolio percentage
	maxPositionPercent := decimal.NewFromFloat(ps.config.MaxPositionPercent)
	maxPositionValue := portfolioValue.Mul(maxPositionPercent)

	// Apply confidence scaling if enabled
	if ps.config.UseConfidenceScaling {
		// Scale between 50% and 100% of max position based on confidence
		// Low confidence (0-50) = 50-75% of max
		// High confidence (50-100) = 75-100% of max
		confidenceFactor := 0.5 + (confidence / 200.0) // Maps 0-100 to 0.5-1.0
		maxPositionValue = maxPositionValue.Mul(decimal.NewFromFloat(confidenceFactor))
	}

	// Limit by available buying power
	if account.BuyingPower.LessThan(maxPositionValue) {
		maxPositionValue = account.BuyingPower
	}

	// Calculate number of shares
	shares := maxPositionValue.Div(currentPrice).Floor()

	// Apply minimum
	minShares := decimal.NewFromInt(ps.config.MinShares)
	if shares.LessThan(minShares) {
		shares = minShares
	}

	// Apply maximum if set
	if ps.config.MaxShares > 0 {
		maxShares := decimal.NewFromInt(ps.config.MaxShares)
		if shares.GreaterThan(maxShares) {
			shares = maxShares
		}
	}

	return shares, nil
}
