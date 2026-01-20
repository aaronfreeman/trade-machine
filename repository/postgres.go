package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNoDatabaseConnection is returned when a database operation is attempted
// but the repository has no active database connection.
var ErrNoDatabaseConnection = errors.New("no database connection available")

// DBTX is an interface that both pgxpool.Pool and pgx.Tx satisfy.
// This allows Repository methods to work with either a connection pool
// or a transaction.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Repository provides database access for all entities
type Repository struct {
	pool *pgxpool.Pool
	db   DBTX // The actual executor (pool or transaction)
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

	return &Repository{pool: pool, db: pool}, nil
}

// WithTx returns a new Repository that uses the given transaction.
// This is useful for running multiple operations atomically.
func (r *Repository) WithTx(tx pgx.Tx) *Repository {
	return &Repository{pool: r.pool, db: tx}
}

// BeginTx starts a new transaction and returns a Repository that uses it.
// The caller is responsible for calling Commit() or Rollback() on the transaction.
func (r *Repository) BeginTx(ctx context.Context) (pgx.Tx, *Repository, error) {
	if r.pool == nil {
		return nil, nil, ErrNoDatabaseConnection
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, r.WithTx(tx), nil
}

// Close closes the database connection pool
func (r *Repository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

// Health checks if the database connection is healthy
func (r *Repository) Health(ctx context.Context) error {
	if r.pool == nil {
		return ErrNoDatabaseConnection
	}
	return r.pool.Ping(ctx)
}

// Pool returns the underlying connection pool for advanced operations.
// This is primarily intended for testing and cleanup operations.
func (r *Repository) Pool() *pgxpool.Pool {
	return r.pool
}

// checkDB returns an error if the database connection is not available.
func (r *Repository) checkDB() error {
	if r.db == nil {
		return ErrNoDatabaseConnection
	}
	return nil
}
