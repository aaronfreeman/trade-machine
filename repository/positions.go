package repository

import (
	"context"
	"fmt"

	"trade-machine/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetPositions returns all positions
func (r *Repository) GetPositions(ctx context.Context) ([]models.Position, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, symbol, quantity, avg_entry_price, current_price, unrealized_pl, side, created_at, updated_at
		FROM positions
		ORDER BY symbol
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	var positions []models.Position
	for rows.Next() {
		var p models.Position
		err := rows.Scan(&p.ID, &p.Symbol, &p.Quantity, &p.AvgEntryPrice, &p.CurrentPrice, &p.UnrealizedPL, &p.Side, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, p)
	}

	return positions, nil
}

// GetPosition returns a single position by ID
func (r *Repository) GetPosition(ctx context.Context, id uuid.UUID) (*models.Position, error) {
	var p models.Position
	err := r.pool.QueryRow(ctx, `
		SELECT id, symbol, quantity, avg_entry_price, current_price, unrealized_pl, side, created_at, updated_at
		FROM positions WHERE id = $1
	`, id).Scan(&p.ID, &p.Symbol, &p.Quantity, &p.AvgEntryPrice, &p.CurrentPrice, &p.UnrealizedPL, &p.Side, &p.CreatedAt, &p.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query position: %w", err)
	}

	return &p, nil
}

// GetPositionBySymbol returns a position by symbol
func (r *Repository) GetPositionBySymbol(ctx context.Context, symbol string) (*models.Position, error) {
	var p models.Position
	err := r.pool.QueryRow(ctx, `
		SELECT id, symbol, quantity, avg_entry_price, current_price, unrealized_pl, side, created_at, updated_at
		FROM positions WHERE symbol = $1
	`, symbol).Scan(&p.ID, &p.Symbol, &p.Quantity, &p.AvgEntryPrice, &p.CurrentPrice, &p.UnrealizedPL, &p.Side, &p.CreatedAt, &p.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query position: %w", err)
	}

	return &p, nil
}

// CreatePosition creates a new position
func (r *Repository) CreatePosition(ctx context.Context, pos *models.Position) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO positions (id, symbol, quantity, avg_entry_price, current_price, unrealized_pl, side, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, pos.ID, pos.Symbol, pos.Quantity, pos.AvgEntryPrice, pos.CurrentPrice, pos.UnrealizedPL, pos.Side, pos.CreatedAt, pos.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create position: %w", err)
	}

	return nil
}

// UpdatePosition updates an existing position
func (r *Repository) UpdatePosition(ctx context.Context, pos *models.Position) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE positions 
		SET quantity = $2, avg_entry_price = $3, current_price = $4, unrealized_pl = $5, side = $6, updated_at = NOW()
		WHERE id = $1
	`, pos.ID, pos.Quantity, pos.AvgEntryPrice, pos.CurrentPrice, pos.UnrealizedPL, pos.Side)

	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	return nil
}

// DeletePosition removes a position
func (r *Repository) DeletePosition(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM positions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete position: %w", err)
	}
	return nil
}
