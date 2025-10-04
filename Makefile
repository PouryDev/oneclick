.PHONY: build run test clean migrate-up migrate-down migrate-create docker-build docker-run

# Variables
BINARY_NAME=oneclick
BUILD_DIR=build
MIGRATIONS_DIR=migrations
DATABASE_URL?=postgres://user:password@localhost:5432/oneclick?sslmode=disable

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	@go run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Install migrate CLI (if not already installed)
install-migrate:
	@echo "Installing migrate CLI..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run database migrations up
migrate-up:
	@echo "Running database migrations up..."
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

# Run database migrations down
migrate-down:
	@echo "Running database migrations down..."
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down

# Create a new migration
migrate-create:
	@echo "Creating new migration..."
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) $$name

# Check migration status
migrate-status:
	@echo "Checking migration status..."
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

# Force migration version (use with caution)
migrate-force:
	@echo "Forcing migration version..."
	@read -p "Enter version number: " version; \
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $$version

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME) .

docker-run:
	@echo "Running Docker container..."
	@docker run --env-file .env -p 8080:8080 $(BINARY_NAME)

# Development setup
dev-setup: deps install-migrate
	@echo "Development setup complete"
	@echo "Don't forget to:"
	@echo "1. Set up your PostgreSQL database"
	@echo "2. Copy .env.example to .env and configure it"
	@echo "3. Run 'make migrate-up' to apply migrations"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Install linting tools
install-lint:
	@echo "Installing linting tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  migrate-up     - Run database migrations up"
	@echo "  migrate-down   - Run database migrations down"
	@echo "  migrate-create - Create a new migration"
	@echo "  migrate-status - Check migration status"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  dev-setup      - Setup development environment"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  help           - Show this help message"
