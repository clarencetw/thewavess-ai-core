# 🤖 Thewavess AI Core

專為成人用戶設計的 AI 聊天後端服務。此版本的 README 已對齊實際程式碼與端點，API 與開發進度請以 [API_PROGRESS.md](./API_PROGRESS.md) 與 Swagger 為準。

徽章：Go 1.23+ | Swagger 可用 | Docker 支援

—

## 重要說明

- 端點、狀態與可用性以下列來源為準：
  - API 進度與可用性：[API_PROGRESS.md](./API_PROGRESS.md)
  - 即時 API 參考：/swagger/index.html（自動生成）
  - 測試腳本：[test_api.sh](./test_api.sh)
- 本 README 移除舊版的功能宣稱與過時端點清單，僅保留經驗證的快速使用資訊。

—

## 快速開始

環境需求：
- Go 1.23+
- PostgreSQL（可選；未連線時以精簡模式啟動）
- OpenAI API Key（必填）；Grok/TTS API Key（可選）

步驟：
```bash
make install
cp .env.example .env  # 至少設定 OPENAI_API_KEY
make dev              # 生成 Swagger 並啟動服務
```

預設端點：
- Web UI: http://localhost:8080/
- Swagger: http://localhost:8080/swagger/index.html
- Health: http://localhost:8080/health
- BasePath: /api/v1

—

## 快速 API 範例

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
curl -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/user/profile
```

建立聊天會話並發送訊息：
```bash
# 建立會話（以實際角色 ID 為準）
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

更多端點與狀態說明請見 [API_PROGRESS.md](./API_PROGRESS.md) 或 Swagger。

—

## 專案結構（重點目錄）

```
handlers/   HTTP handlers（auth、user、chat、character、monitor 等）
services/   核心服務（chat、nsfw、memory、tts、openai/grok 客戶端）
routes/     路由註冊（routes.go）
models/     資料模型
database/   Bun 遷移與工具（cmd/bun）
middleware/ 認證、日誌、CORS
utils/      日誌、錯誤、JWT、輔助工具
public/     靜態頁與 Swagger UI 入口
bin/        編譯輸出
```

—

## 常用指令

開發與執行：
```bash
make install   # 安裝依賴與 swag
make dev       # 生成 Swagger 並啟動
make run       # 僅啟動服務
make build     # 編譯到 bin/thewavess-ai-core
```

文件與測試：
```bash
make docs         # 生成 Swagger
make test         # go test -v ./...
make test-api     # 後台啟動並執行 test_api.sh
```

資料庫（PostgreSQL + Bun）：
```bash
make db-setup         # 初始化遷移表 + 遷移
make migrate          # 執行遷移
make migrate-status   # 查看狀態
make migrate-down     # 回滾一次
make seed             # 填充種子資料
```

Docker：
```bash
make docker-build
make docker-run
```

—

## 設定與環境變數

請參考 .env.example，至少設定：
- OPENAI_API_KEY（必填）
- DB_*（若連線資料庫）
- GROK_API_KEY / TTS_API_KEY（可選）

—

## 文件與指南

- 入門指南：[GETTING_STARTED.md](./GETTING_STARTED.md)
- 系統架構：[ARCHITECTURE.md](./ARCHITECTURE.md)
- 開發流程：[DEVELOPMENT.md](./DEVELOPMENT.md)
- 環境設定：[CONFIGURATION.md](./CONFIGURATION.md)
- API 與進度：[API_PROGRESS.md](./API_PROGRESS.md)（權威來源）
- 完整 API 參考：[API.md](./API.md) + [Swagger UI](http://localhost:8080/swagger/index.html)
- NSFW 政策：[NSFW_GUIDE.md](./NSFW_GUIDE.md)
- 規格與設計：[SPEC.md](./SPEC.md)
- 部署指引：[DEPLOYMENT.md](./DEPLOYMENT.md)

—

## 授權與貢獻

此專案採專有授權。歡迎以 Issue/PR 形式回報問題與建議；提交變更請遵循 Makefile 指令與 go fmt 風格，並附上必要測試。
