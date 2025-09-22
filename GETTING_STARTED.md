# 🚀 快速開始

> 📋 **相關文檔**: 完整文檔索引請參考 [DOCS_INDEX.md](./DOCS_INDEX.md)

## 系統概覽
- **API 端點**: 57 個 (100% 已實現)
- **資料表**: 5 張核心表
- **技術棧**: Go 1.23 + Gin + PostgreSQL + Bun ORM

## 先決條件
- Go 1.23+
- OpenAI API Key（必填）
- PostgreSQL

## 快速啟動

```bash
make install                 # 安裝依賴
cp .env.example .env         # 複製環境變數
# 編輯 .env，至少設定 OPENAI_API_KEY

make dev                     # 生成文檔並啟動
```

啟動後訪問：
- **Web UI**: http://localhost:8080/
- **API文檔**: http://localhost:8080/swagger/index.html
- **健康檢查**: http://localhost:8080/health

## 資料庫設置

```bash
# 啟動 PostgreSQL（Docker 範例）
docker run -d --name pg \
  -e POSTGRES_PASSWORD=pass \
  -e POSTGRES_DB=thewavess_ai_core \
  -p 5432:5432 postgres:15

# 一鍵設置（推薦）
make fresh-start       # 完整重建
make quick-setup       # 快速設置
```

## 基本 API 測試

### 用戶認證
```bash
# 註冊
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","email":"test@example.com","password":"Test123456!"}'

# 登入
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","password":"Test123456!"}'
```

### 對話流程
```bash
# 建立會話
curl -X POST http://localhost:8080/api/v1/chats \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"character_id":"character_01","title":"測試對話"}'

# 發送訊息
curl -X POST http://localhost:8080/api/v1/chats/<CHAT_ID>/messages \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"message":"你好！"}'
```

## 完整測試

```bash
./tests/test-all.sh              # 所有測試（24項，100%通過）
./tests/test-all.sh --type api   # API 功能測試
./tests/test-all.sh --type chat  # 對話功能測試
```

## 相關文檔
- **開發流程**: [DEVELOPMENT.md](./DEVELOPMENT.md)
- **API狀態**: [API_PROGRESS.md](./API_PROGRESS.md)
- **配置說明**: [CONFIGURATION.md](./CONFIGURATION.md)

