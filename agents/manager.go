package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trade-machine/models"
	"trade-machine/repository"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PortfolioManager orchestrates all agents and generates recommendations
type PortfolioManager struct {
	agents []Agent
	repo   *repository.Repository
}

// NewPortfolioManager creates a new PortfolioManager
func NewPortfolioManager(repo *repository.Repository) *PortfolioManager {
	return &PortfolioManager{
		agents: make([]Agent, 0),
		repo:   repo,
	}
}

// RegisterAgent adds an agent to the manager
func (m *PortfolioManager) RegisterAgent(agent Agent) {
	m.agents = append(m.agents, agent)
}

// AnalyzeSymbol runs all agents and generates a recommendation
func (m *PortfolioManager) AnalyzeSymbol(ctx context.Context, symbol string) (*models.Recommendation, error) {
	// Run all agents in parallel
	var wg sync.WaitGroup
	analyses := make([]*Analysis, len(m.agents))
	errors := make([]error, len(m.agents))

	for i, agent := range m.agents {
		wg.Add(1)
		go func(idx int, ag Agent) {
			defer wg.Done()

			// Log agent run start
			run := models.NewAgentRun(ag.Type(), symbol)
			m.repo.CreateAgentRun(ctx, run)

			analysis, err := ag.Analyze(ctx, symbol)
			if err != nil {
				errors[idx] = err
				run.Fail(err)
			} else {
				analyses[idx] = analysis
				run.Complete(map[string]interface{}{
					"score":      analysis.Score,
					"confidence": analysis.Confidence,
					"reasoning":  analysis.Reasoning,
				})
			}

			m.repo.UpdateAgentRun(ctx, run)
		}(i, agent)
	}

	wg.Wait()

	// Collect successful analyses
	var validAnalyses []*Analysis
	for i, analysis := range analyses {
		if analysis != nil {
			validAnalyses = append(validAnalyses, analysis)
		} else if errors[i] != nil {
			fmt.Printf("Agent %s failed: %v\n", m.agents[i].Name(), errors[i])
		}
	}

	if len(validAnalyses) == 0 {
		return nil, fmt.Errorf("all agents failed to analyze %s", symbol)
	}

	// Synthesize recommendation
	rec := m.synthesizeRecommendation(symbol, validAnalyses)

	// Save to database
	if err := m.repo.CreateRecommendation(ctx, rec); err != nil {
		return nil, fmt.Errorf("failed to save recommendation: %w", err)
	}

	return rec, nil
}

// synthesizeRecommendation combines agent analyses into a recommendation
func (m *PortfolioManager) synthesizeRecommendation(symbol string, analyses []*Analysis) *models.Recommendation {
	var fundamentalScore, sentimentScore, technicalScore float64
	var totalWeight float64 = 0
	var weightedScore float64 = 0
	var reasonings []string

	// Weight factors for different analysis types
	weights := map[models.AgentType]float64{
		models.AgentTypeFundamental: 0.4, // Fundamental analysis weighted highest
		models.AgentTypeNews:        0.3, // News sentiment
		models.AgentTypeTechnical:   0.3, // Technical analysis
	}

	for _, analysis := range analyses {
		weight := weights[analysis.AgentType]
		weightedScore += analysis.Score * weight * (analysis.Confidence / 100)
		totalWeight += weight * (analysis.Confidence / 100)

		switch analysis.AgentType {
		case models.AgentTypeFundamental:
			fundamentalScore = analysis.Score
		case models.AgentTypeNews:
			sentimentScore = analysis.Score
		case models.AgentTypeTechnical:
			technicalScore = analysis.Score
		}

		reasonings = append(reasonings, fmt.Sprintf("[%s] %s", analysis.AgentType, analysis.Reasoning))
	}

	// Calculate final score
	var finalScore float64
	if totalWeight > 0 {
		finalScore = weightedScore / totalWeight
	}

	// Determine action based on score
	action := ScoreToAction(finalScore)

	// Calculate overall confidence
	avgConfidence := 0.0
	for _, analysis := range analyses {
		avgConfidence += analysis.Confidence
	}
	avgConfidence /= float64(len(analyses))

	// Build combined reasoning
	combinedReasoning := fmt.Sprintf(
		"Based on analysis from %d agents (Fundamental: %.0f, Sentiment: %.0f, Technical: %.0f), overall score is %.1f. ",
		len(analyses), fundamentalScore, sentimentScore, technicalScore, finalScore,
	)

	for _, r := range reasonings {
		combinedReasoning += r + " "
	}

	rec := &models.Recommendation{
		ID:               uuid.New(),
		Symbol:           symbol,
		Action:           action,
		Confidence:       NormalizeConfidence(avgConfidence),
		Reasoning:        combinedReasoning,
		FundamentalScore: fundamentalScore,
		SentimentScore:   sentimentScore,
		TechnicalScore:   technicalScore,
		Status:           models.RecommendationStatusPending,
		CreatedAt:        time.Now(),
	}

	// Suggest quantity based on confidence (placeholder logic)
	if action == models.RecommendationActionBuy {
		rec.Quantity = decimal.NewFromInt(10) // Default to 10 shares
	} else if action == models.RecommendationActionSell {
		rec.Quantity = decimal.NewFromInt(10)
	}

	return rec
}

// Name returns the manager name
func (m *PortfolioManager) Name() string {
	return "Portfolio Manager"
}

// Type returns the manager type
func (m *PortfolioManager) Type() models.AgentType {
	return models.AgentTypeManager
}

// GetAgents returns all registered agents
func (m *PortfolioManager) GetAgents() []Agent {
	return m.agents
}
