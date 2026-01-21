-- +goose Up
-- API keys storage (encrypted)
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(50) NOT NULL UNIQUE,
    api_key_encrypted BYTEA,
    api_secret_encrypted BYTEA,
    base_url VARCHAR(255),
    region VARCHAR(50),
    model_id VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for quick lookups by service name
CREATE INDEX idx_api_keys_service_name ON api_keys(service_name);

-- +goose Down
DROP TABLE IF EXISTS api_keys;
