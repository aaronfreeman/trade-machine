package agents

import (
	"context"
	"testing"

	"trade-machine/models"

	"github.com/shopspring/decimal"
)

func TestDefaultPositionSizer_CalculateQuantity_Buy(t *testing.T) {
	tests := []struct {
		name             string
		config           PositionSizingConfig
		account          *models.Account
		currentPrice     decimal.Decimal
		action           models.RecommendationAction
		confidence       float64
		existingPosition *models.Position
		wantMin          decimal.Decimal
		wantMax          decimal.Decimal
	}{
		{
			name:   "basic buy with 10% max position",
			config: DefaultPositionSizingConfig(),
			account: &models.Account{
				PortfolioValue: decimal.NewFromInt(100000),
				BuyingPower:    decimal.NewFromInt(100000),
				Equity:         decimal.NewFromInt(100000),
			},
			currentPrice: decimal.NewFromInt(100),
			action:       models.RecommendationActionBuy,
			confidence:   75,
			wantMin:      decimal.NewFromInt(50),  // With confidence scaling
			wantMax:      decimal.NewFromInt(100), // Max 10% = $10000 / $100 = 100 shares
		},
		{
			name:   "buy with low confidence reduces position",
			config: DefaultPositionSizingConfig(),
			account: &models.Account{
				PortfolioValue: decimal.NewFromInt(100000),
				BuyingPower:    decimal.NewFromInt(100000),
				Equity:         decimal.NewFromInt(100000),
			},
			currentPrice: decimal.NewFromInt(100),
			action:       models.RecommendationActionBuy,
			confidence:   20,
			wantMin:      decimal.NewFromInt(50),
			wantMax:      decimal.NewFromInt(70), // Lower confidence = smaller position
		},
		{
			name:   "buy with high confidence increases position",
			config: DefaultPositionSizingConfig(),
			account: &models.Account{
				PortfolioValue: decimal.NewFromInt(100000),
				BuyingPower:    decimal.NewFromInt(100000),
				Equity:         decimal.NewFromInt(100000),
			},
			currentPrice: decimal.NewFromInt(100),
			action:       models.RecommendationActionBuy,
			confidence:   100,
			wantMin:      decimal.NewFromInt(90),
			wantMax:      decimal.NewFromInt(100), // Full position at max confidence
		},
		{
			name: "buy limited by buying power",
			config: PositionSizingConfig{
				MaxPositionPercent:   0.10,
				MinShares:            1,
				UseConfidenceScaling: false,
			},
			account: &models.Account{
				PortfolioValue: decimal.NewFromInt(100000),
				BuyingPower:    decimal.NewFromInt(5000), // Only $5000 available
				Equity:         decimal.NewFromInt(100000),
			},
			currentPrice: decimal.NewFromInt(100),
			action:       models.RecommendationActionBuy,
			confidence:   80,
			wantMin:      decimal.NewFromInt(50),
			wantMax:      decimal.NewFromInt(50), // Limited to $5000 / $100 = 50 shares
		},
		{
			name: "buy with max shares limit",
			config: PositionSizingConfig{
				MaxPositionPercent:   0.10,
				MinShares:            1,
				MaxShares:            25,
				UseConfidenceScaling: false,
			},
			account: &models.Account{
				PortfolioValue: decimal.NewFromInt(100000),
				BuyingPower:    decimal.NewFromInt(100000),
				Equity:         decimal.NewFromInt(100000),
			},
			currentPrice: decimal.NewFromInt(100),
			action:       models.RecommendationActionBuy,
			confidence:   80,
			wantMin:      decimal.NewFromInt(25),
			wantMax:      decimal.NewFromInt(25), // Capped at MaxShares
		},
		{
			name: "buy expensive stock",
			config: PositionSizingConfig{
				MaxPositionPercent:   0.10,
				MinShares:            1,
				UseConfidenceScaling: false,
			},
			account: &models.Account{
				PortfolioValue: decimal.NewFromInt(100000),
				BuyingPower:    decimal.NewFromInt(100000),
				Equity:         decimal.NewFromInt(100000),
			},
			currentPrice: decimal.NewFromInt(5000),
			action:       models.RecommendationActionBuy,
			confidence:   80,
			wantMin:      decimal.NewFromInt(2),
			wantMax:      decimal.NewFromInt(2), // $10000 / $5000 = 2 shares
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewDefaultPositionSizer(tt.config)
			ctx := context.Background()

			got, err := ps.CalculateQuantity(ctx, tt.account, tt.currentPrice, tt.action, tt.confidence, tt.existingPosition)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.LessThan(tt.wantMin) {
				t.Errorf("quantity = %s, want >= %s", got.String(), tt.wantMin.String())
			}
			if got.GreaterThan(tt.wantMax) {
				t.Errorf("quantity = %s, want <= %s", got.String(), tt.wantMax.String())
			}
		})
	}
}

func TestDefaultPositionSizer_CalculateQuantity_Sell(t *testing.T) {
	ps := NewDefaultPositionSizer(DefaultPositionSizingConfig())
	ctx := context.Background()

	t.Run("sell existing position", func(t *testing.T) {
		account := &models.Account{
			PortfolioValue: decimal.NewFromInt(100000),
			BuyingPower:    decimal.NewFromInt(50000),
		}
		existingPosition := &models.Position{
			Symbol:   "AAPL",
			Quantity: decimal.NewFromInt(50),
		}

		got, err := ps.CalculateQuantity(ctx, account, decimal.NewFromInt(150), models.RecommendationActionSell, 80, existingPosition)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should sell entire existing position
		if !got.Equal(decimal.NewFromInt(50)) {
			t.Errorf("quantity = %s, want 50 (full position)", got.String())
		}
	})

	t.Run("sell without existing position returns minimum", func(t *testing.T) {
		account := &models.Account{
			PortfolioValue: decimal.NewFromInt(100000),
			BuyingPower:    decimal.NewFromInt(50000),
		}

		got, err := ps.CalculateQuantity(ctx, account, decimal.NewFromInt(150), models.RecommendationActionSell, 80, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return minimum shares
		if !got.Equal(decimal.NewFromInt(1)) {
			t.Errorf("quantity = %s, want 1 (minimum)", got.String())
		}
	})
}

func TestDefaultPositionSizer_CalculateQuantity_Hold(t *testing.T) {
	ps := NewDefaultPositionSizer(DefaultPositionSizingConfig())
	ctx := context.Background()

	account := &models.Account{
		PortfolioValue: decimal.NewFromInt(100000),
		BuyingPower:    decimal.NewFromInt(100000),
	}

	got, err := ps.CalculateQuantity(ctx, account, decimal.NewFromInt(100), models.RecommendationActionHold, 80, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// HOLD should return zero
	if !got.IsZero() {
		t.Errorf("quantity = %s, want 0 for HOLD action", got.String())
	}
}

func TestDefaultPositionSizer_CalculateQuantity_EdgeCases(t *testing.T) {
	ps := NewDefaultPositionSizer(DefaultPositionSizingConfig())
	ctx := context.Background()

	t.Run("zero price returns minimum shares", func(t *testing.T) {
		account := &models.Account{
			PortfolioValue: decimal.NewFromInt(100000),
			BuyingPower:    decimal.NewFromInt(100000),
		}

		got, err := ps.CalculateQuantity(ctx, account, decimal.Zero, models.RecommendationActionBuy, 80, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !got.Equal(decimal.NewFromInt(1)) {
			t.Errorf("quantity = %s, want 1 (minimum for zero price)", got.String())
		}
	})

	t.Run("negative price returns minimum shares", func(t *testing.T) {
		account := &models.Account{
			PortfolioValue: decimal.NewFromInt(100000),
			BuyingPower:    decimal.NewFromInt(100000),
		}

		got, err := ps.CalculateQuantity(ctx, account, decimal.NewFromInt(-100), models.RecommendationActionBuy, 80, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !got.Equal(decimal.NewFromInt(1)) {
			t.Errorf("quantity = %s, want 1 (minimum for negative price)", got.String())
		}
	})

	t.Run("zero portfolio value returns minimum shares", func(t *testing.T) {
		account := &models.Account{
			PortfolioValue: decimal.Zero,
			BuyingPower:    decimal.NewFromInt(100000),
			Equity:         decimal.Zero,
		}

		got, err := ps.CalculateQuantity(ctx, account, decimal.NewFromInt(100), models.RecommendationActionBuy, 80, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !got.Equal(decimal.NewFromInt(1)) {
			t.Errorf("quantity = %s, want 1 (minimum for zero portfolio)", got.String())
		}
	})

	t.Run("uses equity when portfolio value is zero", func(t *testing.T) {
		config := PositionSizingConfig{
			MaxPositionPercent:   0.10,
			MinShares:            1,
			UseConfidenceScaling: false,
		}
		ps := NewDefaultPositionSizer(config)

		account := &models.Account{
			PortfolioValue: decimal.Zero,
			BuyingPower:    decimal.NewFromInt(100000),
			Equity:         decimal.NewFromInt(50000),
		}

		got, err := ps.CalculateQuantity(ctx, account, decimal.NewFromInt(100), models.RecommendationActionBuy, 80, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 10% of $50000 = $5000 / $100 = 50 shares
		if !got.Equal(decimal.NewFromInt(50)) {
			t.Errorf("quantity = %s, want 50 (using equity)", got.String())
		}
	})
}

func TestDefaultPositionSizer_ConfidenceScaling(t *testing.T) {
	config := PositionSizingConfig{
		MaxPositionPercent:   0.10,
		MinShares:            1,
		MaxShares:            0,
		UseConfidenceScaling: true,
	}
	ps := NewDefaultPositionSizer(config)
	ctx := context.Background()

	account := &models.Account{
		PortfolioValue: decimal.NewFromInt(100000),
		BuyingPower:    decimal.NewFromInt(100000),
	}
	currentPrice := decimal.NewFromInt(100)

	// Test various confidence levels
	confidences := []float64{0, 25, 50, 75, 100}
	var prevQuantity decimal.Decimal

	for _, conf := range confidences {
		got, err := ps.CalculateQuantity(ctx, account, currentPrice, models.RecommendationActionBuy, conf, nil)
		if err != nil {
			t.Fatalf("unexpected error at confidence %v: %v", conf, err)
		}

		// Quantity should increase with confidence
		if !prevQuantity.IsZero() && got.LessThan(prevQuantity) {
			t.Errorf("quantity at confidence %v (%s) should be >= quantity at lower confidence (%s)",
				conf, got.String(), prevQuantity.String())
		}
		prevQuantity = got
	}
}

func TestDefaultPositionSizingConfig(t *testing.T) {
	config := DefaultPositionSizingConfig()

	if config.MaxPositionPercent != 0.10 {
		t.Errorf("MaxPositionPercent = %v, want 0.10", config.MaxPositionPercent)
	}
	if config.RiskPercent != 0.02 {
		t.Errorf("RiskPercent = %v, want 0.02", config.RiskPercent)
	}
	if config.MinShares != 1 {
		t.Errorf("MinShares = %v, want 1", config.MinShares)
	}
	if config.MaxShares != 0 {
		t.Errorf("MaxShares = %v, want 0 (unlimited)", config.MaxShares)
	}
	if !config.UseConfidenceScaling {
		t.Error("UseConfidenceScaling should be true by default")
	}
}

func TestNewDefaultPositionSizer(t *testing.T) {
	config := PositionSizingConfig{
		MaxPositionPercent:   0.05,
		MinShares:            5,
		MaxShares:            100,
		UseConfidenceScaling: false,
	}

	ps := NewDefaultPositionSizer(config)

	if ps == nil {
		t.Fatal("NewDefaultPositionSizer returned nil")
	}
	if ps.config.MaxPositionPercent != 0.05 {
		t.Errorf("config.MaxPositionPercent = %v, want 0.05", ps.config.MaxPositionPercent)
	}
	if ps.config.MinShares != 5 {
		t.Errorf("config.MinShares = %v, want 5", ps.config.MinShares)
	}
}
