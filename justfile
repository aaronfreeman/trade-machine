# trade-machine justfile

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
dev: generate
    wails dev

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
    go test ./...

# Check for issues
check:
    go vet ./...

# Watch templ files and regenerate
watch:
    templ generate --watch
