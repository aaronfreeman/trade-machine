package screener

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trade-machine/config"
	"trade-machine/models"
	"trade-machine/observability"
	"trade-machine/services"

	"github.com/google/uuid"
)

// AnalysisProvider defines the interface for running stock analysis
type AnalysisProvider interface {
	AnalyzeSymbol(ctx context.Context, symbol string) (*models.Recommendation, error)
}

// ScreenerRepository defines the repository operations needed by ValueScreener
type ScreenerRepository interface {
	CreateScreenerRun(ctx context.Context, run *models.ScreenerRun) error
	UpdateScreenerRun(ctx context.Context, run *models.ScreenerRun) error
	GetScreenerRun(ctx context.Context, id uuid.UUID) (*models.ScreenerRun, error)
	GetLatestScreenerRun(ctx context.Context) (*models.ScreenerRun, error)
	GetScreenerRunHistory(ctx context.Context, limit int) ([]models.ScreenerRun, error)
	CreateRecommendation(ctx context.Context, rec *models.Recommendation) error
}

// ValueScreener orchestrates the full value screening workflow
type ValueScreener struct {
	fmpService       services.FMPServiceInterface
	analysisProvider AnalysisProvider
	repo             ScreenerRepository
	cfg              *config.ScreenerConfig
}

// NewValueScreener creates a new ValueScreener
func NewValueScreener(
	fmpService services.FMPServiceInterface,
	analysisProvider AnalysisProvider,
	repo ScreenerRepository,
	cfg *config.ScreenerConfig,
) *ValueScreener {
	return &ValueScreener{
		fmpService:       fmpService,
		analysisProvider: analysisProvider,
		repo:             repo,
		cfg:              cfg,
	}
}

// RunScreen executes a full screening workflow:
// 1. Fetch candidates from FMP
// 2. Pre-filter by value score
// 3. Run full analysis on top candidates
// 4. Return top picks
func (s *ValueScreener) RunScreen(ctx context.Context) (*models.ScreenerRun, error) {
	startTime := time.Now()

	// Build screening criteria from config
	criteria := models.ScreenerCriteria{
		MarketCapMin: s.cfg.MarketCapMin,
		PERatioMax:   s.cfg.PERatioMax,
		PBRatioMax:   s.cfg.PBRatioMax,
		Limit:        s.cfg.PreFilterLimit * 2, // Fetch more than we need for pre-filtering
	}

	// Create screener run record
	run := models.NewScreenerRun(criteria)
	if err := s.repo.CreateScreenerRun(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create screener run: %w", err)
	}

	// Step 1: Fetch candidates from FMP
	screenCriteria := services.ScreenCriteria{
		MarketCapMin: criteria.MarketCapMin,
		PERatioMax:   criteria.PERatioMax,
		PBRatioMax:   criteria.PBRatioMax,
		Limit:        criteria.Limit,
	}

	fmpResults, err := s.fmpService.Screen(ctx, screenCriteria)
	if err != nil {
		durationMs := time.Since(startTime).Milliseconds()
		run.Fail(fmt.Sprintf("failed to fetch candidates: %v", err), durationMs)
		_ = s.repo.UpdateScreenerRun(ctx, run)
		return run, fmt.Errorf("failed to fetch candidates from FMP: %w", err)
	}

	// Convert FMP results to ScreenerCandidates
	candidates := make([]models.ScreenerCandidate, 0, len(fmpResults))
	for _, r := range fmpResults {
		candidates = append(candidates, models.ScreenerCandidate{
			Symbol:        r.Symbol,
			CompanyName:   r.CompanyName,
			MarketCap:     r.MarketCap,
			PERatio:       r.PERatio,
			PBRatio:       r.PBRatio,
			EPS:           r.EPS,
			DividendYield: r.DividendYield,
			Sector:        r.Sector,
			Industry:      r.Industry,
			Price:         r.Price,
			Beta:          r.Beta,
			Analyzed:      false,
		})
	}

	// Step 2: Pre-filter by value score
	preFiltered := RankByValueScore(candidates, s.cfg.PreFilterLimit)
	observability.Info("pre-filtered candidates",
		"total", len(candidates),
		"filtered", len(preFiltered))

	// Step 3: Run full analysis on pre-filtered candidates in parallel
	analyzedCandidates, recommendations := s.analyzeInParallel(ctx, preFiltered)

	// Update candidates with analysis results
	run.SetCandidates(analyzedCandidates)

	// Step 4: Rank by analysis score and get top picks
	topCandidates := RankByAnalysisScore(analyzedCandidates, s.cfg.TopPicksCount)

	// Collect recommendation IDs for top picks
	topPicks := make([]uuid.UUID, 0, len(topCandidates))
	for _, c := range topCandidates {
		// Find the recommendation for this symbol
		for _, rec := range recommendations {
			if rec.Symbol == c.Symbol {
				topPicks = append(topPicks, rec.ID)
				break
			}
		}
	}

	// Complete the run
	durationMs := time.Since(startTime).Milliseconds()
	run.Complete(durationMs, topPicks)

	if err := s.repo.UpdateScreenerRun(ctx, run); err != nil {
		observability.Warn("failed to update screener run", "error", err)
	}

	observability.Info("screener run completed",
		"duration_ms", durationMs,
		"candidates", len(analyzedCandidates),
		"top_picks", len(topPicks))

	return run, nil
}

// analyzeInParallel runs analysis on candidates concurrently with a semaphore limit
func (s *ValueScreener) analyzeInParallel(ctx context.Context, candidates []models.ScreenerCandidate) ([]models.ScreenerCandidate, []*models.Recommendation) {
	// Create timeout context for analysis
	analysisCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.AnalysisTimeoutSec)*time.Second)
	defer cancel()

	// Results channels
	type analysisResult struct {
		index      int
		candidate  models.ScreenerCandidate
		recommendation *models.Recommendation
		err        error
	}

	results := make(chan analysisResult, len(candidates))
	sem := make(chan struct{}, s.cfg.MaxConcurrent)
	var wg sync.WaitGroup

	// Launch analysis goroutines
	for i, candidate := range candidates {
		wg.Add(1)
		go func(idx int, c models.ScreenerCandidate) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-analysisCtx.Done():
				results <- analysisResult{index: idx, candidate: c, err: analysisCtx.Err()}
				return
			}

			// Run analysis
			rec, err := s.analysisProvider.AnalyzeSymbol(analysisCtx, c.Symbol)
			if err != nil || rec == nil {
				observability.Warn("analysis failed for candidate",
					"symbol", c.Symbol,
					"error", err)
				results <- analysisResult{index: idx, candidate: c, err: err}
				return
			}

			// Calculate combined score (weighted average of all agent scores)
			combinedScore := (rec.FundamentalScore*0.4 + rec.SentimentScore*0.3 + rec.TechnicalScore*0.3)
			confidence := rec.Confidence

			// Update candidate with analysis results
			c.Score = &combinedScore
			c.Confidence = &confidence
			c.Analyzed = true

			// Save recommendation to database
			if err := s.repo.CreateRecommendation(analysisCtx, rec); err != nil {
				observability.Warn("failed to save recommendation",
					"symbol", c.Symbol,
					"error", err)
			}

			results <- analysisResult{index: idx, candidate: c, recommendation: rec, err: nil}
		}(i, candidate)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	analyzedCandidates := make([]models.ScreenerCandidate, len(candidates))
	copy(analyzedCandidates, candidates)
	recommendations := make([]*models.Recommendation, 0, len(candidates))

	for result := range results {
		if result.err == nil {
			analyzedCandidates[result.index] = result.candidate
			if result.recommendation != nil {
				recommendations = append(recommendations, result.recommendation)
			}
		}
	}

	return analyzedCandidates, recommendations
}

// GetLatestPicks returns the top picks from the most recent completed screener run
func (s *ValueScreener) GetLatestPicks(ctx context.Context) ([]models.ScreenerCandidate, error) {
	run, err := s.repo.GetLatestScreenerRun(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest screener run: %w", err)
	}
	if run == nil {
		return nil, nil
	}
	if !run.IsCompleted() {
		return nil, nil
	}

	// Get the top candidates from the run
	return RankByAnalysisScore(run.Candidates, s.cfg.TopPicksCount), nil
}

// GetLatestRun returns the most recent screener run
func (s *ValueScreener) GetLatestRun(ctx context.Context) (*models.ScreenerRun, error) {
	return s.repo.GetLatestScreenerRun(ctx)
}

// GetRunHistory returns the history of screener runs
func (s *ValueScreener) GetRunHistory(ctx context.Context, limit int) ([]models.ScreenerRun, error) {
	return s.repo.GetScreenerRunHistory(ctx, limit)
}

// GetRun returns a specific screener run by ID
func (s *ValueScreener) GetRun(ctx context.Context, id uuid.UUID) (*models.ScreenerRun, error) {
	return s.repo.GetScreenerRun(ctx, id)
}
