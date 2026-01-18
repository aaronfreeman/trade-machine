-- +goose Up
-- Screener run history
CREATE TABLE screener_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    criteria JSONB NOT NULL,
    candidates JSONB NOT NULL DEFAULT '[]',
    top_picks UUID[] DEFAULT '{}',
    duration_ms INTEGER,
    status VARCHAR(20) NOT NULL DEFAULT 'running'
        CHECK (status IN ('running', 'completed', 'failed')),
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_screener_runs_run_at ON screener_runs(run_at DESC);
CREATE INDEX idx_screener_runs_status ON screener_runs(status);
CREATE INDEX idx_screener_runs_created_at ON screener_runs(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS screener_runs;
