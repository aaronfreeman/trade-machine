package repository

import (
	"context"
	"fmt"

	"trade-machine/internal/settings"

	"github.com/google/uuid"
)

// GetAPIKey retrieves an API key by service name
func (r *Repository) GetAPIKey(ctx context.Context, serviceName string) (*settings.APIKeyModel, error) {
	if err := r.checkDB(); err != nil {
		return nil, err
	}

	query := `
		SELECT id, service_name, api_key_encrypted, api_secret_encrypted, 
		       base_url, region, model_id
		FROM api_keys
		WHERE service_name = $1
	`

	var apiKey settings.APIKeyModel
	err := r.db.QueryRow(ctx, query, serviceName).Scan(
		&apiKey.ID,
		&apiKey.ServiceName,
		&apiKey.APIKeyEncrypted,
		&apiKey.APISecretEncrypted,
		&apiKey.BaseURL,
		&apiKey.Region,
		&apiKey.ModelID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get api key: %w", err)
	}

	return &apiKey, nil
}

// GetAllAPIKeys retrieves all API keys
func (r *Repository) GetAllAPIKeys(ctx context.Context) ([]settings.APIKeyModel, error) {
	if err := r.checkDB(); err != nil {
		return nil, err
	}

	query := `
		SELECT id, service_name, api_key_encrypted, api_secret_encrypted,
		       base_url, region, model_id
		FROM api_keys
		ORDER BY service_name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query api keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []settings.APIKeyModel
	for rows.Next() {
		var apiKey settings.APIKeyModel
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.ServiceName,
			&apiKey.APIKeyEncrypted,
			&apiKey.APISecretEncrypted,
			&apiKey.BaseURL,
			&apiKey.Region,
			&apiKey.ModelID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan api key: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating api keys: %w", err)
	}

	return apiKeys, nil
}

// UpsertAPIKey inserts or updates an API key
func (r *Repository) UpsertAPIKey(ctx context.Context, apiKey *settings.APIKeyModel) error {
	if err := r.checkDB(); err != nil {
		return err
	}

	if apiKey.ID == uuid.Nil {
		apiKey.ID = uuid.New()
	}

	query := `
		INSERT INTO api_keys (id, service_name, api_key_encrypted, api_secret_encrypted, 
		                      base_url, region, model_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT (service_name) 
		DO UPDATE SET 
			api_key_encrypted = EXCLUDED.api_key_encrypted,
			api_secret_encrypted = EXCLUDED.api_secret_encrypted,
			base_url = EXCLUDED.base_url,
			region = EXCLUDED.region,
			model_id = EXCLUDED.model_id,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query,
		apiKey.ID,
		apiKey.ServiceName,
		apiKey.APIKeyEncrypted,
		apiKey.APISecretEncrypted,
		apiKey.BaseURL,
		apiKey.Region,
		apiKey.ModelID,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert api key: %w", err)
	}

	return nil
}

// DeleteAPIKey deletes an API key by service name
func (r *Repository) DeleteAPIKey(ctx context.Context, serviceName string) error {
	if err := r.checkDB(); err != nil {
		return err
	}

	query := `DELETE FROM api_keys WHERE service_name = $1`

	_, err := r.db.Exec(ctx, query, serviceName)
	if err != nil {
		return fmt.Errorf("failed to delete api key: %w", err)
	}

	return nil
}
