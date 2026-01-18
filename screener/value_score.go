package screener

import (
	"sort"

	"trade-machine/models"
)

// ValueScore calculates a composite value score for a screener candidate.
// Lower P/E and P/B ratios indicate better value, higher dividend yields are favorable.
// Score range: 0-100, where higher is better value.
func ValueScore(c models.ScreenerCandidate) float64 {
	// P/E Score: Lower is better
	// P/E of 0 = 100 score (max value)
	// P/E of 20 = 0 score
	// P/E > 20 = 0 score
	peScore := max(0, 100-c.PERatio*5)

	// P/B Score: Lower is better
	// P/B of 0 = 100 score (max value)
	// P/B of 2.5 = 0 score
	// P/B > 2.5 = 0 score
	pbScore := max(0, 100-c.PBRatio*40)

	// Dividend Score: Higher is better
	// 0% yield = 0 score
	// 5% yield = 100 score
	// > 5% yield capped at 100
	divScore := min(100, c.DividendYield*20)

	// Weighted average: 50% P/E, 30% P/B, 20% dividend
	return peScore*0.5 + pbScore*0.3 + divScore*0.2
}

// RankByValueScore sorts candidates by their value score in descending order
// and returns the top N candidates.
func RankByValueScore(candidates []models.ScreenerCandidate, topN int) []models.ScreenerCandidate {
	if len(candidates) == 0 {
		return candidates
	}

	// Calculate value scores for all candidates
	for i := range candidates {
		candidates[i].ValueScore = ValueScore(candidates[i])
	}

	// Sort by value score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].ValueScore > candidates[j].ValueScore
	})

	// Return top N
	if topN > 0 && topN < len(candidates) {
		return candidates[:topN]
	}
	return candidates
}

// RankByAnalysisScore sorts analyzed candidates by their combined score × confidence
// and returns the top N candidates.
func RankByAnalysisScore(candidates []models.ScreenerCandidate, topN int) []models.ScreenerCandidate {
	// Filter to only analyzed candidates
	analyzed := make([]models.ScreenerCandidate, 0, len(candidates))
	for _, c := range candidates {
		if c.Analyzed && c.Score != nil && c.Confidence != nil {
			analyzed = append(analyzed, c)
		}
	}

	if len(analyzed) == 0 {
		return analyzed
	}

	// Sort by combined score (score × confidence/100) descending
	sort.Slice(analyzed, func(i, j int) bool {
		scoreI := *analyzed[i].Score * (*analyzed[i].Confidence / 100)
		scoreJ := *analyzed[j].Score * (*analyzed[j].Confidence / 100)
		return scoreI > scoreJ
	})

	// Return top N
	if topN > 0 && topN < len(analyzed) {
		return analyzed[:topN]
	}
	return analyzed
}
