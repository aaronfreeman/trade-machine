> [!WARNING]
> This project is a vibe coding experiment. All the code and documentation in the repo has been planned, written, and reviewed by AI with a little human
> intervention only when necessary.
> 
# Trade Machine

An AI-powered desktop application for intelligent stock trading analysis and execution. Trade Machine combines multi-agent AI analysis with paper trading capabilities to help users make informed investment decisions.

## Overview

Trade Machine is a sophisticated desktop application that leverages artificial intelligence to analyze stocks from multiple perspectives. The system implements a Level 1 Advisory model where AI agents provide trading recommendations that users can review and approve before execution.

The application integrates with multiple data sources and AI services to provide comprehensive analysis including fundamental metrics, technical indicators, news sentiment, and real-time market data.

## Key Features

- **Level 1 Advisory System**: AI agents generate trading recommendations that require user approval before execution
- **Multi-Agent Analysis**: Specialized agents analyze stocks from different angles:
  - Fundamental Analysis: Financial metrics and valuation using Alpha Vantage
  - Technical Analysis: Price patterns and indicators using Alpaca market data
  - News Sentiment Analysis: Market sentiment from recent news using NewsAPI
- **Paper Trading**: Execute trades in a simulated environment via Alpaca API without real capital
- **Real-Time Market Data**: Stream current market prices and quotes
- **Portfolio Tracking**: Monitor holdings, performance, and trade history
- **AI-Powered Insights**: Claude 3.5 Sonnet via AWS Bedrock generates analysis and recommendations

## Technology Stack

- **Frontend**: Wails v2 (Go desktop application framework) with embedded web UI
- **Backend**: Go 1.25 with composable service architecture
- **Database**: PostgreSQL (via Docker) for persistent data storage
- **AI**: AWS Bedrock with Claude 3.5 Sonnet for natural language analysis
- **Data Sources**:
  - Alpaca API: Paper trading and real-time market data
  - Alpha Vantage: Fundamental financial data
  - NewsAPI: Market news and sentiment data
- **Templating**: templ for type-safe HTML generation
- **Migrations**: goose for database version control
- **Tool Management**: mise for consistent development environment

## Architecture Overview

The application follows a layered architecture pattern:

```
Desktop UI (Wails)
    ↓
HTTP Handler
    ↓
Services Layer (External APIs)
    ├── Alpaca Service (Trading & Market Data)
    ├── Bedrock Service (AI Analysis)
    ├── Alpha Vantage Service (Fundamentals)
    └── NewsAPI Service (News Sentiment)
    ↓
Agents Layer (Multi-Agent Analysis)
    ├── Fundamental Analyst
    ├── Technical Analyst
    └── News Analyst
    ↓
Repository Layer (Data Persistence)
    ↓
PostgreSQL Database
```

### Project Structure

```
trade-machine/
├── agents/               # AI analysis agents and portfolio manager
├── models/               # Data structures and domain models
├── repository/           # Database access layer
├── services/             # External API integrations
│   ├── alpaca/          # Trading and market data
│   ├── bedrock/         # AI analysis
│   ├── alpha_vantage/   # Fundamental data
│   └── newsapi/         # News and sentiment
├── templates/           # UI templates (templ)
├── frontend/            # Wails frontend assets
├── migrations/          # Database schema migrations
├── main.go              # Application entry point
├── app.go               # Wails application struct
├── handler.go           # HTTP request handlers
├── justfile             # Task runner commands
├── mise.toml            # Tool versions
├── docker-compose.yml   # PostgreSQL container
├── wails.json           # Wails configuration
├── go.mod               # Go module dependencies
└── README.md            # This file
```

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.25 or higher**: [Install Go](https://go.dev/doc/install)
- **Docker**: Required to run PostgreSQL ([Install Docker](https://docs.docker.com/get-docker/))
- **mise**: Tool version manager ([Install mise](https://mise.jq.rs/guide/getting-started.html))

## Setup Instructions

### 1. Clone the Repository

```bash
git clone <repository-url>
cd trade-machine
```

### 2. Install Go Dependencies

Use mise to ensure you have the correct tool versions:

```bash
mise install
```

Then download Go dependencies:

```bash
just install
```

### 3. Start PostgreSQL

Start the PostgreSQL database container:

```bash
just docker-up
```

The database will be initialized with credentials:
- Host: `localhost`
- Port: `5432`
- User: `trademachine`
- Password: `trademachine_dev`
- Database: `trademachine`

Wait for the output confirming PostgreSQL is ready.

### 4. Run Database Migrations

Apply database schema migrations:

```bash
just migrate
```

This creates all necessary tables for the application.

### 5. Configure Environment Variables

Copy the example environment file and update with your credentials:

```bash
cp .env.example .env
```

Edit `.env` and fill in your API keys and configuration:

```
# AWS Bedrock Configuration (for AI analysis)
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
BEDROCK_MODEL_ID=anthropic.claude-3-5-sonnet-20241022-v2:0

# PostgreSQL Configuration (use defaults or customize)
DATABASE_URL=postgres://trademachine:trademachine_dev@localhost:5432/trademachine?sslmode=disable

# Alpaca API Configuration (for paper trading)
ALPACA_API_KEY=your_alpaca_key
ALPACA_API_SECRET=your_alpaca_secret
ALPACA_BASE_URL=https://paper-api.alpaca.markets

# Alpha Vantage Configuration (for fundamental analysis)
ALPHA_VANTAGE_API_KEY=your_alpha_vantage_key

# NewsAPI Configuration (for sentiment analysis)
NEWS_API_KEY=your_news_api_key

# Application Configuration
LOG_LEVEL=info
CACHE_TTL_MINUTES=15
```

### 6. Run Development Server

Start the application in development mode:

```bash
just dev
```

The Wails application will launch with hot-reload enabled. The UI will open in a desktop window.

## Configuration

### Environment Variables

All configuration is managed through environment variables. See `.env.example` for all available options:

| Variable | Purpose | Required |
|----------|---------|----------|
| `DATABASE_URL` | PostgreSQL connection string | Yes (database features) |
| `AWS_REGION` | AWS region for Bedrock | Yes (AI analysis) |
| `AWS_ACCESS_KEY_ID` | AWS credentials | Yes (AI analysis) |
| `AWS_SECRET_ACCESS_KEY` | AWS credentials | Yes (AI analysis) |
| `BEDROCK_MODEL_ID` | Claude model ID | Yes (AI analysis) |
| `ALPACA_API_KEY` | Alpaca trading API | Yes (trading) |
| `ALPACA_API_SECRET` | Alpaca trading API | Yes (trading) |
| `ALPACA_BASE_URL` | Alpaca API endpoint | No (defaults to paper trading) |
| `ALPHA_VANTAGE_API_KEY` | Fundamental data API | Yes (fundamental analysis) |
| `NEWS_API_KEY` | News sentiment API | Yes (news analysis) |
| `LOG_LEVEL` | Logging verbosity | No (defaults to info) |
| `CACHE_TTL_MINUTES` | Data cache duration | No (defaults to 15) |
| `CORS_ALLOWED_ORIGINS` | CORS allowed origins | No (defaults to *) |
| `AGENT_TIMEOUT_SECONDS` | Agent timeout | No (defaults to 30) |
| `ANALYSIS_CONCURRENCY_LIMIT` | Max concurrent analyses | No (defaults to 3) |
| `TECHNICAL_ANALYSIS_LOOKBACK_DAYS` | Historical data period | No (defaults to 100) |
| `AGENT_WEIGHT_FUNDAMENTAL` | Fundamental weight | No (defaults to 0.4) |
| `AGENT_WEIGHT_NEWS` | News weight | No (defaults to 0.3) |
| `AGENT_WEIGHT_TECHNICAL` | Technical weight | No (defaults to 0.3) |
| `BEDROCK_MAX_TOKENS` | Max Claude tokens | No (defaults to 4096) |
| `BEDROCK_ANTHROPIC_VERSION` | Anthropic API version | No (defaults to bedrock-2023-05-31) |

The application will start with graceful degradation if optional services are not configured - you can still use available features.

## Available Commands

The `justfile` provides convenient task runners for common operations:

```bash
just                # List all available commands
just install        # Download Go module dependencies
just generate       # Generate templ template files
just dev            # Run development server with hot-reload
just build          # Build production binary
just test           # Run all tests
just fmt            # Format Go and templ code
just check          # Run go vet for code issues
just watch          # Watch templ files and regenerate on changes
just docker-up      # Start PostgreSQL container
just docker-down    # Stop PostgreSQL container
just migrate        # Run database migrations
just migrate-down   # Rollback last database migration
just clean          # Remove build artifacts
```

## Development Workflow

### Code Generation

The project uses templ for type-safe HTML templates. After modifying `.templ` files, regenerate:

```bash
just generate
```

Or watch for changes automatically:

```bash
just watch
```

### Formatting and Linting

Keep code consistent with:

```bash
just fmt
just check
```

### Testing

Run the test suite:

```bash
just test
```

For integration tests that require a database:

```bash
DATABASE_URL=postgres://trademachine:trademachine_dev@localhost:5432/trademachine?sslmode=disable just test
```

### Building for Production

Create an optimized production build:

```bash
just build
```

The compiled binary will be available in the `build/bin/` directory.

## Testing

The project includes comprehensive tests for all major components.

### Running Tests

Run all tests with:

```bash
just test
```

### Integration Tests

Some tests require a running PostgreSQL database. Set the database URL environment variable:

```bash
DATABASE_URL=postgres://trademachine:trademachine_dev@localhost:5432/trademachine?sslmode=disable just test
```

## Troubleshooting

### PostgreSQL Connection Fails

Ensure the Docker container is running:

```bash
docker ps | grep trademachine-postgres
```

If not running, restart it:

```bash
just docker-down
just docker-up
```

### Wails Application Won't Start

Check that all required Go dependencies are installed:

```bash
just install
```

Ensure templ files are generated:

```bash
just generate
```

### Database Migrations Fail

Verify PostgreSQL is running and accessible:

```bash
docker logs trademachine-postgres
```

Check the migration files in the `migrations/` directory are in order.

### API Key Issues

Verify environment variables are loaded:

```bash
env | grep AWS_
env | grep ALPACA_
```

The application logs warnings if optional services fail to initialize. Check the console output for guidance.

## API Reference

The application exposes HTTP endpoints for analysis and trading operations. Refer to `handler.go` for the complete API specification.

Key endpoints include:
- Stock analysis and recommendations
- Portfolio management operations
- Trade execution and history
- Market data queries

## Contributing

When contributing to this project:

1. Ensure code follows Go conventions
2. Write tests for new functionality
3. Run `just fmt` before committing
4. Update database migrations if schema changes
5. Document significant changes in code comments

## Performance Considerations

- Market data is cached with TTL from `CACHE_TTL_MINUTES`
- Database queries use connection pooling via pgx
- Templ generates efficient HTML at compile time
- Wails provides optimized IPC between Go backend and frontend

## Security Notes

- Never commit `.env` files with real credentials
- Use AWS IAM roles in production instead of access keys
- Alpaca API keys should be kept secret
- PostgreSQL connections can be encrypted with `sslmode=require`

## Getting Help

For issues or questions:

1. Check the troubleshooting section above
2. Review console output and logs for error messages
3. Examine similar code in the `agents/` and `services/` directories
4. Check the project's issue tracker

## License

See LICENSE file for details.
