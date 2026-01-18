package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trade-machine/config"
	"trade-machine/models"
	"trade-machine/observability"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PortfolioManagerRepository defines the repository operations needed by PortfolioManager
type PortfolioManagerRepository interface {
	CreateAgentRun(ctx context.Context, run *models.AgentRun) error
	UpdateAgentRun(ctx context.Context, run *models.AgentRun) error
	CreateRecommendation(ctx context.Context, rec *models.Recommendation) error
}

// AccountProvider provides account and position information for position sizing
type AccountProvider interface {
	GetAccount(ctx context.Context) (*models.Account, error)
	GetPosition(ctx context.Context, symbol string) (*models.Position, error)
	GetQuote(ctx context.Context, symbol string) (*models.Quote, error)
}

// PortfolioManager orchestrates all agents and generates recommendations
type PortfolioManager struct {
	agents          []Agent
	repo            PortfolioManagerRepository
	cfg             *config.Config
	positionSizer   PositionSizer
	accountProvider AccountProvider
	strategy        ActionStrategy
}

// NewPortfolioManager creates a new PortfolioManager
func NewPortfolioManager(repo PortfolioManagerRepository, cfg *config.Config, accountProvider AccountProvider) *PortfolioManager {
	// Create position sizer from config
	sizingConfig := PositionSizingConfig{
		MaxPositionPercent:   cfg.PositionSizing.MaxPositionPercent,
		RiskPercent:          cfg.PositionSizing.RiskPercent,
		MinShares:            cfg.PositionSizing.MinShares,
		MaxShares:            cfg.PositionSizing.MaxShares,
		UseConfidenceScaling: cfg.PositionSizing.UseConfidenceScaling,
	}

	strategy := createStrategyFromConfig(cfg)

	return &PortfolioManager{
		agents:          make([]Agent, 0),
		repo:            repo,
		cfg:             cfg,
		positionSizer:   NewDefaultPositionSizer(sizingConfig),
		accountProvider: accountProvider,
		strategy:        strategy,
	}
}

// createStrategyFromConfig creates the appropriate strategy based on config
func createStrategyFromConfig(cfg *config.Config) ActionStrategy {
	switch cfg.Agent.Strategy {
	case "conservative":
		return NewConservativeStrategy()
	case "aggressive":
		return NewAggressiveStrategy()
	case "custom":
		return NewCustomStrategy(
			cfg.Agent.BuyThreshold,
			cfg.Agent.SellThreshold,
			cfg.Agent.MinConfidence,
		)
	default:
		return NewDefaultStrategy()
	}
}

// RegisterAgent adds an agent to the manager
func (m *PortfolioManager) RegisterAgent(agent Agent) {
	m.agents = append(m.agents, agent)
}

// getAvailableAgents returns agents whose dependencies are healthy
func (m *PortfolioManager) getAvailableAgents(ctx context.Context) []Agent {
	available := make([]Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		if agent.IsAvailable(ctx) {
			available = append(available, agent)
		} else {
			observability.Warn("agent unavailable, skipping",
				"agent", agent.Name(),
				"required_services", agent.GetMetadata().RequiredServices)
		}
	}
	return available
}

// agentResult holds the result of an agent analysis attempt
type agentResult struct {
	agent    Agent
	analysis *Analysis
	err      error
}

// AnalyzeSymbol runs all agents and generates a recommendation
func (m *PortfolioManager) AnalyzeSymbol(ctx context.Context, symbol string) (*models.Recommendation, error) {
	// Record analysis request metric
	metrics := observability.GetMetrics()
	metrics.RecordAnalysisRequest(symbol)
	analysisTimer := metrics.NewTimer()

	// Track unavailable agents
	var unavailableAgents []models.MissingAgentInfo
	availableAgents := make([]Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		if agent.IsAvailable(ctx) {
			availableAgents = append(availableAgents, agent)
		} else {
			unavailableAgents = append(unavailableAgents, models.MissingAgentInfo{
				AgentType: agent.Type(),
				Reason:    fmt.Sprintf("%s unavailable: dependencies not healthy (%v)", agent.Name(), agent.GetMetadata().RequiredServices),
			})
			observability.Warn("agent unavailable, skipping",
				"agent", agent.Name(),
				"required_services", agent.GetMetadata().RequiredServices)
		}
	}

	if len(availableAgents) == 0 {
		analysisTimer.ObserveAnalysis(symbol, "error")
		metrics.RecordAnalysisError(symbol, "no_agents_available")
		return nil, fmt.Errorf("no agents available to analyze %s", symbol)
	}

	// Run all available agents in parallel
	var wg sync.WaitGroup
	results := make([]agentResult, len(availableAgents))

	for i, agent := range availableAgents {
		wg.Add(1)
		go func(idx int, ag Agent) {
			defer wg.Done()

			agentCtx, cancel := context.WithTimeout(ctx, time.Duration(m.cfg.Agent.TimeoutSeconds)*time.Second)
			defer cancel()

			run := models.NewAgentRun(ag.Type(), symbol)
			m.repo.CreateAgentRun(agentCtx, run)

			// Time the agent analysis
			agentTimer := metrics.NewTimer()
			analysis, err := ag.Analyze(agentCtx, symbol)
			agentTimer.ObserveAgent(string(ag.Type()))

			results[idx] = agentResult{agent: ag, analysis: analysis, err: err}

			if err != nil {
				run.Fail(err)
				metrics.RecordAgentError(string(ag.Type()), categorizeError(err))
			} else {
				run.Complete(map[string]interface{}{
					"score":      analysis.Score,
					"confidence": analysis.Confidence,
					"reasoning":  analysis.Reasoning,
				})
				metrics.RecordAgentScore(string(ag.Type()), analysis.Score)
			}

			m.repo.UpdateAgentRun(agentCtx, run)
		}(i, agent)
	}

	wg.Wait()

	// Collect successful analyses and track failed agents
	var validAnalyses []*Analysis
	var failedAgents []models.MissingAgentInfo
	for _, result := range results {
		if result.analysis != nil {
			validAnalyses = append(validAnalyses, result.analysis)
		} else if result.err != nil {
			failedAgents = append(failedAgents, models.MissingAgentInfo{
				AgentType: result.agent.Type(),
				Reason:    fmt.Sprintf("%s failed: %v", result.agent.Name(), result.err),
			})
			observability.Warn("agent analysis failed",
				"agent", result.agent.Name(),
				"symbol", symbol,
				"error", result.err)
		}
	}

	if len(validAnalyses) == 0 {
		analysisTimer.ObserveAnalysis(symbol, "error")
		metrics.RecordAnalysisError(symbol, "all_agents_failed")
		return nil, fmt.Errorf("all agents failed to analyze %s", symbol)
	}

	// Combine unavailable and failed agents
	allMissingAgents := append(unavailableAgents, failedAgents...)

	// Synthesize recommendation with missing agent info
	rec := m.synthesizeRecommendation(ctx, symbol, validAnalyses, allMissingAgents)

	// Save to database
	if err := m.repo.CreateRecommendation(ctx, rec); err != nil {
		analysisTimer.ObserveAnalysis(symbol, "error")
		metrics.RecordAnalysisError(symbol, "db_save_failed")
		return nil, fmt.Errorf("failed to save recommendation: %w", err)
	}

	// Record successful analysis and recommendation metrics
	analysisTimer.ObserveAnalysis(symbol, "success")
	metrics.RecordRecommendation(string(rec.Action), calculateFinalScore(rec), rec.Confidence)

	return rec, nil
}

// categorizeError categorizes an error for metrics labeling
func categorizeError(err error) string {
	if err == nil {
		return "none"
	}
	errStr := err.Error()
	switch {
	case contains(errStr, "timeout"), contains(errStr, "context deadline"):
		return "timeout"
	case contains(errStr, "circuit breaker"):
		return "circuit_breaker"
	case contains(errStr, "rate limit"), contains(errStr, "too many requests"):
		return "rate_limit"
	case contains(errStr, "connection"), contains(errStr, "network"):
		return "network"
	default:
		return "other"
	}
}

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsLower(s, substr)))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if matchLower(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func matchLower(a, b string) bool {
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// calculateFinalScore calculates the weighted final score from a recommendation
func calculateFinalScore(rec *models.Recommendation) float64 {
	// Use simple average since we don't have access to weights here
	return (rec.FundamentalScore + rec.SentimentScore + rec.TechnicalScore) / 3
}

// synthesizeRecommendation combines agent analyses into a recommendation
func (m *PortfolioManager) synthesizeRecommendation(ctx context.Context, symbol string, analyses []*Analysis, missingAgents []models.MissingAgentInfo) *models.Recommendation {
	var fundamentalScore, sentimentScore, technicalScore float64
	var totalWeight float64 = 0
	var weightedScore float64 = 0
	var reasonings []string

	weights := map[models.AgentType]float64{
		models.AgentTypeFundamental: m.cfg.Agent.WeightFundamental,
		models.AgentTypeNews:        m.cfg.Agent.WeightNews,
		models.AgentTypeTechnical:   m.cfg.Agent.WeightTechnical,
	}

	// Track which agent types provided analysis
	providedAnalysis := make(map[models.AgentType]bool)

	for _, analysis := range analyses {
		weight := weights[analysis.AgentType]
		weightedScore += analysis.Score * weight * (analysis.Confidence / 100)
		totalWeight += weight * (analysis.Confidence / 100)
		providedAnalysis[analysis.AgentType] = true

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

	// Calculate overall confidence
	avgConfidence := 0.0
	for _, analysis := range analyses {
		avgConfidence += analysis.Confidence
	}
	avgConfidence /= float64(len(analyses))

	// Calculate data completeness (what percentage of expected agents succeeded)
	totalExpectedAgents := 3 // fundamental, technical, news
	dataCompleteness := float64(len(analyses)) / float64(totalExpectedAgents) * 100

	// Reduce confidence when agents are missing (proportional penalty)
	if len(missingAgents) > 0 {
		// Reduce confidence by 15% for each missing agent, up to 45% reduction
		confidencePenalty := float64(len(missingAgents)) * 15.0
		if confidencePenalty > 45.0 {
			confidencePenalty = 45.0
		}
		avgConfidence = avgConfidence * (1 - confidencePenalty/100)
	}

	// Determine action based on score using strategy
	action := m.strategy.DetermineAction(finalScore, avgConfidence)

	// Build combined reasoning with explicit missing agent info
	var combinedReasoning string
	if len(missingAgents) > 0 {
		// Build list of missing agent types
		var missingTypes []string
		for _, ma := range missingAgents {
			missingTypes = append(missingTypes, string(ma.AgentType))
		}
		combinedReasoning = fmt.Sprintf(
			"Based on analysis from %d of %d agents (%s unavailable). ",
			len(analyses), totalExpectedAgents, formatMissingAgents(missingTypes),
		)
	} else {
		combinedReasoning = fmt.Sprintf(
			"Based on analysis from %d agents. ",
			len(analyses),
		)
	}

	// Add score summary
	combinedReasoning += fmt.Sprintf(
		"Scores - Fundamental: %.0f, Sentiment: %.0f, Technical: %.0f. Overall score: %.1f. ",
		fundamentalScore, sentimentScore, technicalScore, finalScore,
	)

	// Add note about reduced confidence if applicable
	if len(missingAgents) > 0 {
		combinedReasoning += "Note: Confidence reduced due to incomplete data. "
	}

	// Add individual agent reasonings
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
		DataCompleteness: dataCompleteness,
		MissingAgents:    missingAgents,
		Status:           models.RecommendationStatusPending,
		CreatedAt:        time.Now(),
	}

	// Calculate position size using PositionSizer
	rec.Quantity = m.calculatePositionSize(ctx, symbol, action, avgConfidence)

	return rec
}

// formatMissingAgents formats a list of missing agent types for display
func formatMissingAgents(types []string) string {
	if len(types) == 0 {
		return ""
	}
	if len(types) == 1 {
		return types[0]
	}
	if len(types) == 2 {
		return types[0] + " and " + types[1]
	}
	return types[0] + ", " + types[1] + ", and " + types[2]
}

// calculatePositionSize uses the PositionSizer to determine trade quantity
func (m *PortfolioManager) calculatePositionSize(ctx context.Context, symbol string, action models.RecommendationAction, confidence float64) decimal.Decimal {
	// Get account information
	account, err := m.accountProvider.GetAccount(ctx)
	if err != nil {
		observability.Warn("failed to get account for position sizing, using minimum",
			"symbol", symbol,
			"error", err)
		return decimal.NewFromInt(m.cfg.PositionSizing.MinShares)
	}

	// Get current price
	quote, err := m.accountProvider.GetQuote(ctx, symbol)
	if err != nil {
		observability.Warn("failed to get quote for position sizing, using minimum",
			"symbol", symbol,
			"error", err)
		return decimal.NewFromInt(m.cfg.PositionSizing.MinShares)
	}

	currentPrice := quote.Last
	if currentPrice.IsZero() {
		// Fall back to bid/ask midpoint
		if !quote.Bid.IsZero() && !quote.Ask.IsZero() {
			currentPrice = quote.Bid.Add(quote.Ask).Div(decimal.NewFromInt(2))
		}
	}

	// Get existing position (may be nil)
	existingPosition, _ := m.accountProvider.GetPosition(ctx, symbol)

	// Calculate quantity using position sizer
	quantity, err := m.positionSizer.CalculateQuantity(ctx, account, currentPrice, action, confidence, existingPosition)
	if err != nil {
		observability.Warn("position sizer error, using minimum",
			"symbol", symbol,
			"error", err)
		return decimal.NewFromInt(m.cfg.PositionSizing.MinShares)
	}

	return quantity
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

// GetStrategy returns the current action strategy
func (m *PortfolioManager) GetStrategy() ActionStrategy {
	return m.strategy
}

// SetStrategy sets a new action strategy
func (m *PortfolioManager) SetStrategy(strategy ActionStrategy) {
	m.strategy = strategy
}
