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

# Run tests
test:
    templ generate
    DATABASE_URL="postgres://trademachine:trademachine_dev@localhost:5432/trademachine?sslmode=disable" go test -count=1 -cover ./...

# Run tests with coverage report (opens in browser)
coverage:
    DATABASE_URL="postgres://trademachine:trademachine_dev@localhost:5432/trademachine?sslmode=disable" go test -count=1 -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# Check for issues
check:
    go vet ./...

# Watch templ files and regenerate
watch:
    templ generate --watch

migrate:
	goose -dir migrations postgres "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "host=localhost port=5432 user=trademachine password=trademachine_dev dbname=trademachine sslmode=disable" down

docker-up:
	docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3

docker-down:
	docker-compose down

# Start E2E test database
e2e-up:
	docker-compose -f e2e/docker-compose.test.yml up -d
	@echo "Waiting for test PostgreSQL to be ready..."
	@sleep 3
	goose -dir migrations postgres "host=localhost port=5433 user=trademachine_test password=test_password dbname=trademachine_test sslmode=disable" up

# Stop E2E test database
e2e-down:
	docker-compose -f e2e/docker-compose.test.yml down -v

# Run E2E tests (requires e2e-up first)
e2e-test:
	templ generate
	E2E_DATABASE_URL="postgres://trademachine_test:test_password@localhost:5433/trademachine_test?sslmode=disable" go test -v -tags=e2e -count=1 ./e2e/...

# Run all E2E tests with setup/teardown
e2e: e2e-up
	templ generate
	E2E_DATABASE_URL="postgres://trademachine_test:test_password@localhost:5433/trademachine_test?sslmode=disable" go test -v -tags=e2e -count=1 ./e2e/... || (just e2e-down && exit 1)
	just e2e-down