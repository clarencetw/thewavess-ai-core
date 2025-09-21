.PHONY: help install run build test clean docs dev check db-setup fixtures migrate migrate-reset create-migration test-all test-api test-chat test-relationships test-user test-tts test-admin test-search docker-up docker-down fresh-start quick-setup lint nsfw-check nsfw-embeddings

# ==========================================
# 🚀 Thewavess AI Core - 女性向智能對話系統
# ==========================================

help: ## 📋 顯示可用指令
	@echo "🚀 Thewavess AI Core - 女性向智能對話系統"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "🔥 快速開始:"
	@echo "  make fresh-start  🎯 全新安裝 (首次)"
	@echo "  make dev          🔄 開發模式 (日常)"

# ================================
# 📦 基礎指令
# ================================

install: ## 📦 安裝依賴
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/air-verse/air@latest
	@echo "✅ Dependencies installed"

docs: ## 📚 生成 API 文檔
	@echo "📚 Generating documentation..."
	@swag init --parseDependency --parseInternal
	@echo "✅ Documentation generated"

build: ## 🔨 編譯程式
	@echo "🔨 Building application..."
	@mkdir -p bin
	$(eval VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev"))
	$(eval BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ"))
	$(eval GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown"))
	@go build -ldflags "-X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GitCommit=$(GIT_COMMIT)'" -o bin/thewavess-ai-core main.go
	@echo "✅ Build completed: bin/thewavess-ai-core"

test: ## 🧪 Go 單元測試
	@echo "🧪 Running Go tests..."
	@go test -v ./...

clean: ## 🧹 清理產物
	@echo "🧹 Cleaning..."
	@rm -rf bin/ docs/ tmp/
	@pkill -f "air" 2>/dev/null || true
	@echo "✅ Cleanup completed"

lint: ## 🔍 程式碼檢查
	@echo "🔍 Running linters..."
	@go fmt ./...
	@go vet ./...
	@echo "✅ Linting completed"

# ================================
# 🔄 開發與運行
# ================================

dev: docs ## 🔄 開發模式 (推薦)
	@echo "🔄 Starting development server..."
	@echo "🌐 Server: http://localhost:8080"
	@echo "📖 Swagger: http://localhost:8080/swagger/"
	@echo "🎯 Press Ctrl+C to stop"
	@air

run: ## 🚀 直接運行
	@echo "🚀 Starting server..."
	@echo "🌐 http://localhost:8080"
	@go run main.go

check: ## 🔍 健康檢查
	@echo "🔍 Checking server health..."
	@if curl -s http://localhost:8080/health >/dev/null 2>&1; then \
		echo "✅ Server is running"; \
		curl -s http://localhost:8080/health | jq '.' 2>/dev/null || echo "📄 Server responding"; \
	else \
		echo "❌ Server not running - try 'make dev'"; \
	fi

# ================================
# 🗃️ 資料庫管理
# ================================

db-setup: ## 🗃️ 完整資料庫設置
	@echo "🗃️ Setting up database..."
	@go run cmd/bun/main.go db init
	@go run cmd/bun/main.go db migrate
	@echo "✅ Database setup completed"

fixtures: ## 🌱 載入測試資料
	@echo "🌱 Loading fixtures..."
	@go run cmd/bun/main.go db fixtures
	@echo "✅ Fixtures loaded"

migrate: ## ⬆️ 執行遷移
	@echo "⬆️ Running migrations..."
	@go run cmd/bun/main.go db migrate

migrate-reset: ## 🔄 重置資料庫 (危險)
	@echo "🔄 ⚠️  WARNING: This will reset ALL data!"
	@read -p "Continue? (y/N): " confirm && [ "$$confirm" = "y" ]
	@go run cmd/bun/main.go db reset
	@echo "✅ Database reset"

create-migration: ## ➕ 建立遷移檔
ifndef NAME
	@echo "❌ Usage: make create-migration NAME=migration_name"
	@exit 1
endif
	@echo "➕ Creating migration: $(NAME)"
	@go run cmd/bun/main.go create-migration --type=go $(NAME)

# ================================
# 🧪 測試套件
# ================================

test-all: ## 🧪 完整測試套件
	@echo "🧪 Running all tests..."
	@cd tests && ./test-all.sh

test-api: ## 🔌 API 基礎測試
	@echo "🔌 Running API tests..."
	@cd tests && ./chat_api_validation.sh

test-chat: ## 💬 聊天功能測試
	@echo "💬 Running chat tests..."
	@cd tests && ./test_chat_advanced.sh

test-relationships: ## 💕 關係系統測試
	@echo "💕 Running relationship tests..."
	@cd tests && ./test_relationships.sh

test-user: ## 👤 用戶功能測試
	@echo "👤 Running user profile tests..."
	@cd tests && ./test_user_profile.sh

test-tts: ## 🔊 語音合成測試
	@echo "🔊 Running TTS tests..."
	@cd tests && ./test_tts.sh

test-admin: ## 👑 管理後台測試
	@echo "👑 Running admin tests..."
	@cd tests && ./test_admin_advanced.sh

test-search: ## 🔍 搜尋功能測試
	@echo "🔍 Running search tests..."
	@cd tests && ./test_search.sh

# ================================
# 🐳 容器化部署
# ================================

docker-up: ## 🐳 啟動服務
	@echo "🐳 Starting services with docker-compose..."
	@docker compose up -d
	@echo "✅ Services running in background"

docker-down: ## 🐳 停止服務
	@echo "🐳 Stopping services..."
	@docker compose down
	@echo "✅ Services stopped"

# ================================
# 🎯 快速指令
# ================================

fresh-start: clean install db-setup fixtures ## 🎯 全新安裝
	@echo ""
	@echo "🎉 Fresh installation completed!"
	@echo "🚀 Run: make dev"

quick-setup: db-setup fixtures ## ⚡ 快速設置
	@echo ""
	@echo "⚡ Quick setup completed!"
	@echo "🚀 Run: make dev"

# ================================
# 🔒 NSFW 語料管理
# ================================

nsfw-check: ## 🔒 檢查語料狀態
	@echo "🔒 Checking NSFW corpus status..."
	@go run ./tools/nsfw-check

nsfw-embeddings: ## 🧠 重建語料向量
	@echo "🧠 Rebuilding NSFW embeddings..."
	@go run ./tools/nsfw-embeddings
