# 🚀 部署指南

## 快速部署

### 環境需求
- Go 1.22+
- Docker & Docker Compose (推薦)
- Make (可選，但推薦)

### 本地開發環境

#### 1. 環境準備
```bash
# 1. Clone 專案
git clone https://github.com/clarencetw/thewavess-ai-core.git
cd thewavess-ai-core

# 2. 安裝 Go 依賴
go mod tidy

# 3. 安裝 swag 工具（用於生成 OpenAPI 文檔）
go install github.com/swaggo/swag/cmd/swag@latest
```

#### 2. 環境變數設定
```bash
# 複製環境變數模板
cp .env.example .env

# 編輯 .env 檔案，填入你的 API Keys
nano .env
```

**必填環境變數**：
```env
OPENAI_API_KEY=sk-your-openai-api-key-here
GROK_API_KEY=your-grok-api-key-here
```

#### 3. 啟動服務
```bash
# 生成 API 文檔
make docs

# 啟動開發服務器
make dev
# 或
go run main.go
```

#### 4. 驗證部署
訪問以下端點確認服務正常：
- **健康檢查**: http://localhost:8080/health
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **測試介面**: http://localhost:8080

### Docker 部署

#### 使用 Docker Compose (推薦)
```bash
# 1. 準備環境變數
cp .env.example .env
# 編輯 .env 檔案

# 2. 啟動所有服務
docker-compose up -d

# 3. 檢查服務狀態
docker-compose ps
```

#### 單獨 Docker 運行
```bash
# 建立映像
make docker-build

# 運行容器
make docker-run
```

### 生產環境部署

#### 環境變數配置
```env
# 服務配置
PORT=8080
GIN_MODE=release
ENVIRONMENT=production

# 日誌配置
LOG_LEVEL=info

# API Keys
OPENAI_API_KEY=your-production-openai-key
GROK_API_KEY=your-production-grok-key

# 安全配置
JWT_SECRET=your-super-secret-production-key
```

#### 性能優化設定
```env
# OpenAI 配置
OPENAI_MODEL=gpt-4o
OPENAI_MAX_TOKENS=800
OPENAI_TEMPERATURE=0.8

# Grok 配置
GROK_MODEL=grok-beta
GROK_MAX_TOKENS=1000
GROK_TEMPERATURE=0.9
```

## 開發指令

### Make 指令
```bash
make help         # 查看所有可用指令
make install      # 安裝依賴
make docs         # 生成 API 文檔
make run          # 啟動服務器
make test         # 運行測試
make build        # 編譯應用
make clean        # 清理構建檔案
make docker-build # 建立 Docker 映像
make docker-run   # 運行 Docker 容器
```

### 手動指令
```bash
# 生成 Swagger 文檔
swag init

# 編譯應用
go build -o thewavess-ai-core ./main.go

# 運行測試
go test ./...

# 格式化代碼
go fmt ./...
```

## 監控與日誌

### 日誌配置
系統使用結構化 JSON 日誌：
```json
{
  "level": "info",
  "message": "Chat message processed successfully",
  "timestamp": "2025-08-13 03:05:54",
  "session_id": "test_session",
  "user_id": "user_001",
  "ai_engine": "openai",
  "response_time": 1834
}
```

### 健康檢查
```bash
curl http://localhost:8080/health
```

預期回應：
```json
{
  "status": "ok",
  "service": "thewavess-ai-core",
  "version": "1.0.0"
}
```

### 效能監控端點
- **系統狀態**: `GET /api/v1/status`
- **API 版本**: `GET /api/v1/version`

## 故障排除

### 常見問題

#### 1. 端口被佔用
```bash
# 查找佔用端口的進程
lsof -ti:8080

# 終止進程
lsof -ti:8080 | xargs kill -9
```

#### 2. 環境變數載入失敗
確認 `.env` 檔案存在且格式正確：
```bash
# 檢查 .env 檔案
cat .env

# 手動載入環境變數
source .env
```

#### 3. OpenAI API 錯誤
```bash
# 驗證 API key
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     https://api.openai.com/v1/models
```

#### 4. 編譯錯誤
```bash
# 清理模組快取
go clean -modcache

# 重新載入依賴
go mod download
go mod tidy
```

### 日誌除錯
查看應用日誌：
```bash
# Docker 環境
docker-compose logs -f thewavess-ai-core

# 本地開發
tail -f logs/app.log
```

## 安全考量

### API Keys 管理
- 生產環境使用環境變數或密鑰管理服務
- 定期輪換 API keys
- 監控 API 使用量和異常請求

### HTTPS 配置
生產環境建議使用反向代理：
```nginx
server {
    listen 443 ssl;
    server_name api.thewavess.ai;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 頻率限制
API 已內建頻率限制：
- 一般 API: 100 請求/分鐘
- 對話 API: 30 請求/分鐘
- TTS API: 20 請求/分鐘

## 擴展部署

### 水平擴展
```yaml
# docker-compose.yml
version: '3.8'
services:
  thewavess-api:
    image: thewavess-ai-core:latest
    deploy:
      replicas: 3
    ports:
      - "8080-8082:8080"
```

### 負載均衡
```nginx
upstream thewavess_backend {
    server localhost:8080;
    server localhost:8081;
    server localhost:8082;
}
```

### 資料庫部署
```yaml
# 未來資料庫配置
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: thewavess_ai_core
      POSTGRES_USER: thewavess
      POSTGRES_PASSWORD: ${DB_PASSWORD}
  
  redis:
    image: redis:7
    command: redis-server --appendonly yes
```