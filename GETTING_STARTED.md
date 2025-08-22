# 🚀 快速開始（Getting Started）

本指南協助你在本機快速啟動 Thewavess AI Core，並完成基本的 API 測試。

—

## 先決條件

- Go 1.23+
- OpenAI API Key（必填）
- PostgreSQL（可選；未連線時以精簡模式啟動）

—

## 安裝與啟動

```bash
make install                 # 安裝依賴與 swag
cp .env.example .env         # 複製環境變數樣板
# 編輯 .env，至少設定 OPENAI_API_KEY

make dev                     # 生成 Swagger 並啟動服務
```

啟動後常用端點：
- Web UI: http://localhost:8080/
- Swagger: http://localhost:8080/swagger/index.html
- Health: http://localhost:8080/health

—

## 使用 PostgreSQL（可選）

```bash
# 啟動資料庫（示例：Docker）
docker run -d --name pg -e POSTGRES_PASSWORD=pass -e POSTGRES_DB=thewavess_ai_core -p 5432:5432 postgres:15

# 初始化資料庫與種子資料
make db-setup
make seed
```

—

## 基本 API 流程

註冊與登入：
```bash
curl -sS -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","email":"test@example.com","password":"Test123456!"}'

curl -sS -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","password":"Test123456!"}'
```

取得個人資料（需 Bearer Token）：
```bash
curl -H 'Authorization: Bearer <TOKEN>' \
  http://localhost:8080/api/v1/user/profile
```

建立會話並發送訊息：
```bash
# 建立會話（以實際角色 ID 為準，可先 GET /api/v1/character/list）
curl -sS -X POST http://localhost:8080/api/v1/chat/session \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"character_id":"char_001","title":"測試對話"}'

# 發送訊息
curl -sS -X POST http://localhost:8080/api/v1/chat/message \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"<SESSION_ID>","message":"你好！"}'
```

更多端點請見 Swagger（即時）與 API_PROGRESS.md（可用狀態）。

