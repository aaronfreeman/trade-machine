package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"trade-machine/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreateAgentRun creates a new agent run record
func (r *Repository) CreateAgentRun(ctx context.Context, run *models.AgentRun) error {
	inputData, _ := json.Marshal(run.InputData)

	_, err := r.db.Exec(ctx, `
		INSERT INTO agent_runs (id, agent_type, symbol, status, input_data, started_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, run.ID, run.AgentType, run.Symbol, run.Status, inputData, run.StartedAt)

	if err != nil {
		return fmt.Errorf("failed to create agent run: %w", err)
	}

	return nil
}

// UpdateAgentRun updates an existing agent run
func (r *Repository) UpdateAgentRun(ctx context.Context, run *models.AgentRun) error {
	outputData, _ := json.Marshal(run.OutputData)

	_, err := r.db.Exec(ctx, `
		UPDATE agent_runs 
		SET status = $2, output_data = $3, error_message = $4, duration_ms = $5, completed_at = $6
		WHERE id = $1
	`, run.ID, run.Status, outputData, run.ErrorMessage, run.DurationMs, run.CompletedAt)

	if err != nil {
		return fmt.Errorf("failed to update agent run: %w", err)
	}

	return nil
}

// GetAgentRun returns a single agent run by ID
func (r *Repository) GetAgentRun(ctx context.Context, id uuid.UUID) (*models.AgentRun, error) {
	var run models.AgentRun
	var inputData, outputData []byte
	var errorMessage *string
	var durationMs *int

	err := r.db.QueryRow(ctx, `
		SELECT id, agent_type, symbol, status, input_data, output_data, error_message, duration_ms, started_at, completed_at
		FROM agent_runs WHERE id = $1
	`, id).Scan(&run.ID, &run.AgentType, &run.Symbol, &run.Status, &inputData, &outputData, &errorMessage, &durationMs, &run.StartedAt, &run.CompletedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query agent run: %w", err)
	}

	if errorMessage != nil {
		run.ErrorMessage = *errorMessage
	}
	if durationMs != nil {
		run.DurationMs = *durationMs
	}
	if inputData != nil {
		json.Unmarshal(inputData, &run.InputData)
	}
	if outputData != nil {
		json.Unmarshal(outputData, &run.OutputData)
	}

	return &run, nil
}

// GetAgentRuns returns agent runs with optional filtering by agent type
func (r *Repository) GetAgentRuns(ctx context.Context, agentType models.AgentType, limit int) ([]models.AgentRun, error) {
	if limit <= 0 {
		limit = 50
	}

	var rows pgx.Rows
	var err error

	if agentType == "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, agent_type, symbol, status, input_data, output_data, error_message, duration_ms, started_at, completed_at
			FROM agent_runs
			ORDER BY started_at DESC
			LIMIT $1
		`, limit)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, agent_type, symbol, status, input_data, output_data, error_message, duration_ms, started_at, completed_at
			FROM agent_runs
			WHERE agent_type = $1
			ORDER BY started_at DESC
			LIMIT $2
		`, agentType, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query agent runs: %w", err)
	}
	defer rows.Close()

	var runs []models.AgentRun
	for rows.Next() {
		var run models.AgentRun
		var inputData, outputData []byte
		var errorMessage *string
		var durationMs *int

		err := rows.Scan(&run.ID, &run.AgentType, &run.Symbol, &run.Status, &inputData, &outputData, &errorMessage, &durationMs, &run.StartedAt, &run.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent run: %w", err)
		}

		if errorMessage != nil {
			run.ErrorMessage = *errorMessage
		}
		if durationMs != nil {
			run.DurationMs = *durationMs
		}
		if inputData != nil {
			json.Unmarshal(inputData, &run.InputData)
		}
		if outputData != nil {
			json.Unmarshal(outputData, &run.OutputData)
		}

		runs = append(runs, run)
	}

	return runs, nil
}

// GetRecentRunsForSymbol returns recent agent runs for a specific symbol
func (r *Repository) GetRecentRunsForSymbol(ctx context.Context, symbol string, limit int) ([]models.AgentRun, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, agent_type, symbol, status, input_data, output_data, error_message, duration_ms, started_at, completed_at
		FROM agent_runs
		WHERE symbol = $1
		ORDER BY started_at DESC
		LIMIT $2
	`, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent runs: %w", err)
	}
	defer rows.Close()

	var runs []models.AgentRun
	for rows.Next() {
		var run models.AgentRun
		var inputData, outputData []byte
		var errorMessage *string
		var durationMs *int

		err := rows.Scan(&run.ID, &run.AgentType, &run.Symbol, &run.Status, &inputData, &outputData, &errorMessage, &durationMs, &run.StartedAt, &run.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent run: %w", err)
		}

		if errorMessage != nil {
			run.ErrorMessage = *errorMessage
		}
		if durationMs != nil {
			run.DurationMs = *durationMs
		}
		if inputData != nil {
			json.Unmarshal(inputData, &run.InputData)
		}
		if outputData != nil {
			json.Unmarshal(outputData, &run.OutputData)
		}

		runs = append(runs, run)
	}

	return runs, nil
}
