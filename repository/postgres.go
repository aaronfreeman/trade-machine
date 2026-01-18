package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides database access for all entities
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new Repository with a PostgreSQL connection pool
func NewRepository(ctx context.Context, connString string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Repository{pool: pool}, nil
}

// Close closes the database connection pool
func (r *Repository) Close() {
	r.pool.Close()
}

// Health checks if the database connection is healthy
func (r *Repository) Health(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// Pool returns the underlying connection pool for advanced operations.
// This is primarily intended for testing and cleanup operations.
func (r *Repository) Pool() *pgxpool.Pool {
	return r.pool
}
