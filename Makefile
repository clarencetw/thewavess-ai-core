.PHONY: help install run build test clean docs migrate migrate-status db-setup db-reset test-api run-bg stop-bg

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest

run: ## Run the server in development mode
	go run main.go

build: ## Build the application
	@mkdir -p bin
	go build -o bin/thewavess-ai-core main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf docs/

docs: ## Generate Swagger documentation
	swag init --parseDependency --parseInternal

docs-serve: docs ## Generate docs and serve Swagger UI
	@echo "Swagger documentation will be available at: http://localhost:8080/swagger/index.html"
	@echo "Starting server..."
	go run main.go

# Database commands (Bun ORM + bun/migrate)
migrate: ## Run database migrations using Bun
	@echo "Running Bun database migrations..."
	go run cmd/migrate/main.go -cmd=up

migrate-down: ## Rollback last migration using Bun
	@echo "Rolling back last migration..."
	go run cmd/migrate/main.go -cmd=down

migrate-status: ## Show migration status using Bun
	@echo "Checking Bun migration status..."
	go run cmd/migrate/main.go -cmd=status

migrate-reset: ## Reset all migrations using Bun
	@echo "Resetting all migrations..."
	go run cmd/migrate/main.go -cmd=reset

db-setup: migrate ## Setup database with Bun migrations
	@echo "✅ Bun database setup completed"

# Legacy database commands (kept for compatibility)
legacy-migrate: ## Run legacy database migrations
	@echo "Running legacy database migrations..."
	go run database/migrations/migrate.go

legacy-db-reset: ## Reset legacy database (WARNING: This will drop all data)
	@echo "⚠️  WARNING: This will reset the legacy database and lose all data!"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@echo "Dropping schema_migrations table..."
	@psql $$DATABASE_URL -c "DROP TABLE IF EXISTS schema_migrations CASCADE;" || true
	@echo "Running fresh legacy migrations..."
	@$(MAKE) legacy-migrate

docker-build: ## Build Docker image
	docker build -t thewavess-ai-core .

docker-run: ## Run Docker container
	docker run -p 8080:8080 thewavess-ai-core

# Testing commands
test-api: run-bg ## Test API endpoints
	@echo "Testing API endpoints..."
	@sleep 3  # Wait for server to start
	@./test_api.sh
	@make stop-bg

# Background server management
run-bg: ## Start server in background
	@echo "Starting server in background..."
	@go run main.go > server.log 2>&1 & echo $$! > .server.pid

stop-bg: ## Stop background server
	@if [ -f .server.pid ]; then \
		kill `cat .server.pid` 2>/dev/null || true; \
		rm .server.pid; \
		echo "Server stopped"; \
	fi

dev: docs run ## Development mode: generate docs and run server

check: ## Check if all services are running
	@echo "=== Service Health Check ==="
	@echo "Web UI: http://localhost:8080/"
	@echo "Swagger: http://localhost:8080/swagger/index.html"
	@echo "Health: http://localhost:8080/health"
	@echo ""
	@curl -s http://localhost:8080/health | jq '.' || echo "❌ API Server not running"