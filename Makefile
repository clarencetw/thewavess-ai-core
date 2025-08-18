.PHONY: help install run build test clean docs docs-serve migrate migrate-down migrate-status migrate-reset db-setup test-api run-bg stop-bg docker-build docker-run dev check

# 預設目標
help: ## 顯示幫助訊息
	@echo "可用指令："
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## 安裝依賴套件
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest

run: ## 以開發模式運行伺服器
	go run main.go

build: ## 編譯應用程式
	@mkdir -p bin
	go build -o bin/thewavess-ai-core main.go

test: ## 執行測試
	go test -v ./...

clean: ## 清理編譯產物
	rm -rf bin/
	rm -rf docs/

docs: ## 生成 Swagger 文檔
	swag init --parseDependency --parseInternal

docs-serve: docs ## 生成文檔並啟動 Swagger UI
	@echo "Swagger documentation will be available at: http://localhost:8080/swagger/index.html"
	@echo "Starting server..."
	go run main.go

# 資料庫指令 (Bun ORM + bun/migrate)
migrate: ## 使用 Bun 執行資料庫遷移
	@echo "Running Bun database migrations..."
	go run cmd/migrate/main.go -cmd=up

migrate-down: ## 使用 Bun 回滾上一次遷移
	@echo "Rolling back last migration..."
	go run cmd/migrate/main.go -cmd=down

migrate-status: ## 使用 Bun 顯示遷移狀態
	@echo "Checking Bun migration status..."
	go run cmd/migrate/main.go -cmd=status

migrate-reset: ## 使用 Bun 重置所有遷移
	@echo "Resetting all migrations..."
	go run cmd/migrate/main.go -cmd=reset

db-setup: migrate ## 使用 Bun 遷移設置資料庫
	@echo "✅ Bun database setup completed"

docker-build: ## 構建 Docker 映像
	docker build -t thewavess-ai-core .

docker-run: ## 運行 Docker 容器
	docker run -p 8080:8080 thewavess-ai-core

# 測試指令
test-api: run-bg ## 測試 API 端點
	@echo "Testing API endpoints..."
	@sleep 3  # Wait for server to start
	@./test_api.sh
	@$(MAKE) stop-bg

# 後台伺服器管理
run-bg: ## 在後台啟動伺服器
	@echo "Starting server in background..."
	@go run main.go > server.log 2>&1 & echo $$! > .server.pid

stop-bg: ## 停止後台伺服器
	@if [ -f .server.pid ]; then \
		kill `cat .server.pid` 2>/dev/null || true; \
		rm .server.pid; \
		echo "Server stopped"; \
	fi

dev: docs run ## 開發模式：生成文檔並運行伺服器

check: ## 檢查所有服務是否正在運行
	@echo "=== Service Health Check ==="
	@echo "Web UI: http://localhost:8080/"
	@echo "Swagger: http://localhost:8080/swagger/index.html"
	@echo "Health: http://localhost:8080/health"
	@echo ""
	@if curl -s http://localhost:8080/health >/dev/null 2>&1; then \
		echo "✅ API Server is running"; \
		curl -s http://localhost:8080/health | jq '.' 2>/dev/null || echo "  (Health endpoint response not JSON)"; \
	else \
		echo "❌ API Server not running"; \
	fi