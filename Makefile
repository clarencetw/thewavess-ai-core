.PHONY: help install run build test clean docs docs-serve migrate migrate-down migrate-status migrate-reset db-init db-setup fixtures fixtures-recreate create-migration nsfw-embeddings nsfw-check test-api test-all test-integration test-system run-bg stop-bg docker-build docker-run dev check

# 預設目標
help: ## 📋 顯示幫助訊息
	@echo "🚀 Thewavess AI Core - 可用指令："
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "💡 常用指令組合："
	@echo "  make dev           🔄 開發模式 (Ctrl+C停止)"
	@echo "  make stop-dev      ⏹️ 停止開發伺服器"
	@echo "  make db-setup      🏗️ 完整資料庫設置"
	@echo "  make fixtures      🌱 載入 fixtures 數據"

# ===============================
# 📦 基礎開發指令
# ===============================

install: ## 📦 安裝依賴套件
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/air-verse/air@latest
	@echo "✅ Dependencies installed successfully"

build: ## 🔨 編譯應用程式
	@echo "🔨 Building application..."
	@mkdir -p bin
	@echo "🔖 Setting build variables..."
	$(eval VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0"))
	$(eval BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ"))
	$(eval GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown"))
	@go build \
		-ldflags "-X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'" \
		-o bin/thewavess-ai-core main.go
	@echo "✅ Build completed: bin/thewavess-ai-core"
	@echo "   Version: $(VERSION)"
	@echo "   Build Time: $(BUILD_TIME)"
	@echo "   Git Commit: $(GIT_COMMIT)"

test: ## 🧪 執行測試
	@echo "🧪 Running tests..."
	@go test -v ./...

clean: ## 🧹 清理編譯產物
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf docs/
	@echo "✅ Cleanup completed"

# ===============================
# 📚 文檔相關指令
# ===============================

docs: ## 📚 生成 Swagger 文檔
	@echo "📚 Generating Swagger documentation..."
	@swag init --parseDependency --parseInternal
	@echo "✅ Documentation generated"

docs-serve: docs ## 🌐 生成文檔並啟動 Swagger UI
	@echo "🌐 Starting server with Swagger UI..."
	@echo "📖 Swagger documentation: http://localhost:8080/swagger/index.html"
	@echo "🚀 Starting server with enhanced logging..."
	@echo "📊 Logs will be displayed below and saved to server.log"
	@go run main.go 2>&1 | tee server.log

# ===============================
# 🏗️ 資料庫管理指令
# ===============================

db-init: ## 🏗️ 初始化遷移表
	@echo "🏗️ Initializing migration tables..."
	@go run cmd/bun/main.go db init

migrate: ## ⬆️ 執行資料庫遷移
	@echo "⬆️ Running database migrations..."
	@go run cmd/bun/main.go db migrate

migrate-down: ## ⬇️ 回滾上一次遷移
	@echo "⬇️ Rolling back last migration..."
	@go run cmd/bun/main.go db rollback

migrate-status: ## 📊 顯示遷移狀態
	@echo "📊 Checking migration status..."
	@go run cmd/bun/main.go db status

migrate-reset: ## 🔄 重置所有遷移
	@echo "🔄 ⚠️  WARNING: This will reset ALL migrations!"
	@echo "💭 Use with caution in production environments"
	@go run cmd/bun/main.go db reset

db-setup: db-init migrate ## 🏗️ 完整資料庫設置
	@echo ""
	@echo "🎉 Database setup completed successfully!"
	@echo "💡 Next step: run 'make fixtures' to load test data"

create-migration: ## ➕ 創建新的 SQL 遷移文件 (使用: make create-migration NAME=migration_name)
ifndef NAME
	@echo "❌ Error: NAME parameter is required"
	@echo "💡 Usage: make create-migration NAME=your_migration_name"
	@exit 1
endif
	@echo "➕ Creating new migration: $(NAME)"
ifdef TYPE
	@go run cmd/bun/main.go create-migration --type=$(TYPE) $(NAME)
else
	@go run cmd/bun/main.go create-migration --type=go $(NAME)
endif

# ===============================
# 🌱 種子數據管理
# ===============================

fixtures: ## 🌱 載入 fixtures 數據
	@echo "🌱 Loading fixtures..."
	@go run cmd/bun/main.go db fixtures

fixtures-recreate: ## 🔄 重建表格並載入 fixtures
	@echo "🔄 Recreating tables and loading fixtures..."
	@go run cmd/bun/main.go db fixtures --recreate


# ===============================
# 🚀 服務器管理指令
# ===============================

run: ## 🚀 以開發模式運行伺服器
	@echo "🚀 Starting development server..."
	@echo "🌐 Web interface: http://localhost:8080"
	@echo "📖 Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "💚 Health check: http://localhost:8080/health"
	@echo "📊 Logs will be displayed below and saved to server.log"
	@echo ""
	@go run main.go 2>&1 | tee server.log

run-bg: ## 🔙 在後台啟動伺服器
	@echo "🔙 Starting server in background..."
	@go run main.go > server.log 2>&1 & echo $$! > .server.pid
	@echo "✅ Server started in background (PID: $$(cat .server.pid))"
	@echo "📊 Logs: tail -f server.log"

stop-bg: ## ⏹️ 停止後台伺服器
	@if [ -f .server.pid ]; then \
		echo "⏹️ Stopping background server..."; \
		kill `cat .server.pid` 2>/dev/null || true; \
		rm .server.pid; \
		echo "✅ Server stopped"; \
	else \
		echo "❌ No background server found"; \
	fi

stop-dev: ## ⏹️ 停止開發伺服器
	@pkill -f "air" 2>/dev/null || echo "✅ No air processes found"

dev: docs ## 🔄 開發模式 (前台運行，Ctrl+C停止)
	@echo "🔄 Starting development server..."
	@echo "🌐 http://localhost:8080 | 📖 http://localhost:8080/swagger/"
	@echo "🎯 Press Ctrl+C to stop"
	@air

dev-manual: docs ## 🔄 開發模式：生成文檔並運行伺服器 (手動重啟)
	@echo "🔄 Starting development mode (manual restart)..."
	@echo "📚 Documentation generated"
	@echo "🌐 Web interface: http://localhost:8080"
	@echo "📖 Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "💚 Health check: http://localhost:8080/health"
	@echo "📊 Enhanced logging enabled (console + server.log)"
	@echo ""
	@echo "🎯 Press Ctrl+C to stop the server"
	@echo "================================================"
	@go run main.go 2>&1 | tee server.log

check: ## 🔍 檢查所有服務是否正在運行
	@echo "🔍 === Service Health Check ==="
	@echo "🌐 Web UI: http://localhost:8080/"
	@echo "📖 Swagger: http://localhost:8080/swagger/index.html"
	@echo "💚 Health: http://localhost:8080/health"
	@echo ""
	@if curl -s http://localhost:8080/health >/dev/null 2>&1; then \
		echo "✅ API Server is running"; \
		echo "📊 Health Status:"; \
		curl -s http://localhost:8080/health | jq '.' 2>/dev/null || echo "  📄 (Raw response - JSON parsing not available)"; \
	else \
		echo "❌ API Server not running"; \
		echo "💡 Start with: make run or make dev"; \
	fi

# ===============================
# 🔒 NSFW 內容分級相關指令
# ===============================

nsfw-embeddings: ## 🧠 預計算 NSFW 語料庫的 embedding 向量
	@echo "🧠 Computing NSFW corpus embeddings..."
	@if [ ! -f configs/nsfw/corpus.json ]; then \
		echo "❌ NSFW corpus data file not found: configs/nsfw/corpus.json"; \
		exit 1; \
	fi
	@go run tools/nsfw_embeddings.go
	@echo "✅ NSFW embeddings computation completed"


nsfw-check: ## 🔍 檢查 NSFW 語料庫向量完整性
	@echo "🔍 Checking NSFW corpus embeddings..."
	@if [ ! -f configs/nsfw/corpus.json ]; then \
		echo "❌ NSFW corpus data file not found: configs/nsfw/corpus.json"; \
		exit 1; \
	fi
	@if [ ! -f configs/nsfw/embeddings.json ]; then \
		echo "❌ NSFW embeddings file not found: configs/nsfw/embeddings.json"; \
		exit 1; \
	fi
	@go run tools/nsfw_check.go

# ===============================
# 🧪 測試相關指令
# ===============================

test-api: run-bg ## 🧪 測試 API 端點
	@echo "🧪 Testing API endpoints..."
	@echo "⏰ Waiting for server to start..."
	@sleep 3
	@if [ -f tests/api/test_api.sh ]; then \
		cd tests && ./api/test_api.sh; \
	else \
		echo "❌ tests/api/test_api.sh not found"; \
	fi
	@$(MAKE) stop-bg

# 測試套件指令
test-all: run-bg ## 🧪 執行所有測試
	@echo "🧪 Running all tests..."
	@echo "⏰ Waiting for server to start..."
	@sleep 3
	@if [ -f tests/run-all.sh ]; then \
		cd tests && ./run-all.sh all; \
	else \
		echo "❌ tests/run-all.sh not found"; \
	fi
	@$(MAKE) stop-bg

test-integration: run-bg ## 🔄 執行整合測試
	@echo "🔄 Running integration tests..."
	@echo "⏰ Waiting for server to start..."
	@sleep 3
	@if [ -f tests/run-all.sh ]; then \
		cd tests && ./run-all.sh integration; \
	else \
		echo "❌ tests/run-all.sh not found"; \
	fi
	@$(MAKE) stop-bg

test-system: run-bg ## 🔧 執行系統測試
	@echo "🔧 Running system tests..."
	@echo "⏰ Waiting for server to start..."
	@sleep 3
	@if [ -f tests/run-all.sh ]; then \
		cd tests && ./run-all.sh system; \
	else \
		echo "❌ tests/run-all.sh not found"; \
	fi
	@$(MAKE) stop-bg

# ===============================
# 🐳 Docker 相關指令  
# ===============================

docker-build: ## 🐳 構建 Docker 映像
	@echo "🐳 Building Docker image..."
	@echo "🔖 Setting build variables..."
	$(eval VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0"))
	$(eval BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ"))
	$(eval GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown"))
	@docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t thewavess-ai-core .
	@echo "✅ Docker image built: thewavess-ai-core"
	@echo "   Version: $(VERSION)"
	@echo "   Build Time: $(BUILD_TIME)"
	@echo "   Git Commit: $(GIT_COMMIT)"

docker-compose-build: ## 🐳 使用 docker-compose 構建映像
	@echo "🐳 Building with docker-compose..."
	@echo "🔖 Setting build variables..."
	$(eval VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0"))
	$(eval BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ"))
	$(eval GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown"))
	@VERSION=$(VERSION) BUILD_TIME=$(BUILD_TIME) GIT_COMMIT=$(GIT_COMMIT) docker compose build
	@echo "✅ Docker-compose build completed"
	@echo "   Version: $(VERSION)"
	@echo "   Build Time: $(BUILD_TIME)"
	@echo "   Git Commit: $(GIT_COMMIT)"

docker-run: ## 🐳 運行 Docker 容器
	@echo "🐳 Running Docker container..."
	@echo "🌐 Container will be available at: http://localhost:8080"
	@docker run -p 8080:8080 thewavess-ai-core

# ===============================
# 🎯 常用組合指令
# ===============================

fresh-start: clean install db-setup fixtures ## 🎯 全新開始：清理+安裝+資料庫設置+fixtures
	@echo ""
	@echo "🎉 Fresh start completed!"
	@echo "💡 Ready to run: make dev"

quick-setup: db-setup fixtures ## ⚡ 快速設置：資料庫+fixtures
	@echo ""
	@echo "⚡ Quick setup completed!"
	@echo "💡 Ready to run: make dev"

