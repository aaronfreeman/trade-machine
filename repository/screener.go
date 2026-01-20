package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"trade-machine/models"
	"trade-machine/observability"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreateScreenerRun creates a new screener run
func (r *Repository) CreateScreenerRun(ctx context.Context, run *models.ScreenerRun) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	metrics := observability.GetMetrics()
	timer := metrics.NewTimer()
	defer timer.ObserveDB("insert", "screener_runs")

	criteriaJSON, err := json.Marshal(run.Criteria)
	if err != nil {
		return fmt.Errorf("failed to marshal criteria: %w", err)
	}

	candidatesJSON, err := json.Marshal(run.Candidates)
	if err != nil {
		return fmt.Errorf("failed to marshal candidates: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO screener_runs (id, run_at, criteria, candidates, top_picks, duration_ms, status, error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, run.ID, run.RunAt, criteriaJSON, candidatesJSON, run.TopPicks, run.DurationMs, run.Status, run.Error, run.CreatedAt)

	if err != nil {
		metrics.RecordDBError("insert", "screener_runs")
		return fmt.Errorf("failed to create screener run: %w", err)
	}

	return nil
}

// UpdateScreenerRun updates an existing screener run
func (r *Repository) UpdateScreenerRun(ctx context.Context, run *models.ScreenerRun) error {
	if err := r.checkDB(); err != nil {
		return err
	}
	metrics := observability.GetMetrics()
	timer := metrics.NewTimer()
	defer timer.ObserveDB("update", "screener_runs")

	candidatesJSON, err := json.Marshal(run.Candidates)
	if err != nil {
		return fmt.Errorf("failed to marshal candidates: %w", err)
	}

	_, err = r.db.Exec(ctx, `
		UPDATE screener_runs
		SET candidates = $2, top_picks = $3, duration_ms = $4, status = $5, error = $6
		WHERE id = $1
	`, run.ID, candidatesJSON, run.TopPicks, run.DurationMs, run.Status, run.Error)

	if err != nil {
		metrics.RecordDBError("update", "screener_runs")
		return fmt.Errorf("failed to update screener run: %w", err)
	}

	return nil
}

// GetScreenerRun returns a screener run by ID
func (r *Repository) GetScreenerRun(ctx context.Context, id uuid.UUID) (*models.ScreenerRun, error) {
	if err := r.checkDB(); err != nil {
		return nil, err
	}
	metrics := observability.GetMetrics()
	timer := metrics.NewTimer()
	defer timer.ObserveDB("select", "screener_runs")

	var run models.ScreenerRun
	var criteriaJSON, candidatesJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, run_at, criteria, candidates, top_picks, duration_ms, status, error, created_at
		FROM screener_runs
		WHERE id = $1
	`, id).Scan(&run.ID, &run.RunAt, &criteriaJSON, &candidatesJSON, &run.TopPicks, &run.DurationMs, &run.Status, &run.Error, &run.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		metrics.RecordDBError("select", "screener_runs")
		return nil, fmt.Errorf("failed to get screener run: %w", err)
	}

	if err := json.Unmarshal(criteriaJSON, &run.Criteria); err != nil {
		return nil, fmt.Errorf("failed to unmarshal criteria: %w", err)
	}

	if err := json.Unmarshal(candidatesJSON, &run.Candidates); err != nil {
		return nil, fmt.Errorf("failed to unmarshal candidates: %w", err)
	}

	return &run, nil
}

// GetLatestScreenerRun returns the most recent screener run
func (r *Repository) GetLatestScreenerRun(ctx context.Context) (*models.ScreenerRun, error) {
	if err := r.checkDB(); err != nil {
		return nil, err
	}
	metrics := observability.GetMetrics()
	timer := metrics.NewTimer()
	defer timer.ObserveDB("select", "screener_runs")

	var run models.ScreenerRun
	var criteriaJSON, candidatesJSON []byte

	err := r.db.QueryRow(ctx, `
		SELECT id, run_at, criteria, candidates, top_picks, duration_ms, status, error, created_at
		FROM screener_runs
		ORDER BY run_at DESC
		LIMIT 1
	`).Scan(&run.ID, &run.RunAt, &criteriaJSON, &candidatesJSON, &run.TopPicks, &run.DurationMs, &run.Status, &run.Error, &run.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		metrics.RecordDBError("select", "screener_runs")
		return nil, fmt.Errorf("failed to get latest screener run: %w", err)
	}

	if err := json.Unmarshal(criteriaJSON, &run.Criteria); err != nil {
		return nil, fmt.Errorf("failed to unmarshal criteria: %w", err)
	}

	if err := json.Unmarshal(candidatesJSON, &run.Candidates); err != nil {
		return nil, fmt.Errorf("failed to unmarshal candidates: %w", err)
	}

	return &run, nil
}

// GetScreenerRunHistory returns a list of recent screener runs (summary only)
func (r *Repository) GetScreenerRunHistory(ctx context.Context, limit int) ([]models.ScreenerRun, error) {
	if err := r.checkDB(); err != nil {
		return nil, err
	}
	metrics := observability.GetMetrics()
	timer := metrics.NewTimer()
	defer timer.ObserveDB("select", "screener_runs")

	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, run_at, criteria, candidates, top_picks, duration_ms, status, error, created_at
		FROM screener_runs
		ORDER BY run_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		metrics.RecordDBError("select", "screener_runs")
		return nil, fmt.Errorf("failed to get screener run history: %w", err)
	}
	defer rows.Close()

	var runs []models.ScreenerRun
	for rows.Next() {
		var run models.ScreenerRun
		var criteriaJSON, candidatesJSON []byte

		err := rows.Scan(&run.ID, &run.RunAt, &criteriaJSON, &candidatesJSON, &run.TopPicks, &run.DurationMs, &run.Status, &run.Error, &run.CreatedAt)
		if err != nil {
			metrics.RecordDBError("select", "screener_runs")
			return nil, fmt.Errorf("failed to scan screener run: %w", err)
		}

		if err := json.Unmarshal(criteriaJSON, &run.Criteria); err != nil {
			return nil, fmt.Errorf("failed to unmarshal criteria: %w", err)
		}

		if err := json.Unmarshal(candidatesJSON, &run.Candidates); err != nil {
			return nil, fmt.Errorf("failed to unmarshal candidates: %w", err)
		}

		runs = append(runs, run)
	}

	return runs, nil
}
