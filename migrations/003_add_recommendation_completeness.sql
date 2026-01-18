-- +goose Up
-- Add data completeness tracking to recommendations
ALTER TABLE recommendations
ADD COLUMN data_completeness DECIMAL(5,2) DEFAULT 100.0 CHECK (data_completeness >= 0 AND data_completeness <= 100),
ADD COLUMN missing_agents JSONB DEFAULT '[]';

COMMENT ON COLUMN recommendations.data_completeness IS 'Percentage of agents that successfully provided analysis (0-100)';
COMMENT ON COLUMN recommendations.missing_agents IS 'JSON array of agents that were unavailable or failed';

-- +goose Down
ALTER TABLE recommendations
DROP COLUMN IF EXISTS data_completeness,
DROP COLUMN IF EXISTS missing_agents;
