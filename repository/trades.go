package repository

import (
	"context"
	"fmt"

	"trade-machine/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetTrades returns trades with optional limit
func (r *Repository) GetTrades(ctx context.Context, limit int) ([]models.Trade, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, symbol, side, quantity, price, total_value, commission, status, alpaca_order_id, executed_at, created_at
		FROM trades
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query trades: %w", err)
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var t models.Trade
		err := rows.Scan(&t.ID, &t.Symbol, &t.Side, &t.Quantity, &t.Price, &t.TotalValue, &t.Commission, &t.Status, &t.AlpacaOrderID, &t.ExecutedAt, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, t)
	}

	return trades, nil
}

// GetTrade returns a single trade by ID
func (r *Repository) GetTrade(ctx context.Context, id uuid.UUID) (*models.Trade, error) {
	var t models.Trade
	err := r.db.QueryRow(ctx, `
		SELECT id, symbol, side, quantity, price, total_value, commission, status, alpaca_order_id, executed_at, created_at
		FROM trades WHERE id = $1
	`, id).Scan(&t.ID, &t.Symbol, &t.Side, &t.Quantity, &t.Price, &t.TotalValue, &t.Commission, &t.Status, &t.AlpacaOrderID, &t.ExecutedAt, &t.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query trade: %w", err)
	}

	return &t, nil
}

// CreateTrade creates a new trade record
func (r *Repository) CreateTrade(ctx context.Context, trade *models.Trade) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO trades (id, symbol, side, quantity, price, total_value, commission, status, alpaca_order_id, executed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, trade.ID, trade.Symbol, trade.Side, trade.Quantity, trade.Price, trade.TotalValue, trade.Commission, trade.Status, trade.AlpacaOrderID, trade.ExecutedAt, trade.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create trade: %w", err)
	}

	return nil
}

// UpdateTradeStatus updates the status of a trade
func (r *Repository) UpdateTradeStatus(ctx context.Context, id uuid.UUID, status models.TradeStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE trades SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("failed to update trade status: %w", err)
	}
	return nil
}

// GetTradesBySymbol returns trades for a specific symbol
func (r *Repository) GetTradesBySymbol(ctx context.Context, symbol string, limit int) ([]models.Trade, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, symbol, side, quantity, price, total_value, commission, status, alpaca_order_id, executed_at, created_at
		FROM trades
		WHERE symbol = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query trades: %w", err)
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var t models.Trade
		err := rows.Scan(&t.ID, &t.Symbol, &t.Side, &t.Quantity, &t.Price, &t.TotalValue, &t.Commission, &t.Status, &t.AlpacaOrderID, &t.ExecutedAt, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, t)
	}

	return trades, nil
}
