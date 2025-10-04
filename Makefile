# OneClick Makefile

.PHONY: help dev-start dev-stop full-start full-stop monitoring logs status clean build test lint

# Default target
help: ## Show this help message
	@echo "OneClick Development Commands"
	@echo "============================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development infrastructure
dev-start: ## Start development infrastructure (PostgreSQL + Redis)
	@echo "Starting development infrastructure..."
	@if [ ! -f .env ]; then cp env.dev.example .env; echo "Created .env file from template"; fi
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	docker-compose -f docker-compose.dev.yml up migrate
	@echo "Development infrastructure is ready!"
	@echo "PostgreSQL: localhost:5433"
	@echo "Redis: localhost:6380"

dev-stop: ## Stop development infrastructure
	docker-compose -f docker-compose.dev.yml down

# Full stack
full-start: ## Start full stack (infrastructure + backend)
	@echo "Starting full OneClick stack..."
	@if [ ! -f .env ]; then cp env.dev.example .env; echo "Created .env file from template"; fi
	docker-compose up -d
	@echo "Full OneClick stack is ready!"
	@echo "Backend API: http://localhost:8080"

full-stop: ## Stop full stack
	docker-compose down

# Monitoring
monitoring: ## Start with monitoring (Prometheus + Grafana)
	@echo "Starting OneClick with monitoring..."
	@if [ ! -f .env ]; then cp env.dev.example .env; echo "Created .env file from template"; fi
	docker-compose --profile monitoring up -d
	@echo "OneClick with monitoring is ready!"
	@echo "Backend API: http://localhost:8080"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000 (admin/admin123)"

# Logs
logs: ## Show logs for all services
	docker-compose logs -f

logs-backend: ## Show backend logs
	docker-compose logs -f backend

logs-postgres: ## Show PostgreSQL logs
	docker-compose logs -f postgres

logs-redis: ## Show Redis logs
	docker-compose logs -f redis

# Status and cleanup
status: ## Show status of all services
	@echo "OneClick Docker Services Status:"
	@docker-compose ps
	@echo ""
	@docker-compose -f docker-compose.dev.yml ps

clean: ## Clean up Docker resources
	@echo "Cleaning up OneClick Docker resources..."
	docker-compose down --remove-orphans
	docker-compose -f docker-compose.dev.yml down --remove-orphans
	@echo "Cleanup completed!"

clean-volumes: ## Clean up Docker resources and volumes (WARNING: Deletes all data)
	@echo "WARNING: This will delete all data volumes!"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ]
	docker-compose down --remove-orphans -v
	docker-compose -f docker-compose.dev.yml down --remove-orphans -v
	docker volume prune -f
	@echo "All data volumes have been deleted!"

# Development commands
build: ## Build the Go application
	go build -o bin/oneclick cmd/server/main.go

run: ## Run the Go application locally
	go run cmd/server/main.go

test: ## Run tests
	go test ./...

test-coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run

lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix

# Database commands
migrate-up: ## Run database migrations
	migrate -path migrations -database "postgres://oneclick_dev:oneclick_dev123@localhost:5433/oneclick_dev?sslmode=disable" up

migrate-down: ## Rollback database migrations
	migrate -path migrations -database "postgres://oneclick_dev:oneclick_dev123@localhost:5433/oneclick_dev?sslmode=disable" down

migrate-create: ## Create a new migration file
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations $$name

# Frontend commands
frontend-dev: ## Start frontend development server
	cd oneclick-web && pnpm dev

frontend-build: ## Build frontend for production
	cd oneclick-web && pnpm build

frontend-install: ## Install frontend dependencies
	cd oneclick-web && pnpm install

# Docker commands
docker-build: ## Build Docker image
	docker build -t oneclick:latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env oneclick:latest

# Utility commands
env-setup: ## Setup environment file
	@if [ ! -f .env ]; then \
		cp env.dev.example .env; \
		echo "Created .env file from template"; \
		echo "Please review and update .env file with your configuration"; \
	else \
		echo ".env file already exists"; \
	fi

check-deps: ## Check if all dependencies are installed
	@echo "Checking dependencies..."
	@command -v docker >/dev/null 2>&1 || { echo "Docker is not installed"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose is not installed"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "Go is not installed"; exit 1; }
	@command -v migrate >/dev/null 2>&1 || { echo "Migrate is not installed"; exit 1; }
	@echo "All dependencies are installed!"