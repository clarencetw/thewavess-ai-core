.PHONY: help install run build test clean docs docker

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

docker-build: ## Build Docker image
	docker build -t thewavess-ai-core .

docker-run: ## Run Docker container
	docker run -p 8080:8080 thewavess-ai-core

dev: docs run ## Development mode: generate docs and run server