# trade-machine justfile

# Load environment variables from .env
set dotenv-load

# Default recipe
default:
    @just --list

# Install dependencies
install:
    go mod download

# Generate templ files
generate:
    templ generate

# Run in development mode
start: generate
    templ generate
    wails dev -devserver "localhost:$DEV_PORT"

# Build for production
build: generate
    wails build

# Clean build artifacts
clean:
    rm -rf build/
    rm -f *_templ.go
    rm -f **/*_templ.go

# Format code
fmt:
    go fmt ./...
    templ fmt .

# Test database URL (separate from development database)
test_db_url := "postgres://trademachine:trademachine_dev@localhost:5432/trademachine_test?sslmode=disable"

# Run tests (uses separate test database, runs migrations automatically)
test:
    templ generate
    @goose -dir migrations postgres "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine_test sslmode=disable" up 2>&1 | grep -v "no migrations to run" || true
    DATABASE_URL="{{test_db_url}}" go test -count=1 -cover ./...

# Run tests with coverage report (opens in browser)
coverage:
    @goose -dir migrations postgres "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine_test sslmode=disable" up 2>&1 | grep -v "no migrations to run" || true
    DATABASE_URL="{{test_db_url}}" go test -count=1 -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# Check for issues
check:
    go vet ./...

# Watch templ files and regenerate
watch:
    templ generate --watch

# Run migrations on development database
migrate:
	goose -dir migrations postgres "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable" up

# Rollback migrations on development database
migrate-down:
	goose -dir migrations postgres "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable" down

# Run migrations on test database
migrate-test:
	goose -dir migrations postgres "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine_test sslmode=disable" up

# Rollback migrations on test database
migrate-test-down:
	goose -dir migrations postgres "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine_test sslmode=disable" down

# Reset test database (rollback all and re-apply)
migrate-test-reset:
	goose -dir migrations postgres "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine_test sslmode=disable" reset
	goose -dir migrations postgres "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine_test sslmode=disable" up

docker-up:
	docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3

docker-down:
	docker-compose down

# Initialize test database (for existing containers where init script didn't run)
init-test-db:
	@echo "Creating test database if it doesn't exist..."
	@docker exec trademachine-postgres psql -U trademachine -d postgres -c "CREATE DATABASE trademachine_test;" 2>/dev/null || echo "Test database already exists"
	@echo "Running migrations on test database..."
	goose -dir migrations postgres "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine_test sslmode=disable" up

# Start E2E test database
e2e-up:
	docker-compose -f e2e/docker-compose.test.yml up -d
	@echo "Waiting for test PostgreSQL to be ready..."
	@sleep 3
	goose -dir migrations postgres "host=localhost port=5433 user=trademachine_test password=test_password dbname=trademachine_test sslmode=disable" up

# Stop E2E test database
e2e-down:
	docker-compose -f e2e/docker-compose.test.yml down -v

# E2E database URL for Playwright tests
e2e_db_url := "postgres://trademachine_test:test_password@localhost:5433/trademachine_test?sslmode=disable"

# Install Playwright dependencies
playwright-install:
	cd e2e/playwright && npm install
	cd e2e/playwright && npx playwright install chromium

# Run Playwright E2E tests (browser-based)
playwright-test: e2e-up
	templ generate
	cd e2e/playwright && E2E_DATABASE_URL="{{e2e_db_url}}" npm test || (just e2e-down && exit 1)
	just e2e-down

# Run Playwright tests in headed mode (visible browser)
playwright-headed: e2e-up
	templ generate
	cd e2e/playwright && E2E_DATABASE_URL="{{e2e_db_url}}" npm run test:headed || (just e2e-down && exit 1)
	just e2e-down

# Run Playwright tests with debug UI
playwright-debug: e2e-up
	templ generate
	cd e2e/playwright && E2E_DATABASE_URL="{{e2e_db_url}}" npm run test:debug

# Run Playwright tests with interactive UI
playwright-ui: e2e-up
	templ generate
	cd e2e/playwright && E2E_DATABASE_URL="{{e2e_db_url}}" npm run test:ui

# Show Playwright test report
playwright-report:
	cd e2e/playwright && npm run report

# Start the E2E test server manually (for development)
e2e-server: e2e-up
	templ generate
	E2E_DATABASE_URL="{{e2e_db_url}}" go run ./cmd/e2e-server

# Alias for playwright-test (shorter to type)
pw: playwright-test

# Run all tests (unit tests + Playwright E2E tests)
test-all: test playwright-test