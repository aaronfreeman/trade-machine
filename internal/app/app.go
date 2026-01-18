package app

import (
	"context"
	"fmt"

	"trade-machine/config"
	"trade-machine/models"
	"trade-machine/services"

	"github.com/google/uuid"
)

// RepositoryInterface defines the repository operations needed by App
type RepositoryInterface interface {
	Close()
	Health(ctx context.Context) error
	GetRecommendations(ctx context.Context, status models.RecommendationStatus, limit int) ([]models.Recommendation, error)
	GetRecommendation(ctx context.Context, id uuid.UUID) (*models.Recommendation, error)
	GetPendingRecommendations(ctx context.Context) ([]models.Recommendation, error)
	ApproveRecommendation(ctx context.Context, id uuid.UUID) error
	RejectRecommendation(ctx context.Context, id uuid.UUID) error
	GetPositions(ctx context.Context) ([]models.Position, error)
	GetTrades(ctx context.Context, limit int) ([]models.Trade, error)
	GetAgentRuns(ctx context.Context, agentType models.AgentType, limit int) ([]models.AgentRun, error)
}

// PortfolioManagerInterface defines the analysis operations
type PortfolioManagerInterface interface {
	AnalyzeSymbol(ctx context.Context, symbol string) (*models.Recommendation, error)
}

// App struct holds application dependencies using interfaces for testability
type App struct {
	ctx              context.Context
	cfg              *config.Config
	repo             RepositoryInterface
	portfolioManager PortfolioManagerInterface
	alpacaService    services.AlpacaServiceInterface
	analysisSem      chan struct{}
}

// New creates a new App application struct
func New(cfg *config.Config, repo RepositoryInterface, manager PortfolioManagerInterface, alpaca services.AlpacaServiceInterface) *App {
	return &App{
		cfg:              cfg,
		repo:             repo,
		portfolioManager: manager,
		alpacaService:    alpaca,
		analysisSem:      make(chan struct{}, cfg.Agent.ConcurrencyLimit),
	}
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Shutdown is called when the app is closing
func (a *App) Shutdown(ctx context.Context) {
	if a.repo != nil {
		a.repo.Close()
	}
}

// Repo returns the repository interface for API handlers
func (a *App) Repo() RepositoryInterface {
	return a.repo
}

// AnalyzeStock runs all agents to analyze a stock and generate a recommendation
func (a *App) AnalyzeStock(symbol string) (*models.Recommendation, error) {
	if a.portfolioManager == nil {
		return nil, fmt.Errorf("portfolio manager not initialized")
	}

	select {
	case a.analysisSem <- struct{}{}:
		defer func() { <-a.analysisSem }()
	default:
		return nil, fmt.Errorf("analysis queue full, too many concurrent requests - try again later")
	}

	return a.portfolioManager.AnalyzeSymbol(a.ctx, symbol)
}

// GetRecommendations returns recent recommendations
func (a *App) GetRecommendations(limit int) ([]models.Recommendation, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.repo.GetRecommendations(a.ctx, "", limit)
}

// GetPendingRecommendations returns pending recommendations awaiting approval
func (a *App) GetPendingRecommendations() ([]models.Recommendation, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.repo.GetPendingRecommendations(a.ctx)
}

// ApproveRecommendation approves a recommendation for execution
func (a *App) ApproveRecommendation(id string) error {
	if a.repo == nil {
		return fmt.Errorf("database not initialized")
	}

	uuid, err := ParseUUID(id)
	if err != nil {
		return err
	}

	return a.repo.ApproveRecommendation(a.ctx, uuid)
}

// RejectRecommendation rejects a recommendation
func (a *App) RejectRecommendation(id string) error {
	if a.repo == nil {
		return fmt.Errorf("database not initialized")
	}

	uuid, err := ParseUUID(id)
	if err != nil {
		return err
	}

	return a.repo.RejectRecommendation(a.ctx, uuid)
}

// GetRecommendationByID returns a single recommendation by ID
func (a *App) GetRecommendationByID(id string) (*models.Recommendation, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	uuid, err := ParseUUID(id)
	if err != nil {
		return nil, err
	}

	return a.repo.GetRecommendation(a.ctx, uuid)
}

// GetPositions returns all current positions
func (a *App) GetPositions() ([]models.Position, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.repo.GetPositions(a.ctx)
}

// GetTrades returns recent trades
func (a *App) GetTrades(limit int) ([]models.Trade, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.repo.GetTrades(a.ctx, limit)
}

// GetAgentRuns returns recent agent runs
func (a *App) GetAgentRuns(limit int) ([]models.AgentRun, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return a.repo.GetAgentRuns(a.ctx, "", limit)
}

// ParseUUID parses a string UUID into a [16]byte
func ParseUUID(id string) ([16]byte, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return [16]byte{}, fmt.Errorf("invalid UUID: %w", err)
	}
	return parsed, nil
}

// AnalysisSemCapacity returns the capacity of the analysis semaphore (for testing)
func (a *App) AnalysisSemCapacity() int {
	return cap(a.analysisSem)
}
