-- +goose Up
-- Portfolio holdings
CREATE TABLE positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(10) NOT NULL,
    quantity DECIMAL(20,8) NOT NULL,
    avg_entry_price DECIMAL(20,8) NOT NULL,
    current_price DECIMAL(20,8),
    unrealized_pl DECIMAL(20,8),
    side VARCHAR(10) NOT NULL CHECK (side IN ('long', 'short')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Trade history
CREATE TABLE trades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(10) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    quantity DECIMAL(20,8) NOT NULL,
    price DECIMAL(20,8) NOT NULL,
    total_value DECIMAL(20,8) NOT NULL,
    commission DECIMAL(20,8) DEFAULT 0,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'executed', 'rejected', 'cancelled')),
    alpaca_order_id VARCHAR(50),
    executed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Agent recommendations (Level 1 control - user approves)
CREATE TABLE recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(10) NOT NULL,
    action VARCHAR(10) NOT NULL CHECK (action IN ('buy', 'sell', 'hold')),
    quantity DECIMAL(20,8),
    target_price DECIMAL(20,8),
    confidence DECIMAL(5,2) CHECK (confidence >= 0 AND confidence <= 100),
    reasoning TEXT NOT NULL,
    fundamental_score DECIMAL(5,2),
    sentiment_score DECIMAL(5,2),
    technical_score DECIMAL(5,2),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'approved', 'rejected', 'executed')),
    approved_at TIMESTAMP,
    rejected_at TIMESTAMP,
    executed_trade_id UUID REFERENCES trades(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Agent execution logs
CREATE TABLE agent_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_type VARCHAR(50) NOT NULL CHECK (agent_type IN ('fundamental', 'news', 'technical', 'manager')),
    symbol VARCHAR(10),
    status VARCHAR(20) NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
    input_data JSONB,
    output_data JSONB,
    error_message TEXT,
    duration_ms INTEGER,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- Market data cache (reduce API calls)
CREATE TABLE market_data_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(10) NOT NULL,
    data_type VARCHAR(20) NOT NULL CHECK (data_type IN ('quote', 'fundamentals', 'news', 'technical')),
    data JSONB NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(symbol, data_type)
);

-- Indexes for performance
CREATE INDEX idx_positions_symbol ON positions(symbol);
CREATE INDEX idx_trades_symbol ON trades(symbol);
CREATE INDEX idx_trades_created_at ON trades(created_at DESC);
CREATE INDEX idx_recommendations_status ON recommendations(status);
CREATE INDEX idx_recommendations_created_at ON recommendations(created_at DESC);
CREATE INDEX idx_agent_runs_agent_type ON agent_runs(agent_type);
CREATE INDEX idx_agent_runs_started_at ON agent_runs(started_at DESC);
CREATE INDEX idx_market_data_cache_symbol_type ON market_data_cache(symbol, data_type);
CREATE INDEX idx_market_data_cache_expires_at ON market_data_cache(expires_at);

-- +goose Down
DROP TABLE IF EXISTS market_data_cache;
DROP TABLE IF EXISTS agent_runs;
DROP TABLE IF EXISTS recommendations;
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS positions;
