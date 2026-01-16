package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// GetCachedData retrieves cached data for a symbol and data type
func (r *Repository) GetCachedData(ctx context.Context, symbol, dataType string) (map[string]interface{}, error) {
	var data []byte

	// Let the database handle expiry check to avoid timezone issues
	err := r.pool.QueryRow(ctx, `
		SELECT data FROM market_data_cache
		WHERE symbol = $1 AND data_type = $2 AND expires_at > NOW()
	`, symbol, dataType).Scan(&data)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query cache: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return result, nil
}

// SetCachedData stores data in the cache with a TTL
func (r *Repository) SetCachedData(ctx context.Context, symbol, dataType string, data map[string]interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO market_data_cache (symbol, data_type, data, expires_at)
		VALUES ($1, $2, $3, NOW() + $4::interval)
		ON CONFLICT (symbol, data_type) 
		DO UPDATE SET data = EXCLUDED.data, expires_at = NOW() + $4::interval, created_at = NOW()
	`, symbol, dataType, jsonData, ttl.String())

	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// InvalidateCache removes cached data for a symbol and data type
func (r *Repository) InvalidateCache(ctx context.Context, symbol, dataType string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM market_data_cache WHERE symbol = $1 AND data_type = $2
	`, symbol, dataType)

	if err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	return nil
}

// InvalidateAllCacheForSymbol removes all cached data for a symbol
func (r *Repository) InvalidateAllCacheForSymbol(ctx context.Context, symbol string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM market_data_cache WHERE symbol = $1`, symbol)
	if err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}
	return nil
}

// CleanExpiredCache removes all expired cache entries
func (r *Repository) CleanExpiredCache(ctx context.Context) (int64, error) {
	result, err := r.pool.Exec(ctx, `DELETE FROM market_data_cache WHERE expires_at < NOW()`)
	if err != nil {
		return 0, fmt.Errorf("failed to clean expired cache: %w", err)
	}
	return result.RowsAffected(), nil
}
