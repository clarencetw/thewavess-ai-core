# Thewavess AI Core

Thewavess AI Core 是一個以 Golang 打造的智慧聊天後端，整合多種 AI 引擎（OpenAI、Grok），提供女性向對話、互動小說、情感陪伴與 TTS 語音等核心能力。詳細產品規格請參考 `SPEC.md`。

## 主要特性
- **多模式對話**：普通 / 小說 / NSFW（引擎動態路由）
- **角色系統**：預設人物設定、角色語音配置
- **記憶系統**：短期（Redis）/ 長期（PostgreSQL + 向量資料庫）/ 情感記憶
- **情感系統**：好感度、關係狀態、事件觸發
- **TTS 語音**：OpenAI TTS 預設，支援多種音色
- **標籤系統**：內容/場景/NSFW 觸發標籤
- **可觀測性**：結構化日誌、指標與追蹤（規劃中）

## 技術棧
- **語言/框架**：Go + Gin（HTTP）/ GORM（ORM）
- **資料庫**：PostgreSQL（主資料）/ Redis（快取與會話）
- **向量資料庫**：Qdrant（語義搜尋與記憶檢索）
- **訊息隊列**：RabbitMQ（異步任務與事件）
- **AI 服務**：OpenAI（GPT-4o / TTS）、Grok（NSFW）
- **部署**：Docker / Docker Compose

## 架構總覽（高層）
- **API 層**：REST HTTP 服務
- **服務層**：聊天服務、記憶服務、角色服務、情感服務、TTS 服務
- **引擎路由**：內容分類 → OpenAI 或 Grok（NSFW）
- **資料層**：PostgreSQL / Redis / 向量庫 + MQ（事件與任務）

更多細節請見 `SPEC.md` 的「技術架構」「API 設計」「資料模型」。

## 快速開始（目前為初始化階段）
> 專案仍在規劃與初始化，以下為建議的本地開發步驟（尚未提供完整程式碼）。

1. 安裝環境
   - Go 1.22+
   - PostgreSQL 14+、Redis 6+（或使用 Docker）
2. 初始化 Go 模組（僅首次）
   ```bash
   go mod init github.com/clarencetw/thewavess-ai-core
   go mod tidy
   ```
3. 設定環境變數（建立 `.env` 或使用 shell 匯入）
   ```bash
   export OPENAI_API_KEY=...
   export GROK_API_KEY=...
   export POSTGRES_DSN="postgres://user:pass@localhost:5432/ai_core?sslmode=disable"
   export REDIS_URL="redis://localhost:6379/0"
   export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
   export QDRANT_URL="http://localhost:6333"
   ```
4. 服務啟動
   - 待後端框架落地（Gin 專案骨架、路由與 handler 建置後補）

## API 概覽
核心端點已在 `SPEC.md` 的「API 設計」定義：
- 對話：`/api/v1/chat/*` (9個端點)
- 角色：`/api/v1/character/*` + `/api/v1/user/character` (3個端點)
- 小說模式：`/api/v1/novel/*` (5個端點)
- 情感系統：`/api/v1/emotion/*` (3個端點)
- TTS：`/api/v1/tts/*` (3個端點)
- 記憶系統：`/api/v1/memory/*` (5個端點)
- 用戶系統：`/api/v1/user/*` (7個端點)
- 標籤系統：`/api/v1/tags/*` (2個端點)
- 系統管理：`/api/v1/health`, `/api/v1/version`, `/api/v1/status` (3個端點)

**總計 40 個 API 端點**

實作時將同步補充：
- 請求/回應範例（含錯誤格式）
- 認證（API Key / JWT — 規劃中）

## 安全與合規
- 面向成年用戶，完全開放 NSFW 內容
- 前端負責年齡驗證與內容分級
- 違法內容一律禁止
- 敏感資料加密儲存

## 路線圖（14週）
- Phase 1：基礎建設（專案骨架 / 資料庫 / API 框架 / 系統 API）
- Phase 2：用戶與會話（註冊登入 / JWT 認證 / 會話管理）
- Phase 3：核心對話（OpenAI 整合 / 對話功能 / 兩個預設角色）
- Phase 4：記憶系統（Redis / PostgreSQL / Qdrant / 檢索機制）
- Phase 5：模式系統（普通 / 小說 / NSFW 模式 / 標籤系統）
- Phase 6：NSFW 功能（Grok 整合 / 內容路由 / 觸發機制）
- Phase 7：情感系統（好感度 / 關係狀態 / 特殊事件）
- Phase 8：語音功能（OpenAI TTS / 角色語音 / API）
- Phase 9：測試優化（測試 / 優化 / Docker 容器化）

> 詳細清單見 `SPEC.md` 的「開發計劃」。後續將以單一權威來源維護，避免重複與矛盾。

## 貢獻
- 請先閱讀 `SPEC.md` 後開發，提出 PR 前可在 Issue 討論設計
- 待建立 `CONTRIBUTING.md` 與程式碼規範（lint/formatter/test）

## 授權
- Copyright © 2024 clarencetw
- 專有軟體，保留所有權利
