# 🤖 Thewavess AI Core

## 1. 專案摘要
| 項目 | 說明 |
|------|------|
| 核心定位 | 女性向情感聊天後端，支援 AI 雙引擎（OpenAI + Grok）與 5 級 NSFW 語意檢測 |
| 架構 | Go 1.23 · Gin · Bun · PostgreSQL |
| API 狀態 | 57 / 57 端點完成，詳見 [API_PROGRESS.md](./API_PROGRESS.md) |
| 測試狀態 | Shell 整合測試 24/24 (`./tests/test-all.sh`) |
| 重要文件 | 架構：[ARCHITECTURE.md](./ARCHITECTURE.md) · 部署：[DEPLOYMENT.md](./DEPLOYMENT.md) · 配置：[CONFIGURATION.md](./CONFIGURATION.md) |

## 2. 快速開始
| 步驟 | 指令 | 說明 |
|------|------|------|
| 安裝依賴 | `make install` | 安裝 Go 模組與 swagger 工具 |
| 複製設定 | `cp .env.example .env` | 至少填寫 `OPENAI_API_KEY`、資料庫連線、`JWT_SECRET` |
| 資料庫初始化 | `make db-setup` 或 `make fresh-start` | 建立資料表並載入 fixtures（含預設管理員/角色）|
| 啟動服務 | `make dev` | 生成 Swagger 並啟動 API（預設 8080）|
| 全套測試 | `./tests/test-all.sh` | 執行 24 項整合測試 |

常用端點：
- Web UI: http://localhost:8080/
- Swagger: http://localhost:8080/swagger/index.html
- 健康檢查: http://localhost:8080/health
- BasePath: `/api/v1`

## 3. 基礎 API 範例
```bash
# 註冊
curl -sS -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","email":"test@example.com","password":"Test123456!"}'

# 登入
curl -sS -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","password":"Test123456!"}'

# 建立對話
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

## 4. 目錄總覽
| 目錄 | 內容 |
|------|------|
| `handlers/` | HTTP Handler：auth、user、chat、character、relationships、search、monitor 等 |
| `services/` | 核心服務：chat、nsfw、character、tts、openai/grok 客戶端、prompt builder |
| `routes/` | 路由註冊（共 57 條 API）|
| `models/` | 資料模型與 API 響應結構 (`models/db` 為 Bun ORM 定義) |
| `cmd/bun/` | Bun 遷移與 fixtures 工具 |
| `tests/` | Shell 整合測試與共用腳本 |
| `public/` | 靜態頁面 / Swagger UI |

## 5. 常用 Make 指令
| 指令 | 功能 |
|------|------|
| `make dev` | 生成 Swagger 並啟動 API |
| `make fresh-start` | 清理 → 安裝 → 遷移 → Fixtures |
| `make quick-setup` | 只執行遷移與 fixtures |
| `make build` | 編譯可執行檔 (`bin/thewavess-ai-core`) |
| `make docker-build` | 建置 Docker 映像 |
| `make test` | 執行 Go 單元測試 |

## 6. 環境變數速覽
| 類別 | 是否必填 | 主要變數 |
|------|----------|-----------|
| AI 金鑰 | ✅ | `OPENAI_API_KEY`（若需 L4/L5 內容，補 `GROK_API_KEY`）|
| 資料庫 | ✅ | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` |
| JWT | ✅ | `JWT_SECRET`（管理端可另設 `ADMIN_JWT_SECRET`）|
| 伺服器 | 建議 | `PORT`, `GIN_MODE`, `LOG_LEVEL`, `API_HOST` |
| NSFW RAG | 建議 | `NSFW_CORPUS_DATA_PATH`, `NSFW_CORPUS_EMBEDDING_PATH`, `NSFW_RAG_LEVEL_THRESHOLDS` |
| 其他 | 選擇性 | `TTS_API_KEY`, `MISTRAL_API_KEY`（目前保留介面）|

詳見 [CONFIGURATION.md](./CONFIGURATION.md) 與 `.env.example`。

## 7. 延伸指南
| 文件 | 說明 |
|------|------|
| [ADMIN_SYSTEM.md](ADMIN_SYSTEM.md) | 管理端權限、端點與安全設定 |
| [AFFECTION_GUIDE.md](AFFECTION_GUIDE.md) | 好感度欄位與更新流程總覽 |
| [AGENTS.md](AGENTS.md) | 協助 AI 代理（如 Claude）理解專案規範 |
| [API_PROGRESS.md](API_PROGRESS.md) | 57 條 API 狀態與權限列表 |
| [ARCHITECTURE.md](ARCHITECTURE.md) | 系統架構、模組對應與技術棧 |
| [CHARACTER_GUIDE.md](CHARACTER_GUIDE.md) | 角色資料模型、設計建議與範例 |
| [CHAT_MODES.md](CHAT_MODES.md) | chat / novel 模式差異與使用方式 |
| [CLAUDE.md](CLAUDE.md) | Claude 代理作業指引 |
| [CONFIGURATION.md](CONFIGURATION.md) | 環境變數設定表與流程 |
| [DEPLOYMENT.md](DEPLOYMENT.md) | 部署步驟、端點檢查與排錯 |
| [DEVELOPMENT.md](DEVELOPMENT.md) | 開發流程、常用指令與測試範圍 |
| [GETTING_STARTED.md](GETTING_STARTED.md) | 快速入門教學與基礎 API 範例 |
| [MONITORING_GUIDE.md](MONITORING_GUIDE.md) | 監控指標、健康檢查與告警建議 |
| [NSFW_GUIDE.md](NSFW_GUIDE.md) | NSFW RAG 辨識與路由決策詳解 |
| [NSFW_RAG_GUIDE.md](NSFW_RAG_GUIDE.md) | NSFW 系統快速參考表 |
| [RELATIONSHIP_GUIDE.md](RELATIONSHIP_GUIDE.md) | 關係狀態 / 歷史 API 資料說明 |
| [SPEC.md](SPEC.md) | 產品規格與功能定位 |
