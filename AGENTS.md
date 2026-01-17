# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds

# AI Agent Architecture

This document describes the agent-based architecture that powers trade-machine's AI-driven stock analysis and recommendation system.

## Overview

The trade-machine system uses a multi-agent architecture where specialized AI agents analyze different aspects of a stock to generate trading recommendations. Each agent provides independent analysis, which are then synthesized by the Portfolio Manager to produce a final recommendation.

### Key Principles

- **Modularity**: Each agent has a single responsibility and can be developed independently
- **Parallelism**: Agents run concurrently to minimize analysis latency
- **Transparency**: Each recommendation includes reasoning from all contributing agents
- **Objectivity**: Agents use Claude for consistent, AI-driven analysis across all data sources

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Input: Stock Symbol                     │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┴───────────┐
         │  Portfolio Manager    │
         │  (Orchestrator)       │
         └───────────┬───────────┘
                     │
        ┌────────────┼────────────┬────────────┐
        │            │            │            │
        ▼            ▼            ▼            ▼
    ┌────────┐  ┌────────┐  ┌────────┐  
    │Funda-  │  │Technical│  │ News  │
    │mental  │  │Analyst │  │Analyst│
    │Analyst │  │        │  │       │
    └────┬───┘  └────┬───┘  └───┬──┘
         │           │          │
    ┌────▼────────────▼──────────▼────┐
    │    Synthesize Analyses           │
    │  • Weight scores (40/30/30)     │
    │  • Apply confidence multipliers │
    │  • Generate combined reasoning  │
    └────┬─────────────────────────────┘
         │
         ▼
    ┌──────────────────────┐
    │  Recommendation      │
    │  Action + Reasoning  │
    └──────────────────────┘
```

## Agent Interface

All agents implement the `Agent` interface:

```go
type Agent interface {
    Analyze(ctx context.Context, symbol string) (*Analysis, error)
    Name() string
    Type() models.AgentType
}
```

### Method Descriptions

- **Analyze**: Performs analysis on the given stock symbol and returns an Analysis result
- **Name**: Returns the human-readable name of the agent
- **Type**: Returns the agent's type identifier (fundamental, technical, news, or manager)

## Analysis Structure

Each agent returns an `Analysis` struct containing:

```go
type Analysis struct {
    Symbol     string                 // Stock ticker symbol
    AgentType  models.AgentType       // Type of analysis (fundamental, technical, news)
    Score      float64                // -100 to +100 (negative=bearish, positive=bullish)
    Confidence float64                // 0 to 100 (strength of conviction)
    Reasoning  string                 // Human-readable explanation
    Data       map[string]interface{} // Agent-specific detailed data
    Timestamp  time.Time              // When analysis was performed
}
```

## Agent Types

### 1. Fundamental Analyst

**Purpose**: Evaluates the company's financial health and intrinsic value

**Input Data**:
- P/E ratio (Price-to-Earnings)
- EPS (Earnings Per Share)
- Market capitalization
- 52-week high/low
- Beta (volatility measure)
- Dividend yield

**Analysis Method**:
1. Fetches fundamental data from Alpha Vantage API
2. Constructs a prompt with financial metrics
3. Sends to Claude via AWS Bedrock for analysis
4. Receives JSON response with score, confidence, and key factors

**Output**:
- Score: Based on valuation, growth, and stability metrics
- Confidence: Strength of the fundamental analysis
- Key Factors: List of influential factors (e.g., "High P/E ratio", "Strong earnings growth")
- Data: Raw fundamentals and key factors list

**Coding**
- Keep code as simple as you can to implement the needed features
- Always write tests for code
- Try to keep the overall code coverage above 80%
- To run tests you can run "just test" it will also show a code coverage amount

**Example Reasoning**:
"Company is trading at 15x P/E with 25% YoY earnings growth and 3% dividend yield. Metrics suggest good value with moderate growth potential."

**Weighting**: 40% of final recommendation (highest weight)

### 2. Technical Analyst

**Purpose**: Analyzes price action and technical indicators for short-term trading signals

**Input Data**:
- Historical price data (100 days)
- RSI (Relative Strength Index, 14-period)
- MACD (Moving Average Convergence Divergence)
- MACD Signal line
- SMA 20 (20-day Simple Moving Average)
- SMA 50 (50-day Simple Moving Average)
- 52-day high/low
- Current price vs moving averages

**Analysis Method**:
1. Fetches 100 days of historical price bars from Alpaca
2. Calculates technical indicators:
   - RSI (Relative Strength Index): Overbought (>70) / Oversold (<30)
   - MACD: Trend and momentum
   - SMA: Support/resistance levels
3. Constructs prompt with indicator values
4. Sends to Claude for technical analysis
5. Receives JSON response with score, confidence, and trading signals

**Output**:
- Score: Based on indicator readings and chart patterns
- Confidence: Quality of technical signals
- Signals: List of specific technical signals (e.g., "RSI oversold", "Bullish MACD crossover")
- Data: All calculated indicators and signal list

**Example Reasoning**:
"Price recently broke above 50-day MA. RSI near 60 (not overbought). MACD just crossed above signal line with positive histogram."

**Weighting**: 30% of final recommendation

**Note**: Requires at least 50 bars of historical data to produce meaningful analysis

### 3. News Sentiment Analyst

**Purpose**: Evaluates market sentiment from recent news coverage

**Input Data**:
- Recent news headlines (up to 15 articles)
- Article descriptions
- Article sources
- Publication dates

**Analysis Method**:
1. Fetches recent news articles from NewsAPI
2. Formats articles into a readable prompt
3. Sends to Claude for sentiment analysis
4. Receives JSON response with score, confidence, themes, and notable articles

**Output**:
- Score: Based on sentiment (positive news = bullish, negative news = bearish)
- Confidence: Strength and consistency of sentiment
- Key Themes: List of recurring themes (e.g., "Earnings beat", "Regulatory concerns")
- Notable Articles: Most impactful headlines
- Data: Key themes list and notable articles

**Example Reasoning**:
"Three analyst upgrades this week offset one earnings miss. Overall sentiment is positive with focus on new product announcements."

**Weighting**: 30% of final recommendation

**Note**: Returns neutral score (0) if fewer than 15 articles are available; confidence will be low

## Scoring System

### Score Range: -100 to +100

Scores represent the agent's conviction on a scale where:
- **-100 to -75**: Strongly bearish (sell signal)
- **-75 to -25**: Moderately bearish (lean toward selling)
- **-25 to +25**: Neutral (hold or insufficient data)
- **+25 to +75**: Moderately bullish (lean toward buying)
- **+75 to +100**: Strongly bullish (buy signal)

### Score Normalization

Both scores and confidence are normalized to valid ranges:
- Scores > 100 are clamped to 100
- Scores < -100 are clamped to -100
- Confidence > 100 is clamped to 100
- Confidence < 0 is clamped to 0

### Confidence Range: 0 to 100

Represents the agent's certainty in its analysis:
- **0-20**: Low confidence (insufficient data, ambiguous signals)
- **20-60**: Moderate confidence (clear signals present)
- **60-100**: High confidence (strong, corroborating signals)

## Recommendation Generation

### Weighting Strategy

The Portfolio Manager combines agent analyses using the following weights:
- **Fundamental Analysis: 40%** (highest priority)
- **News Sentiment: 30%**
- **Technical Analysis: 30%**

This weighting prioritizes fundamental value over sentiment and short-term technical signals.

### Score Synthesis Algorithm

```
weighted_score = 0
total_weight = 0

For each agent's analysis:
  weight = agent_weight * (confidence / 100)
  weighted_score += (score * weight)
  total_weight += weight

final_score = weighted_score / total_weight
```

Example calculation:
- Fundamental: Score=50, Confidence=80 → Contribution: 50 * 0.4 * 0.8 = 16
- Technical: Score=30, Confidence=60 → Contribution: 30 * 0.3 * 0.6 = 5.4
- News: Score=40, Confidence=70 → Contribution: 40 * 0.3 * 0.7 = 8.4
- Total weight: 0.4*0.8 + 0.3*0.6 + 0.3*0.7 = 0.76
- Final score: (16 + 5.4 + 8.4) / 0.76 = 38.2

### Action Thresholds

The final score is converted to a trading action:

```go
if score > 25:
    return BUY
if score < -25:
    return SELL
else:
    return HOLD
```

**Action Meanings**:
- **BUY**: Score > 25 (bullish signal, consider acquiring position)
- **SELL**: Score < -25 (bearish signal, consider closing position)
- **HOLD**: Score -25 to +25 (neutral signal, maintain current position)

## Recommendation Structure

The final output from the Portfolio Manager is a `Recommendation` struct:

```go
type Recommendation struct {
    Symbol           string                  // Stock ticker
    Action           models.RecommendationAction  // BUY, SELL, or HOLD
    Quantity         decimal.Decimal         // Suggested number of shares
    TargetPrice      decimal.Decimal         // Target price (future enhancement)
    Confidence       float64                 // 0-100 (average of agent confidences)
    Reasoning        string                  // Combined reasoning from all agents
    FundamentalScore float64                 // Raw fundamental score
    SentimentScore   float64                 // Raw sentiment score
    TechnicalScore   float64                 // Raw technical score
}
```

### Reasoning Format

The combined reasoning includes analysis from all agents:
```
Based on analysis from 3 agents (Fundamental: 50.0, Sentiment: 40.0, Technical: 30.0), overall score is 38.2. 
[Fundamental Analyst] Company is trading at 15x P/E with strong earnings growth...
[Technical Analyst] Price recently broke above 50-day MA with positive MACD...
[News Sentiment Analyst] Recent news shows analyst upgrades with focus on new products...
```

## Parallel Execution

The Portfolio Manager uses goroutines to run all agents concurrently:

1. Agents are executed in parallel using `sync.WaitGroup`
2. Each agent runs independently with its own context
3. Results are collected as agents complete
4. If an agent fails, its analysis is skipped (at least one must succeed)
5. All agent runs are logged to the database for audit/debugging

**Performance**: Total analysis time ≈ slowest agent (typically 2-5 seconds)

## Error Handling

### Agent Failures

If an agent fails to analyze:
- The error is logged
- Analysis from failed agent is skipped
- Recommendation is generated from successful agents
- If all agents fail, an error is returned

### Partial Data

Agents handle incomplete data gracefully:
- **Technical Analyst**: Returns neutral score (0) with low confidence if < 50 bars available
- **News Analyst**: Returns neutral score (0) with low confidence if < 15 articles available
- **Fundamental Analyst**: Returns score 0 with confidence 50 if JSON parsing fails

## Data Flow

1. **User requests analysis** for a stock symbol (e.g., "AAPL")
2. **Portfolio Manager routes** to all registered agents
3. **Each agent independently**:
   - Fetches relevant data from external APIs
   - Prepares analysis prompt
   - Invokes Claude via AWS Bedrock
   - Parses JSON response
   - Returns Analysis with score and reasoning
4. **Portfolio Manager synthesizes**:
   - Extracts individual scores
   - Applies weighting and confidence multipliers
   - Calculates final score
   - Determines action (BUY/SELL/HOLD)
   - Creates combined reasoning
5. **Recommendation is saved** to database and returned

## Integration with External Services

### AWS Bedrock

Provides Claude 3 access for all LLM-based analysis:
- Each agent sends a system prompt + user prompt
- Claude returns structured JSON response
- Allows consistent, repeatable analysis

### Alpaca Markets

Provides market data:
- **Technical Analyst**: Uses historical bars (OHLCV data)
- Provides real-time trading execution capability

### Alpha Vantage

Provides fundamental company data:
- **Fundamental Analyst**: Uses company metrics (P/E, EPS, market cap, beta, dividend yield)

### NewsAPI

Provides news article data:
- **News Sentiment Analyst**: Uses recent articles for sentiment analysis

## Usage Example

```go
// Create agents
fundamental := agents.NewFundamentalAnalyst(bedrockService, alphaVantageService)
technical := agents.NewTechnicalAnalyst(bedrockService, alpacaService)
news := agents.NewNewsAnalyst(bedrockService, newsAPIService)

// Create manager and register agents
manager := agents.NewPortfolioManager(repo)
manager.RegisterAgent(fundamental)
manager.RegisterAgent(technical)
manager.RegisterAgent(news)

// Analyze a symbol
ctx := context.Background()
recommendation, err := manager.AnalyzeSymbol(ctx, "AAPL")
if err != nil {
    log.Fatal(err)
}

// recommendation.Action will be BUY, SELL, or HOLD
// recommendation.Reasoning will contain detailed explanation
// recommendation.Confidence will indicate overall conviction
```

## Debugging and Monitoring

### Agent Runs

Every agent execution creates an `AgentRun` record:
- Captures agent type and symbol analyzed
- Records start/completion time
- Captures output data (score, confidence, reasoning)
- Logs any errors encountered
- Measures execution duration

Query agent runs to:
- Understand why a recommendation was made
- Debug failing agents
- Monitor performance over time
- Audit analysis history

### Score Interpretation

- Scores clustered around 0: Conflicting signals or insufficient data
- Scores with low confidence: Use caution in trading
- Scores with high confidence from all agents: High conviction signal
- Divergent agent scores: Indicates different market perspectives

## Best Practices

1. **Monitor agent performance**: Track success rates and score distributions
2. **Validate recommendations**: Cross-check suggestions with domain knowledge
3. **Use confidence scores**: Higher confidence recommendations are more reliable
4. **Consider market conditions**: Agents analyze in isolation; external events matter
5. **Regular audits**: Review agent reasoning for consistency and sanity
6. **Tune thresholds**: Adjust BUY/SELL thresholds (currently ±25) based on results

## Configuration

### Environment Variables

The agent system can be configured through environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `AGENT_TIMEOUT_SECONDS` | 30 | Timeout for each agent analysis |
| `ANALYSIS_CONCURRENCY_LIMIT` | 3 | Max concurrent analysis requests |
| `TECHNICAL_ANALYSIS_LOOKBACK_DAYS` | 100 | Historical data period for technical analysis |
| `AGENT_WEIGHT_FUNDAMENTAL` | 0.4 | Weight for fundamental analysis (40%) |
| `AGENT_WEIGHT_NEWS` | 0.3 | Weight for news sentiment (30%) |
| `AGENT_WEIGHT_TECHNICAL` | 0.3 | Weight for technical analysis (30%) |
| `BEDROCK_MAX_TOKENS` | 4096 | Max tokens for Claude API responses |
| `BEDROCK_ANTHROPIC_VERSION` | bedrock-2023-05-31 | Anthropic API version |

**Note**: Agent weights should sum to 1.0 for proper score synthesis.

### Timeout Behavior

Each agent runs with a configurable timeout (default: 30 seconds). When a timeout occurs:
- The agent's goroutine is cancelled via context
- The analysis returns with an error
- The Portfolio Manager continues with successful agents
- At least one agent must succeed for a recommendation

### Rate Limiting

Analysis requests are rate-limited using a semaphore (default: 3 concurrent). When the queue is full:
- New requests immediately return an error
- User receives: "analysis queue full, too many concurrent requests - try again later"
- This prevents API quota exhaustion and controls costs

## Error Handling

### Resilience Features

1. **Agent Timeouts**: Each agent has a 30-second timeout to prevent hung requests
2. **Graceful Degradation**: Failed agents don't prevent recommendations
3. **Rate Limiting**: Prevents API quota exhaustion
4. **Partial Data Handling**:
   - Technical Analyst: Returns neutral score (0) with low confidence if < 50 bars
   - News Analyst: Returns neutral score (0) with low confidence if < 15 articles
   - Fundamental Analyst: Logs warnings for unparseable metrics
5. **Error Logging**: All agent failures are logged with symbol context

### Data Parsing

All external API responses are validated with proper error handling:
- **Time parsing**: Invalid timestamps log warnings and use current time
- **Numeric parsing**: Invalid numbers log warnings and use zero values
- **JSON parsing**: Malformed responses return neutral scores with medium confidence

## Future Enhancements

Potential improvements to the agent architecture:
- Additional agents (sentiment from social media, competitor analysis)
- Machine learning-based confidence adjustment
- Dynamic weighting based on market regime
- Target price calculations based on fundamental valuation
- Risk assessment agent
- Sector rotation signals
