package screener

import (
	"testing"

	"trade-machine/models"
)

func TestValueScore(t *testing.T) {
	tests := []struct {
		name      string
		candidate models.ScreenerCandidate
		wantMin   float64
		wantMax   float64
	}{
		{
			name: "perfect value stock",
			candidate: models.ScreenerCandidate{
				PERatio:       0,
				PBRatio:       0,
				DividendYield: 5.0, // 5% yield
			},
			wantMin: 99.0, // Should be ~100
			wantMax: 100.0,
		},
		{
			name: "decent value stock",
			candidate: models.ScreenerCandidate{
				PERatio:       10,
				PBRatio:       1.0,
				DividendYield: 2.5,
			},
			wantMin: 50.0,
			wantMax: 70.0,
		},
		{
			name: "expensive stock",
			candidate: models.ScreenerCandidate{
				PERatio:       25,
				PBRatio:       3.0,
				DividendYield: 0,
			},
			wantMin: 0.0,
			wantMax: 10.0,
		},
		{
			name: "high dividend low P/E",
			candidate: models.ScreenerCandidate{
				PERatio:       5,
				PBRatio:       2.5, // At threshold
				DividendYield: 3.0,
			},
			wantMin: 40.0,
			wantMax: 60.0,
		},
		{
			name: "zero everything",
			candidate: models.ScreenerCandidate{
				PERatio:       0,
				PBRatio:       0,
				DividendYield: 0,
			},
			wantMin: 80.0, // 100*0.5 + 100*0.3 = 80
			wantMax: 80.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := ValueScore(tt.candidate)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("ValueScore() = %v, want between %v and %v", score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestValueScore_Components(t *testing.T) {
	// Test P/E component
	t.Run("P/E score", func(t *testing.T) {
		lowPE := models.ScreenerCandidate{PERatio: 5}
		highPE := models.ScreenerCandidate{PERatio: 20}

		lowPEScore := ValueScore(lowPE)
		highPEScore := ValueScore(highPE)

		if lowPEScore <= highPEScore {
			t.Errorf("Lower P/E should have higher score: low=%v, high=%v", lowPEScore, highPEScore)
		}
	})

	// Test P/B component
	t.Run("P/B score", func(t *testing.T) {
		lowPB := models.ScreenerCandidate{PBRatio: 0.5}
		highPB := models.ScreenerCandidate{PBRatio: 2.0}

		lowPBScore := ValueScore(lowPB)
		highPBScore := ValueScore(highPB)

		if lowPBScore <= highPBScore {
			t.Errorf("Lower P/B should have higher score: low=%v, high=%v", lowPBScore, highPBScore)
		}
	})

	// Test dividend component
	t.Run("Dividend score", func(t *testing.T) {
		lowDiv := models.ScreenerCandidate{DividendYield: 1.0}
		highDiv := models.ScreenerCandidate{DividendYield: 4.0}

		lowDivScore := ValueScore(lowDiv)
		highDivScore := ValueScore(highDiv)

		if lowDivScore >= highDivScore {
			t.Errorf("Higher dividend should have higher score: low=%v, high=%v", lowDivScore, highDivScore)
		}
	})
}

func TestRankByValueScore(t *testing.T) {
	candidates := []models.ScreenerCandidate{
		{Symbol: "EXPENSIVE", PERatio: 30, PBRatio: 5.0, DividendYield: 0},
		{Symbol: "VALUE", PERatio: 5, PBRatio: 0.5, DividendYield: 4.0},
		{Symbol: "MIDDLE", PERatio: 15, PBRatio: 2.0, DividendYield: 2.0},
	}

	t.Run("rank all", func(t *testing.T) {
		ranked := RankByValueScore(candidates, 0)

		if len(ranked) != 3 {
			t.Errorf("RankByValueScore() returned %d candidates, want 3", len(ranked))
		}

		// VALUE should be first (best value)
		if ranked[0].Symbol != "VALUE" {
			t.Errorf("First ranked should be VALUE, got %s", ranked[0].Symbol)
		}

		// EXPENSIVE should be last (worst value)
		if ranked[2].Symbol != "EXPENSIVE" {
			t.Errorf("Last ranked should be EXPENSIVE, got %s", ranked[2].Symbol)
		}

		// Verify scores are populated
		for _, c := range ranked {
			if c.ValueScore == 0 && c.Symbol != "EXPENSIVE" {
				t.Errorf("ValueScore should be calculated for %s", c.Symbol)
			}
		}
	})

	t.Run("rank top 2", func(t *testing.T) {
		ranked := RankByValueScore(candidates, 2)

		if len(ranked) != 2 {
			t.Errorf("RankByValueScore(2) returned %d candidates, want 2", len(ranked))
		}

		// VALUE and MIDDLE should be in top 2
		if ranked[0].Symbol != "VALUE" {
			t.Errorf("First should be VALUE, got %s", ranked[0].Symbol)
		}
	})

	t.Run("empty candidates", func(t *testing.T) {
		ranked := RankByValueScore([]models.ScreenerCandidate{}, 5)

		if len(ranked) != 0 {
			t.Errorf("Empty input should return empty result, got %d", len(ranked))
		}
	})

	t.Run("topN exceeds candidates", func(t *testing.T) {
		ranked := RankByValueScore(candidates, 100)

		if len(ranked) != 3 {
			t.Errorf("Should return all candidates when topN > len, got %d", len(ranked))
		}
	})
}

func TestRankByAnalysisScore(t *testing.T) {
	score1, conf1 := 80.0, 90.0
	score2, conf2 := 70.0, 80.0
	score3, conf3 := 90.0, 60.0 // High score but low confidence

	candidates := []models.ScreenerCandidate{
		{Symbol: "LOW", Score: &score2, Confidence: &conf2, Analyzed: true},   // 70 * 0.8 = 56
		{Symbol: "HIGH", Score: &score1, Confidence: &conf1, Analyzed: true},  // 80 * 0.9 = 72
		{Symbol: "RISKY", Score: &score3, Confidence: &conf3, Analyzed: true}, // 90 * 0.6 = 54
		{Symbol: "UNANALYZED", Analyzed: false},
	}

	t.Run("rank by combined score", func(t *testing.T) {
		ranked := RankByAnalysisScore(candidates, 0)

		if len(ranked) != 3 {
			t.Errorf("Should exclude unanalyzed, got %d", len(ranked))
		}

		if ranked[0].Symbol != "HIGH" {
			t.Errorf("First should be HIGH (best combined), got %s", ranked[0].Symbol)
		}
		if ranked[1].Symbol != "LOW" {
			t.Errorf("Second should be LOW, got %s", ranked[1].Symbol)
		}
		if ranked[2].Symbol != "RISKY" {
			t.Errorf("Third should be RISKY (low confidence hurts), got %s", ranked[2].Symbol)
		}
	})

	t.Run("rank top 1", func(t *testing.T) {
		ranked := RankByAnalysisScore(candidates, 1)

		if len(ranked) != 1 {
			t.Errorf("Should return 1, got %d", len(ranked))
		}
		if ranked[0].Symbol != "HIGH" {
			t.Errorf("Top 1 should be HIGH, got %s", ranked[0].Symbol)
		}
	})

	t.Run("no analyzed candidates", func(t *testing.T) {
		unanalyzed := []models.ScreenerCandidate{
			{Symbol: "A", Analyzed: false},
			{Symbol: "B", Analyzed: false},
		}

		ranked := RankByAnalysisScore(unanalyzed, 5)

		if len(ranked) != 0 {
			t.Errorf("Should return empty for unanalyzed, got %d", len(ranked))
		}
	})

	t.Run("missing score or confidence", func(t *testing.T) {
		partial := []models.ScreenerCandidate{
			{Symbol: "NO_SCORE", Confidence: &conf1, Analyzed: true},
			{Symbol: "NO_CONF", Score: &score1, Analyzed: true},
			{Symbol: "COMPLETE", Score: &score1, Confidence: &conf1, Analyzed: true},
		}

		ranked := RankByAnalysisScore(partial, 0)

		if len(ranked) != 1 {
			t.Errorf("Should only include complete candidates, got %d", len(ranked))
		}
		if ranked[0].Symbol != "COMPLETE" {
			t.Errorf("Should be COMPLETE, got %s", ranked[0].Symbol)
		}
	})
}
