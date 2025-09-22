# 📚 Thewavess AI Core - 完整文檔索引

> **📋 總覽**: 這是 Thewavess AI Core 項目的完整文檔索引，包含所有詳細配置、開發指南與系統說明。快速入門請參考 [README.md](./README.md)。

## 🎯 專案摘要

| 項目 | 說明 |
|------|------|
| **核心定位** | 女性向情感聊天後端，支援 AI 雙引擎（OpenAI + Grok）與 5 級 NSFW 語意檢測 |
| **技術架構** | Go 1.23 · Gin · Bun ORM · PostgreSQL · Redis · Docker |
| **API 狀態** | ✅ 57 / 57 端點完成，詳見 [API_PROGRESS.md](./API_PROGRESS.md) |
| **測試狀態** | ✅ Shell 整合測試 24/24 (`./tests/test-all.sh`) |
| **部署狀態** | 🚀 生產就緒，支援 Docker 容器化部署 |

## 📖 完整文檔列表

### 🏗️ 系統架構與設計
| 文檔 | 說明 | 適用對象 |
|------|------|----------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | 系統架構、模組對應與技術棧詳解 | 架構師、後端開發 |
| [SPEC.md](./SPEC.md) | 產品規格與功能定位 | 產品經理、開發團隊 |
| [AGENTS.md](./AGENTS.md) | 協助 AI 代理（如 Claude）理解專案規範 | AI 輔助開發 |

### ⚙️ 配置與部署
| 文檔 | 說明 | 適用對象 |
|------|------|----------|
| [CONFIGURATION.md](./CONFIGURATION.md) | 環境變數設定表與詳細流程 | DevOps、系統管理員 |
| [DEPLOYMENT.md](./DEPLOYMENT.md) | 部署步驟、端點檢查與排錯指南 | DevOps、運維團隊 |
| [GETTING_STARTED.md](./GETTING_STARTED.md) | 快速入門教學與基礎 API 範例 | 新手開發者 |

### 🔧 開發與測試
| 文檔 | 說明 | 適用對象 |
|------|------|----------|
| [DEVELOPMENT.md](./DEVELOPMENT.md) | 開發流程、常用指令與測試範圍 | 後端開發者 |
| [API_PROGRESS.md](./API_PROGRESS.md) | 57 條 API 狀態與權限列表 | 前端、後端開發 |
| [CLAUDE.md](./CLAUDE.md) | Claude AI 代理作業指引與項目規範 | AI 輔助開發 |

### 🤖 AI 系統與功能
| 文檔 | 說明 | 適用對象 |
|------|------|----------|
| [NSFW_GUIDE.md](./NSFW_GUIDE.md) | NSFW RAG 辨識與路由決策詳解 | AI 工程師、後端開發 |
| [NSFW_RAG_GUIDE.md](./NSFW_RAG_GUIDE.md) | NSFW 系統快速參考表與故障排除 | 運維、開發者 |
| [CHARACTER_GUIDE.md](./CHARACTER_GUIDE.md) | 角色資料模型、設計建議與範例 | 內容設計、後端開發 |
| [CHAT_MODES.md](./CHAT_MODES.md) | chat / novel 模式差異與使用方式 | 前端、後端開發 |

### 👥 用戶與關係管理
| 文檔 | 說明 | 適用對象 |
|------|------|----------|
| [RELATIONSHIP_GUIDE.md](./RELATIONSHIP_GUIDE.md) | 關係狀態 / 歷史 API 資料說明 | 前端、後端開發 |
| [AFFECTION_GUIDE.md](./AFFECTION_GUIDE.md) | 好感度欄位與更新流程總覽 | 遊戲設計、後端開發 |

### 🔐 管理與監控
| 文檔 | 說明 | 適用對象 |
|------|------|----------|
| [ADMIN_SYSTEM.md](./ADMIN_SYSTEM.md) | 管理端權限、端點與安全設定 | 系統管理員、安全團隊 |
| [MONITORING_GUIDE.md](./MONITORING_GUIDE.md) | 監控指標、健康檢查與告警建議 | DevOps、運維團隊 |

## 🚀 快速導航

### 📋 按角色分類

**🆕 新手開發者:**
1. [README.md](./README.md) - 項目概覽
2. [GETTING_STARTED.md](./GETTING_STARTED.md) - 快速入門
3. [DEVELOPMENT.md](./DEVELOPMENT.md) - 開發流程

**👨‍💻 後端開發者:**
1. [ARCHITECTURE.md](./ARCHITECTURE.md) - 系統架構
2. [API_PROGRESS.md](./API_PROGRESS.md) - API 狀態
3. [CHARACTER_GUIDE.md](./CHARACTER_GUIDE.md) - 角色系統
4. [RELATIONSHIP_GUIDE.md](./RELATIONSHIP_GUIDE.md) - 關係管理

**🤖 AI 工程師:**
1. [NSFW_GUIDE.md](./NSFW_GUIDE.md) - NSFW 分類系統
2. [NSFW_RAG_GUIDE.md](./NSFW_RAG_GUIDE.md) - RAG 快速參考
3. [CLAUDE.md](./CLAUDE.md) - AI 輔助開發

**⚙️ DevOps/運維:**
1. [CONFIGURATION.md](./CONFIGURATION.md) - 環境配置
2. [DEPLOYMENT.md](./DEPLOYMENT.md) - 部署指南
3. [MONITORING_GUIDE.md](./MONITORING_GUIDE.md) - 監控告警

**🔐 系統管理員:**
1. [ADMIN_SYSTEM.md](./ADMIN_SYSTEM.md) - 管理後台
2. [MONITORING_GUIDE.md](./MONITORING_GUIDE.md) - 系統監控

### 🎯 按功能分類

**🛠️ 環境搭建與配置:**
- [GETTING_STARTED.md](./GETTING_STARTED.md) - 基礎搭建
- [CONFIGURATION.md](./CONFIGURATION.md) - 詳細配置
- [DEPLOYMENT.md](./DEPLOYMENT.md) - 生產部署

**🤖 AI 功能:**
- [NSFW_GUIDE.md](./NSFW_GUIDE.md) - 內容分級
- [CHARACTER_GUIDE.md](./CHARACTER_GUIDE.md) - 角色系統
- [CHAT_MODES.md](./CHAT_MODES.md) - 對話模式

**👥 用戶管理:**
- [RELATIONSHIP_GUIDE.md](./RELATIONSHIP_GUIDE.md) - 關係系統
- [AFFECTION_GUIDE.md](./AFFECTION_GUIDE.md) - 好感度機制
- [ADMIN_SYSTEM.md](./ADMIN_SYSTEM.md) - 管理功能

**🔧 開發運維:**
- [DEVELOPMENT.md](./DEVELOPMENT.md) - 開發規範
- [API_PROGRESS.md](./API_PROGRESS.md) - API 文檔
- [MONITORING_GUIDE.md](./MONITORING_GUIDE.md) - 監控運維

## 📝 目錄結構總覽

```
├── handlers/          # HTTP Handler：auth、user、chat、character、relationships、search、monitor 等
├── services/          # 核心服務：chat、nsfw、character、tts、openai/grok 客戶端、prompt builder
├── routes/           # 路由註冊（共 57 條 API）
├── models/           # 資料模型與 API 響應結構 (models/db 為 Bun ORM 定義)
├── cmd/bun/          # Bun 遷移與 fixtures 工具
├── tests/            # Shell 整合測試與共用腳本
├── public/           # 靜態頁面 / Swagger UI
├── docs/             # 自動生成的 API 文檔
└── *.md              # 各類文檔與指南
```

## 🔧 常用 Make 指令速查

| 指令 | 功能 | 使用時機 |
|------|------|----------|
| `make dev` | 生成 Swagger 並啟動 API | 日常開發 |
| `make fresh-start` | 清理 → 安裝 → 遷移 → Fixtures | 初次搭建/重置 |
| `make quick-setup` | 只執行遷移與 fixtures | 快速重置數據 |
| `make build` | 編譯可執行檔 (`bin/thewavess-ai-core`) | 打包部署 |
| `make docker-build` | 建置 Docker 映像 | 容器化部署 |
| `make test` | 執行 Go 單元測試 | 代碼驗證 |
| `make docs` | 重新生成 Swagger 文檔 | API 更新後 |
| `make nsfw-embeddings` | 更新 NSFW 向量嵌入 | NSFW 語料更新後 |

## 🔑 環境變數速查表

| 類別 | 是否必填 | 主要變數 | 說明 |
|------|----------|-----------|------|
| **AI 金鑰** | ✅ | `OPENAI_API_KEY` | OpenAI GPT-4o 必需 |
| | 建議 | `GROK_API_KEY` | L4/L5 內容需要 |
| **資料庫** | ✅ | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | PostgreSQL 連線 |
| **JWT** | ✅ | `JWT_SECRET` | 用戶認證簽名 |
| | 選擇性 | `ADMIN_JWT_SECRET` | 管理端獨立簽名 |
| **伺服器** | 建議 | `PORT`, `GIN_MODE`, `LOG_LEVEL`, `API_HOST` | 服務配置 |
| **NSFW RAG** | 建議 | `NSFW_CORPUS_DATA_PATH`, `NSFW_CORPUS_EMBEDDING_PATH` | RAG 分類器路徑 |
| | | `NSFW_RAG_LEVEL_THRESHOLDS` | L1-L5 閾值設定 |
| **其他** | 選擇性 | `TTS_API_KEY` | OpenAI TTS 語音 |
| | | `MISTRAL_API_KEY` | 保留介面 |

詳細說明請參考 [CONFIGURATION.md](./CONFIGURATION.md) 與 `.env.example`。

---

**💡 提示**: 這個文檔索引會隨著項目發展持續更新。如果您找不到需要的信息，請參考對應角色或功能分類中的相關文檔。