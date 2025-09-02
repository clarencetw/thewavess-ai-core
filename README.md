# 🤖 Thewavess AI Core

—

## 重要說明

- 端點、狀態與可用性以下列來源為準：
  - API 進度與可用性：[API_PROGRESS.md](./API_PROGRESS.md)
  - 即時 API 參考：/swagger/index.html（自動生成）
  - 統一測試工具：[tests/test-all.sh](./tests/test-all.sh)
- 本 README 移除舊版的功能宣稱與過時端點清單，僅保留經驗證的快速使用資訊。

—

## 快速開始

環境需求：
- Go 1.23+
- PostgreSQL（必需；用於數據存儲）
- OpenAI API Key（必填）；Grok/TTS API Key（可選）

步驟：
```bash
make install
cp .env.example .env     # 至少設定 OPENAI_API_KEY
make fresh-start         # 完整設置：清理+安裝+資料庫+fixtures
make dev                 # 生成 Swagger 並啟動服務

# 可選：驗證系統運行狀態
./tests/test-all.sh      # 執行完整測試套件 (24個測試項目，100%通過率)
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
curl -sS -X POST http://localhost:8080/api/v1/chats \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"character_id":"character_01","title":"測試對話"}'

# 發送訊息
curl -sS -X POST http://localhost:8080/api/v1/chats/<CHAT_ID>/messages \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"message":"你好！"}'
```

更多端點與狀態說明請見 [API_PROGRESS.md](./API_PROGRESS.md) 或 Swagger。

—

## 專案結構（重點目錄）

```
handlers/   HTTP handlers（auth、user、chat、character、monitor 等）
services/   核心服務（chat、nsfw、emotion、tts、openai/grok 客戶端）
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

基本指令：
```bash
make install      # 安裝依賴與 swag  
make dev          # 生成 Swagger 並啟動
make fresh-start  # 完整重建（推薦首次使用）
make quick-setup  # 快速設置（資料庫+fixtures）
```

> 📋 完整指令說明請參考 [DEVELOPMENT.md](./DEVELOPMENT.md)

—

## 設定與環境變數

請參考 .env.example，至少設定：
- OPENAI_API_KEY（必填）
- DB_*（若連線資料庫）
- GROK_API_KEY / TTS_API_KEY（可選）

—

## 文件與指南

### 核心文檔
- 入門指南：[GETTING_STARTED.md](./GETTING_STARTED.md)
- 系統架構：[ARCHITECTURE.md](./ARCHITECTURE.md)
- 開發流程：[DEVELOPMENT.md](./DEVELOPMENT.md)
- 環境設定：[CONFIGURATION.md](./CONFIGURATION.md)
- API 與進度：[API_PROGRESS.md](./API_PROGRESS.md)（權威來源）
- 完整 API 參考：[Swagger UI](http://localhost:8080/swagger/index.html)
- 規格與設計：[SPEC.md](./SPEC.md)

### 系統指南
- 角色系統：[CHARACTER_GUIDE.md](./CHARACTER_GUIDE.md)
- 關係系統：[RELATIONSHIP_GUIDE.md](./RELATIONSHIP_GUIDE.md)
- 好感度系統：[AFFECTION_GUIDE.md](./AFFECTION_GUIDE.md)
- 聊天模式：[CHAT_MODES.md](./CHAT_MODES.md)
- NSFW 設計指南：[NSFW_GUIDE.md](./NSFW_GUIDE.md)

### 操作指南
- 部署指引：[DEPLOYMENT.md](./DEPLOYMENT.md)
- 監控指南：[MONITORING_GUIDE.md](./MONITORING_GUIDE.md)

### 開發工具
- AI 代理配置：[AGENTS.md](./AGENTS.md)
- Claude 使用指南：[CLAUDE.md](./CLAUDE.md)
- 測試系統說明：[tests/README.md](./tests/README.md)
