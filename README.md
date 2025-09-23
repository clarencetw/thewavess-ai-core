# 🤖 Thewavess AI Core

**女性向情感聊天後端系統** - 生產就緒的 AI 雙引擎架構，支援智能 NSFW 分級與完整關係管理系統。

[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://golang.org/)
[![API](https://img.shields.io/badge/API-57%2F57-green)](./API_PROGRESS.md)
[![Tests](https://img.shields.io/badge/Tests-24%2F24-green)](./tests/)

## ✨ 核心特色

- **🤖 AI 雙引擎**: OpenAI GPT-4o (L1-L3) + Grok AI (L4-L5) 智能路由
- **🛡️ 智能 NSFW 檢測**: 關鍵字分類器 (L1-L5)
- **💕 關係管理系統**: 動態角色個性與好感度追蹤
- **🎵 語音合成**: OpenAI TTS 整合
- **⚡ 零運行成本**: 內建關鍵字規則，微秒級響應
- **📊 完整管理後台**: 用戶管理、監控與分析

## 🚀 快速開始

```bash
# 1. 安裝依賴
make install

# 2. 設定環境變數
cp .env.example .env
# 編輯 .env: 填入 OPENAI_API_KEY、資料庫連線、JWT_SECRET

# 3. 初始化資料庫
make fresh-start

# 4. 啟動服務
make dev
```

**服務端點:**
- 🌐 Web UI: http://localhost:8080/
- 📖 API 文檔: http://localhost:8080/swagger/index.html
- 💚 健康檢查: http://localhost:8080/health

## 📋 API 範例

```bash
# 用戶註冊
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","email":"test@example.com","password":"Test123456!"}'

# 登入
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","password":"Test123456!"}'

# 建立對話
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

## 🏗️ 系統架構

**技術棧**: Go 1.23 · Gin · Bun ORM · PostgreSQL · Redis · Docker

```
├── handlers/     # HTTP 處理器 (Auth, Chat, Admin)
├── services/     # 核心業務邏輯 (AI引擎, NSFW分類, 關係管理)
├── models/       # 數據模型 & API 結構
├── routes/       # 路由配置 (57 個端點)
├── cmd/          # CLI 工具 & 遷移腳本
└── tests/        # 整合測試套件
```

## 🔧 常用指令

```bash
# 開發
make dev              # 生成文檔 + 啟動服務
make fresh-start      # 完整重置 + 安裝
make build           # 編譯執行檔
make test            # 運行測試

# 資料庫
make db-setup        # 遷移 + 初始數據
make fixtures        # 載入測試數據

# 測試
./tests/test-all.sh  # 完整測試套件 (24/24)
```

## 🔑 環境配置

**必需變數:**
- `OPENAI_API_KEY` - OpenAI API 金鑰
- `GROK_API_KEY` - Grok AI 金鑰 (L4/L5 內容)
- `DB_*` - PostgreSQL 連線設定
- `JWT_SECRET` - JWT 簽名金鑰

**可選配置:**
- `NSFW_CORPUS_*` - NSFW 分類器路徑
- `TTS_API_KEY` - 語音合成服務

完整配置請參考 [.env.example](./.env.example) 與 [CONFIGURATION.md](./CONFIGURATION.md)

## 📚 文檔與資源

**📋 完整文檔索引**: [DOCS_INDEX.md](./DOCS_INDEX.md) - 所有文檔的完整列表與分類導航

**🚀 快速入門:**
- [⚙️ 配置指南](./CONFIGURATION.md) - 環境變數設定
- [🚀 部署手冊](./DEPLOYMENT.md) - 生產環境部署

**🔧 開發資源:**
- [📊 API 文檔](./API_PROGRESS.md) - 57 個端點狀態
- [🏛️ 系統架構](./ARCHITECTURE.md) - 技術架構設計
- [🧪 開發流程](./DEVELOPMENT.md) - 開發規範與測試
