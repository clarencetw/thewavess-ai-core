# 🚀 部署指南

## 環境需求
- Go 1.23+
- PostgreSQL 15+
- Docker & Docker Compose (推薦)

## 快速部署

### 1. 環境準備
```bash
git clone https://github.com/clarencetw/thewavess-ai-core.git
cd thewavess-ai-core
cp .env.example .env
# 編輯 .env，至少設定 OPENAI_API_KEY
```

### 2. 一鍵啟動
```bash
make fresh-start    # 完整設置
make dev           # 啟動服務
```

### 3. 驗證部署
- **健康檢查**: http://localhost:8080/health
- **API文檔**: http://localhost:8080/swagger/index.html
- **Web介面**: http://localhost:8080

## Docker 部署

```bash
# 使用 Docker Compose
docker-compose up -d

# 檢查服務狀態
docker-compose ps
```

## 生產環境配置

### 必填環境變數
```env
# 服務配置
PORT=8080
GIN_MODE=release
ENVIRONMENT=production

# API Keys
OPENAI_API_KEY=your-production-openai-key
GROK_API_KEY=your-production-grok-key

# 安全配置
JWT_SECRET=your-super-secret-production-key

# 資料庫配置
DB_HOST=your-db-host
DB_USER=your-db-user
DB_PASSWORD=your-db-password
DB_NAME=thewavess_ai_core
```

### AI 引擎配置
```env
# OpenAI (Level 1-4)
OPENAI_MODEL=gpt-4o
OPENAI_MAX_TOKENS=1200

# Grok (Level 5)
GROK_MODEL=grok-3
GROK_MAX_TOKENS=2000
```

## 常用指令

### 開發
```bash
make help           # 查看所有指令
make fresh-start    # 完整重建
make dev           # 開發模式
make test-all      # 完整測試
```

### 數據庫
```bash
make db-setup      # 數據庫初始化
make fixtures      # 載入種子數據
make migrate-reset # 重置數據庫
```

### 構建
```bash
make build         # 編譯應用
make docker-build  # 建立 Docker 映像
```

## 監控

### 健康檢查
```bash
curl http://localhost:8080/health
```

### 監控端點
- `GET /health` - 健康檢查
- `GET /api/v1/monitor/stats` - 詳細狀態
- `GET /api/v1/monitor/metrics` - Prometheus 指標

### 測試套件
```bash
./tests/test-all.sh           # 全部測試 (24項)
./tests/test-all.sh --type api # 僅 API 測試
```

## 故障排除

### 常見問題
```bash
# 端口被佔用
lsof -ti:8080 | xargs kill -9

# 清理構建
make clean && make install

# 查看日誌
docker-compose logs -f  # Docker環境
tail -f server.log      # 本地環境
```

### API Key 驗證
```bash
# 測試 OpenAI 連接
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     https://api.openai.com/v1/models
```

## 安全配置

### HTTPS 反向代理 (Nginx)
```nginx
server {
    listen 443 ssl;
    server_name api.example.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 頻率限制
- 一般 API: 100 請求/分鐘
- 對話 API: 30 請求/分鐘  
- TTS API: 20 請求/分鐘

## 系統特色
- **5個數據表**: 精簡架構，JSONB 整合
- **49個API端點**: 100% 實現
- **AI智能路由**: OpenAI + Grok 自動選擇
- **5級NSFW分類**: 女性向優化