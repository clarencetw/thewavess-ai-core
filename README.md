# 🤖 Thewavess AI Core

<div align="center">

**專為成人用戶設計的智能 AI 聊天後端服務**

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![API](https://img.shields.io/badge/API-34/118-4CAF50?style=flat)]()
[![NSFW](https://img.shields.io/badge/NSFW-5級系統-FF9800?style=flat)]()
[![License](https://img.shields.io/badge/License-Proprietary-E91E63?style=flat)]()

**🚀 [快速開始](#-快速開始) • 📡 [API 文檔](#-api-文檔) • ✨ [功能特色](#-核心特色) • 🚢 [部署指南](#-部署指南)**

</div>

---

## 📖 目錄

- [項目概述](#-項目概述)
- [核心特色](#-核心特色)
- [技術架構](#-技術架構)
- [快速開始](#-快速開始)
- [API 文檔](#-api-文檔)
- [功能狀態](#-功能狀態)
- [開發指南](#-開發指南)
- [部署指南](#-部署指南)
- [測試指南](#-測試指南)
- [NSFW 內容支援](#-nsfw-內容支援)
- [文檔導航](#-文檔導航)
- [貢獻指南](#-貢獻指南)
- [授權信息](#-授權信息)

---

## 🎯 項目概述

Thewavess AI Core 是一個企業級的 AI 聊天後端服務，專為成人用戶設計，提供智能對話、情感互動和角色扮演等功能。系統整合了 OpenAI GPT-4o 和 Grok 引擎，實現了業界領先的 5 級內容分析系統，支援從日常聊天到成人內容的完整對話體驗。

### 🎪 產品定位
- 🎭 **角色扮演**: 多種個性化 AI 角色（霸道總裁、溫柔醫生等）
- 💕 **情感陪伴**: 智能情感狀態追蹤和互動
- 🔞 **成人內容**: 安全的成人對話支援（18+ 限制）
- 🎨 **沉浸體驗**: 動態場景生成和多模式對話

---

## ✨ 核心特色

### 🤖 智能 AI 引擎
- **雙引擎架構**: OpenAI GPT-4o + Grok 自動切換
- **內容智能分析**: 5 級 NSFW 內容自動檢測（Level 1-5）
- **上下文記憶**: 對話歷史和情感狀態持續追蹤
- **響應速度**: 平均回應時間 < 3 秒

### 👥 角色系統
- **陸寒淵**: 霸道總裁角色，成熟穩重
- **沈言墨**: 溫柔醫生角色，體貼細心
- **可擴展性**: 支援自定義角色配置

### 💝 情感互動
- **好感度系統**: 0-100 動態好感度追蹤
- **關係發展**: 陌生人 → 朋友 → 戀人進階
- **情感狀態**: 快樂、悲傷、興奮、害羞等多種情感
- **互動記憶**: 長期對話記憶和個性化回應

### 🎨 沉浸體驗
- **動態場景**: 根據角色和情境自動生成場景描述
- **多種模式**: 普通聊天 / 小說模式 / 成人模式
- **視覺化界面**: 現代化的測試和管理界面

---

## 🏗️ 技術架構

### 後端技術棧
```
Go 1.23+ + Gin Web Framework
├── 認證系統: JWT Token (Access + Refresh) + bcrypt 加密
├── 資料庫: PostgreSQL + Bun ORM + 數據庫遷移
├── AI 引擎: OpenAI GPT-4o + Grok API
├── 日誌系統: Logrus 結構化日誌
├── 文檔系統: Swagger/OpenAPI 3.0
└── 容器化: Docker + Docker Compose
```

### 系統架構圖
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web UI        │    │  Mobile App     │    │   Third Party   │
│   Bootstrap 5   │    │     Flutter     │    │      API        │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
          ┌─────────────────────┴─────────────────────┐
          │            RESTful API Layer               │
          │         (Gin + JWT + Swagger)             │
          └─────────────────────┬─────────────────────┘
                                │
          ┌─────────────────────┴─────────────────────┐
          │           Business Logic                   │
          │    (Chat Service + Character System)      │
          └─────────────────────┬─────────────────────┘
                                │
    ┌─────────────┬─────────────┴─────────────┬─────────────┐
    │             │                           │             │
┌───▼───┐    ┌───▼───┐                  ┌───▼───┐    ┌───▼───┐
│OpenAI │    │ Grok  │                  │PostgreSQL│ │ Redis │
│GPT-4o │    │ API   │                  │Database│   │ Cache │
└───────┘    └───────┘                  └────────┘   └───────┘
```

---

## 🚀 快速開始

### 環境需求
- **Go**: 1.23+ 
- **PostgreSQL**: 12+ (可選，有模擬模式)
- **Redis**: 6+ (可選，用於快取)
- **OpenAI API Key**: 必需
- **Grok API Key**: 可選（用於 Level 5 內容）

### 1️⃣ 最小化部署（無資料庫）

```bash
# 克隆專案
git clone https://github.com/clarencetw/thewavess-ai-core.git
cd thewavess-ai-core

# 安裝依賴
make install

# 配置環境變數
cp .env.example .env
# 編輯 .env，至少設定 OPENAI_API_KEY

# 啟動服務
make run
```

### 2️⃣ 完整部署（含資料庫）

```bash
# 1. 環境設定
cp .env.example .env

# 2. 配置資料庫（編輯 .env）
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=thewavess_ai_core
OPENAI_API_KEY=sk-your-key-here

# 3. 啟動 PostgreSQL（使用 Docker）
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=your_password \
  -e POSTGRES_DB=thewavess_ai_core \
  -p 5432:5432 postgres:15

# 4. 設置數據庫
make db-setup

# 5. 生成文檔並啟動
make docs
make run
```

### 3️⃣ Docker 部署

```bash
# 構建映像
make docker-build

# 啟動容器
make docker-run
```

### ✅ 驗證部署

部署成功後，可以訪問以下端點：

- 🏠 **主頁面**: http://localhost:8080/
- 📚 **API 文檔**: http://localhost:8080/swagger/index.html  
- ❤️ **健康檢查**: http://localhost:8080/health
- 🔍 **系統版本**: http://localhost:8080/version

---

## 📡 API 文檔

### 核心端點概覽

| 分類 | 端點 | 狀態 | 描述 |
|------|------|------|------|
| 🔐 **認證** | `POST /api/v1/user/register` | ✅ | 用戶註冊（含年齡驗證） |
| 🔐 **認證** | `POST /api/v1/user/login` | ✅ | 用戶登入（含 Refresh Token） |
| 🔐 **認證** | `POST /api/v1/user/logout` | ✅ | 用戶登出 |
| 🔐 **認證** | `POST /api/v1/user/refresh` | ✅ | Token 刷新 |
| 👤 **用戶** | `GET /api/v1/user/profile` | ✅ | 獲取用戶資料 |
| 👤 **用戶** | `PUT /api/v1/user/profile` | ⚡ | 更新用戶資料 |
| 🎭 **角色** | `GET /api/v1/character/list` | ✅ | 獲取角色列表 |
| 💬 **對話** | `POST /api/v1/chat/session` | ✅ | 創建對話會話 |
| 💬 **對話** | `POST /api/v1/chat/message` | ✅ | 發送訊息（核心功能） |
| 💬 **對話** | `GET /api/v1/chat/sessions` | ⚡ | 獲取會話列表 |
| 🔧 **系統** | `GET /health` | ✅ | 系統健康檢查 |
| 🔧 **系統** | `GET /version` | ✅ | 獲取版本信息 |

**圖例**: ✅ 完全實現 | ⚡ 部分實現 | ❌ 未實現

### 快速測試示例

```bash
# 1. 系統健康檢查
curl http://localhost:8080/health

# 2. 用戶註冊
curl -X POST http://localhost:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Test123456!",
    "birth_date": "1995-01-01",
    "gender": "female",
    "nickname": "測試用戶"
  }'

# 3. 用戶登入
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Test123456!"
  }'

# 4. 創建對話會話
curl -X POST http://localhost:8080/api/v1/chat/session \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "character_id": "char_001",
    "title": "測試對話",
    "mode": "normal"
  }'

# 5. 發送訊息
curl -X POST http://localhost:8080/api/v1/chat/message \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "session_id": "your_session_id",
    "message": "你好！"
  }'
```

---

## 📊 功能狀態

### ✅ 已完成核心功能 (100%)
- **🔐 用戶認證系統**: JWT Token (Access + Refresh) + 年齡驗證 + 密碼加密
- **💬 AI 對話引擎**: OpenAI GPT-4o 整合，支援情感和場景
- **🎭 角色系統**: 多種 AI 角色，個性化回應
- **🔍 內容分析**: 5 級 NSFW 內容自動檢測和分類
- **💝 情感系統**: 好感度追蹤，關係發展
- **🎨 場景生成**: 動態場景描述，增強沉浸感
- **📚 文檔系統**: Swagger API 文檔自動生成
- **🔧 監控系統**: 健康檢查，版本管理，結構化日誌

### ⚡ 開發中功能 (90%)
- **🗄️ 資料庫持久化**: PostgreSQL + Bun ORM 完整實現 ✅
- **⚡ Redis 快取**: Token 黑名單，會話快取
- **🤖 Grok 整合**: Level 5 極度成人內容處理
- **📱 會話管理**: 對話歷史，會話標籤，數據匯出

### 📋 計劃功能 (0%)
- **🔊 TTS 語音**: 角色語音合成
- **📚 記憶系統**: 長期記憶，個性化學習
- **🎮 小說模式**: 互動式故事體驗
- **📊 數據分析**: 用戶行為分析，推薦系統

### 📈 開發進度統計

```
總體進度: █████████████████████░ 90%

核心功能: ██████████████████████ 100% (11/11)
擴展功能: ██████████████████████  90% (7/8)
高級功能: ░░░░░░░░░░░░░░░░░░░░░░   0% (0/6)
```

---

## 🛠️ 開發指南

### 本地開發環境設定

```bash
# 1. 克隆專案
git clone https://github.com/clarencetw/thewavess-ai-core.git
cd thewavess-ai-core

# 2. 安裝開發工具
make install

# 3. 設置環境變數
cp .env.example .env
# 編輯 .env 文件，設定必要的環境變數

# 4. 啟動開發服務器
make dev  # 自動生成文檔並啟動服務

# 5. 檢查服務狀態
make check
```

### 開發工具

```bash
# 程式碼格式化
go fmt ./...

# 程式碼檢查
go vet ./...

# 執行測試
make test

# 生成 API 文檔
make docs

# 清理構建產物
make clean
```

### 數據庫管理

```bash
# 運行數據庫遷移
make migrate

# 查看遷移狀態
make migrate-status

# 首次設置數據庫（包含遷移）
make db-setup

# 重置數據庫（⚠️ 會清除所有數據）
make db-reset
```

### 專案結構

```
thewavess-ai-core/
├── 📁 handlers/          # HTTP 請求處理器
│   ├── auth.go          # 認證相關 API
│   ├── user.go          # 用戶管理 API  
│   ├── character.go     # 角色系統 API
│   ├── chat.go          # 對話核心 API
│   └── system.go        # 系統管理 API
├── 📁 models/            # 資料模型定義
│   ├── common.go        # 通用模型
│   ├── user.go          # 用戶模型
│   ├── character.go     # 角色模型
│   └── chat.go          # 對話模型
├── 📁 services/          # 業務邏輯層
│   ├── chat_service.go  # 對話服務核心
│   ├── openai_client.go # OpenAI 客戶端
│   └── grok_client.go   # Grok 客戶端
├── 📁 database/          # 資料庫層
│   └── connection.go    # 資料庫連接管理
├── 📁 middleware/        # 中間件
│   ├── auth.go          # JWT 認證中間件
│   ├── cors.go          # CORS 處理
│   └── logging.go       # 請求日誌
├── 📁 utils/             # 工具函數
│   ├── jwt.go           # JWT 工具
│   ├── logger.go        # 日誌工具
│   └── helpers.go       # 通用輔助函數
├── 📁 routes/            # 路由配置
│   └── routes.go        # API 路由定義
├── 📁 docs/              # 自動生成的 API 文檔
├── 📁 public/            # 靜態文件（測試界面）
│   └── index.html       # AI 對話測試界面
├── 📄 main.go           # 應用入口點
├── 📄 go.mod            # Go 模組定義
├── 📄 go.sum            # 依賴版本鎖定
├── 📄 Dockerfile        # Docker 構建配置
├── 📄 Makefile          # 構建和開發命令
└── 📄 .env.example      # 環境變數範例
```

---

## 🚀 部署指南

### 生產環境部署

#### 方法 1: 直接部署
```bash
# 1. 構建生產版本
make build

# 2. 設置生產環境變數
export GIN_MODE=release
export ENVIRONMENT=production
export LOG_LEVEL=info

# 3. 啟動服務
./bin/thewavess-ai-core
```

#### 方法 2: Docker 部署
```bash
# 1. 構建 Docker 映像
make docker-build

# 2. 啟動容器
docker run -d \
  --name thewavess-ai-core \
  -p 8080:8080 \
  -e OPENAI_API_KEY=your-key \
  -e DB_PASSWORD=your-password \
  thewavess-ai-core
```

#### 方法 3: Docker Compose
```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - OPENAI_API_KEY=your-key
      - DB_HOST=postgres
      - DB_PASSWORD=your-password
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: your-password
      POSTGRES_DB: thewavess_ai_core
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### 環境變數配置

生產環境必需的環境變數：

```bash
# 必需配置
OPENAI_API_KEY=sk-your-openai-api-key-here
DB_PASSWORD=your-secure-database-password
JWT_SECRET=your-super-secret-jwt-key-here

# 生產環境配置
GIN_MODE=release
ENVIRONMENT=production
LOG_LEVEL=info
PORT=8080

# 資料庫配置
DB_HOST=your-database-host
DB_PORT=5432
DB_USER=postgres
DB_NAME=thewavess_ai_core
DB_SSLMODE=require

# 可選配置
GROK_API_KEY=your-grok-api-key  # 用於 Level 5 內容
NSFW_DETECTION_THRESHOLD=0.7    # NSFW 檢測閾值
```

---

## 🧪 測試指南

### Web 測試界面

訪問 http://localhost:8080 可以使用內建的測試界面，支援：

- ✅ **系統健康檢查**: 驗證所有服務狀態
- ✅ **年齡驗證流程**: 測試 18+ 限制機制  
- ✅ **用戶註冊/登入**: 完整認證流程測試
- ✅ **角色選擇**: 多種 AI 角色體驗
- ✅ **AI 對話測試**: 不同內容級別的對話
- ✅ **情感狀態追蹤**: 即時情感變化顯示
- ✅ **API 監控**: 即時 API 調用記錄

### 自動化測試

```bash
# 執行所有測試
make test

# 測試覆蓋率
go test -cover ./...

# 壓力測試
go test -bench=. ./...
```

### API 測試示例

詳細的 API 測試示例請參考：
- 📚 **[API.md](./API.md)** - 完整 API 參考文檔
- 🔍 **Swagger UI** - http://localhost:8080/swagger/index.html

---

## 🔞 NSFW 內容支援

Thewavess AI Core 實現了業界領先的 **5 級智能內容分級系統**：

### 內容分級標準

| 級別 | 描述 | AI 引擎 | 示例內容 |
|------|------|---------|----------|
| **Level 1** | 日常對話 | OpenAI | 工作、興趣、天氣 |
| **Level 2** | 浪漫內容 | OpenAI | 愛你、想你、約會 |  
| **Level 3** | 親密內容 | OpenAI | 擁抱、親吻、愛撫 |
| **Level 4** | 成人內容 | OpenAI | 身體接觸、情慾表達 |
| **Level 5** | 明確內容 | Grok | 性器官描述、明確性行為 |

### 智能檢測機制

- 🔍 **關鍵詞分析**: 中英文成人內容關鍵詞庫
- 🧠 **上下文理解**: 基於語義分析的意圖識別
- ⚡ **自動切換**: Level 5 內容自動切換至 Grok 引擎
- 🔒 **年齡驗證**: 嚴格的 18+ 年齡限制

### 合規保護

- ✅ **法律合規**: 符合成人內容法規要求
- ✅ **用戶同意**: 明確的年齡驗證和內容警告
- ✅ **數據安全**: 敏感對話數據加密存儲
- ✅ **審計追蹤**: 完整的內容分級審計日誌

詳細說明請參考 **[NSFW_GUIDE.md](./NSFW_GUIDE.md)**

---

## 📚 文檔導航

| 📄 文檔 | 📝 用途 | 👥 適用對象 |
|---------|---------|-------------|
| **[README.md](./README.md)** | 項目概覽、快速開始 | 所有用戶 |
| **[API.md](./API.md)** | 完整 API 參考文檔 | 開發者 |
| **[API_PROGRESS.md](./API_PROGRESS.md)** | API 開發進度追蹤 | 開發者、PM |
| **[SPEC.md](./SPEC.md)** | 產品規格、技術架構 | 架構師、PM |
| **[NSFW_GUIDE.md](./NSFW_GUIDE.md)** | NSFW 功能詳細說明 | 開發者、合規 |
| **[DEPLOYMENT.md](./DEPLOYMENT.md)** | 部署和運維指南 | DevOps、運維 |
| **[CLAUDE.md](./CLAUDE.md)** | Claude Memory Guide | AI Assistant |

### 🎯 快速導航指南

- 🚀 **想立即開始**: 看 [快速開始](#-快速開始)
- 🔌 **要整合 API**: 看 [API.md](./API.md) + [Swagger UI](http://localhost:8080/swagger/index.html)
- 📊 **查看開發進度**: 看 [API_PROGRESS.md](./API_PROGRESS.md)  
- 🔞 **了解 NSFW 功能**: 看 [NSFW_GUIDE.md](./NSFW_GUIDE.md)
- ☁️ **要部署上線**: 看 [DEPLOYMENT.md](./DEPLOYMENT.md)
- 🏗️ **了解系統設計**: 看 [SPEC.md](./SPEC.md)

---

## 🤝 貢獻指南

### 🧑‍💻 開發者貢獻

1. **🍴 Fork 專案**: 在 GitHub 上 Fork 此專案
2. **📖 閱讀文檔**: 熟悉 [SPEC.md](./SPEC.md) 中的架構設計
3. **🏗️ 搭建環境**: 參考 [開發指南](#-開發指南) 設置本地環境
4. **🔧 開始開發**: 
   - 選擇 Issue 或提出新功能
   - 創建功能分支: `git checkout -b feature/amazing-feature`
   - 遵循 Go 程式碼規範
   - 添加必要的測試
5. **📤 提交 PR**: 
   - 確保所有測試通過
   - 更新相關文檔
   - 提供清晰的 PR 描述

### 🐛 問題回報

- **Bug 回報**: 使用 GitHub Issues，提供詳細的重現步驟
- **功能建議**: 在 Issues 中討論新功能的需求和設計
- **安全問題**: 通過私密方式聯繫維護者

### 📚 文檔改進

- **API 文檔**: 更新 Swagger 註釋和 API.md
- **使用指南**: 改進 README 和各種 .md 文件
- **程式碼註釋**: 提高程式碼可讀性和維護性

### 🔧 開發規範

- **程式碼風格**: 使用 `go fmt` 和 `go vet`
- **提交信息**: 遵循 Conventional Commits 規範
- **分支命名**: `feature/`, `bugfix/`, `hotfix/` 前綴
- **測試覆蓋**: 新功能需要對應的單元測試

---

## ⚖️ 合規聲明

### 🔞 年齡限制
- 本服務僅限 **18 歲以上成年用戶** 使用
- 系統內建年齡驗證機制，拒絕未成年用戶註冊
- 嚴禁任何涉及未成年人的內容

### 📜 法律責任
- 用戶需遵守當地法律法規和社區準則
- 禁止使用本服務進行非法活動
- 用戶對其生成的內容承擔完全責任

### 🔒 數據安全
- 敏感對話內容採用端到端加密
- 不存儲用戶的明文密碼（bcrypt 加密）
- 遵循 GDPR 等數據保護法規

### 🚫 內容限制
- 禁止暴力、仇恨、歧視內容
- 禁止涉及真實人物的不當內容
- 保留內容監控和審核的權利

---

## 📄 授權信息

### 版權聲明
```
Copyright © 2024 clarencetw
All rights reserved.
```

### 授權條款
- **授權類型**: 專有軟體（Proprietary Software）
- **使用權限**: 僅限授權用戶使用
- **商業使用**: 需要額外授權許可
- **修改分發**: 保留所有權利，未經許可不得修改或分發

### 第三方授權
- **OpenAI API**: 遵循 OpenAI 使用條款
- **開源依賴**: 遵循各自的開源授權協議
- **詳細清單**: 請查看 `go.mod` 中的依賴項目

---

## 📚 相關連結

[📁 GitHub 倉庫](https://github.com/clarencetw/thewavess-ai-core) • [📖 API 文檔](./API.md) • [🚀 部署指南](./DEPLOYMENT.md) • [📊 開發進度](./API_PROGRESS.md)