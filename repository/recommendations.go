package repository

import (
	"context"
	"fmt"
	"time"

	"trade-machine/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetRecommendations returns recommendations filtered by status
func (r *Repository) GetRecommendations(ctx context.Context, status models.RecommendationStatus, limit int) ([]models.Recommendation, error) {
	if limit <= 0 {
		limit = 50
	}

	var rows pgx.Rows
	var err error

	if status == "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, symbol, action, quantity, target_price, confidence, reasoning,
				   fundamental_score, sentiment_score, technical_score,
				   status, approved_at, rejected_at, executed_trade_id, created_at
			FROM recommendations
			ORDER BY created_at DESC
			LIMIT $1
		`, limit)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, symbol, action, quantity, target_price, confidence, reasoning,
				   fundamental_score, sentiment_score, technical_score,
				   status, approved_at, rejected_at, executed_trade_id, created_at
			FROM recommendations
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2
		`, status, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query recommendations: %w", err)
	}
	defer rows.Close()

	var recs []models.Recommendation
	for rows.Next() {
		var rec models.Recommendation
		err := rows.Scan(&rec.ID, &rec.Symbol, &rec.Action, &rec.Quantity, &rec.TargetPrice, &rec.Confidence, &rec.Reasoning,
			&rec.FundamentalScore, &rec.SentimentScore, &rec.TechnicalScore,
			&rec.Status, &rec.ApprovedAt, &rec.RejectedAt, &rec.ExecutedTradeID, &rec.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recommendation: %w", err)
		}
		recs = append(recs, rec)
	}

	return recs, nil
}

// GetRecommendation returns a single recommendation by ID
func (r *Repository) GetRecommendation(ctx context.Context, id uuid.UUID) (*models.Recommendation, error) {
	var rec models.Recommendation
	err := r.pool.QueryRow(ctx, `
		SELECT id, symbol, action, quantity, target_price, confidence, reasoning,
			   fundamental_score, sentiment_score, technical_score,
			   status, approved_at, rejected_at, executed_trade_id, created_at
		FROM recommendations WHERE id = $1
	`, id).Scan(&rec.ID, &rec.Symbol, &rec.Action, &rec.Quantity, &rec.TargetPrice, &rec.Confidence, &rec.Reasoning,
		&rec.FundamentalScore, &rec.SentimentScore, &rec.TechnicalScore,
		&rec.Status, &rec.ApprovedAt, &rec.RejectedAt, &rec.ExecutedTradeID, &rec.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query recommendation: %w", err)
	}

	return &rec, nil
}

// CreateRecommendation creates a new recommendation
func (r *Repository) CreateRecommendation(ctx context.Context, rec *models.Recommendation) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO recommendations (id, symbol, action, quantity, target_price, confidence, reasoning,
			fundamental_score, sentiment_score, technical_score, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, rec.ID, rec.Symbol, rec.Action, rec.Quantity, rec.TargetPrice, rec.Confidence, rec.Reasoning,
		rec.FundamentalScore, rec.SentimentScore, rec.TechnicalScore, rec.Status, rec.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create recommendation: %w", err)
	}

	return nil
}

// ApproveRecommendation marks a recommendation as approved
func (r *Repository) ApproveRecommendation(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE recommendations 
		SET status = $2, approved_at = $3 
		WHERE id = $1
	`, id, models.RecommendationStatusApproved, time.Now())

	if err != nil {
		return fmt.Errorf("failed to approve recommendation: %w", err)
	}

	return nil
}

// RejectRecommendation marks a recommendation as rejected
func (r *Repository) RejectRecommendation(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE recommendations 
		SET status = $2, rejected_at = $3 
		WHERE id = $1
	`, id, models.RecommendationStatusRejected, time.Now())

	if err != nil {
		return fmt.Errorf("failed to reject recommendation: %w", err)
	}

	return nil
}

// ExecuteRecommendation marks a recommendation as executed with the trade ID
func (r *Repository) ExecuteRecommendation(ctx context.Context, id uuid.UUID, tradeID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE recommendations 
		SET status = $2, executed_trade_id = $3 
		WHERE id = $1
	`, id, models.RecommendationStatusExecuted, tradeID)

	if err != nil {
		return fmt.Errorf("failed to execute recommendation: %w", err)
	}

	return nil
}

// GetPendingRecommendations returns all pending recommendations
func (r *Repository) GetPendingRecommendations(ctx context.Context) ([]models.Recommendation, error) {
	return r.GetRecommendations(ctx, models.RecommendationStatusPending, 100)
}
