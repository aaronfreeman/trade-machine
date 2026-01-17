package repository

import (
	"context"
	"time"

	"trade-machine/models"

	"github.com/google/uuid"
)

// RepositoryInterface defines all repository operations
type RepositoryInterface interface {
	// Health and lifecycle
	Close()
	Health(ctx context.Context) error

	// Recommendations
	GetRecommendations(ctx context.Context, status models.RecommendationStatus, limit int) ([]models.Recommendation, error)
	GetRecommendation(ctx context.Context, id uuid.UUID) (*models.Recommendation, error)
	CreateRecommendation(ctx context.Context, rec *models.Recommendation) error
	ApproveRecommendation(ctx context.Context, id uuid.UUID) error
	RejectRecommendation(ctx context.Context, id uuid.UUID) error
	ExecuteRecommendation(ctx context.Context, id uuid.UUID, tradeID uuid.UUID) error
	GetPendingRecommendations(ctx context.Context) ([]models.Recommendation, error)

	// Positions
	GetPositions(ctx context.Context) ([]models.Position, error)
	GetPosition(ctx context.Context, id uuid.UUID) (*models.Position, error)
	GetPositionBySymbol(ctx context.Context, symbol string) (*models.Position, error)
	CreatePosition(ctx context.Context, pos *models.Position) error
	UpdatePosition(ctx context.Context, pos *models.Position) error
	DeletePosition(ctx context.Context, id uuid.UUID) error

	// Trades
	GetTrades(ctx context.Context, limit int) ([]models.Trade, error)
	GetTrade(ctx context.Context, id uuid.UUID) (*models.Trade, error)
	CreateTrade(ctx context.Context, trade *models.Trade) error
	UpdateTradeStatus(ctx context.Context, id uuid.UUID, status models.TradeStatus) error
	GetTradesBySymbol(ctx context.Context, symbol string, limit int) ([]models.Trade, error)

	// Agent runs
	CreateAgentRun(ctx context.Context, run *models.AgentRun) error
	UpdateAgentRun(ctx context.Context, run *models.AgentRun) error
	GetAgentRun(ctx context.Context, id uuid.UUID) (*models.AgentRun, error)
	GetAgentRuns(ctx context.Context, agentType models.AgentType, limit int) ([]models.AgentRun, error)
	GetRecentRunsForSymbol(ctx context.Context, symbol string, limit int) ([]models.AgentRun, error)

	// Cache
	GetCachedData(ctx context.Context, symbol, dataType string) (map[string]interface{}, error)
	SetCachedData(ctx context.Context, symbol, dataType string, data map[string]interface{}, ttl time.Duration) error
	InvalidateCache(ctx context.Context, symbol, dataType string) error
	InvalidateAllCacheForSymbol(ctx context.Context, symbol string) error
	CleanExpiredCache(ctx context.Context) (int64, error)
}

// Compile-time interface verification
var _ RepositoryInterface = (*Repository)(nil)
