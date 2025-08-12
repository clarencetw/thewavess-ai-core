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

## 快速開始（HTTP API 雛形階段）
> 專案已完成 HTTP API 雛形架構，包含 22 個端點的 Swagger 文檔，但業務邏輯尚未實現。

1. **安裝 Go 1.22+** 
2. **安裝 swag 工具**（用於生成 OpenAPI 文檔）
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   ```
3. **安裝依賴並生成文檔**
   ```bash
   go mod tidy
   make docs  # 生成 Swagger 文檔
   ```
4. **啟動 API 伺服器**
   ```bash
   go run main.go
   ```
5. **訪問 Swagger UI**
   ```
   http://localhost:8080/swagger/index.html
   ```

### 🔧 開發指令
```bash
make docs    # 生成 OpenAPI 文檔
make build   # 編譯應用程式
make clean   # 清理生成的檔案
```

詳細 API 端點說明請參考 [API 文檔](./API.md)

## API 文檔

### 📚 完整文檔
- **[API.md](./API.md)** - 完整 API 端點文檔，包含請求/回應範例和快速開始
- **[Swagger UI](http://localhost:8080/swagger/index.html)** - 🔗 互動式 API 文檔 (啟動服務後可用)  
- **[SPEC.md](./SPEC.md)** - 產品規格與技術架構

### 🎯 自動生成 OpenAPI 規格
現在使用程式碼註解自動生成 OpenAPI 文檔：
```bash
make docs  # 自動生成 docs/swagger.json 和 docs/swagger.yaml
```

### 🔗 端點概覽  
核心 API 模組設計 (**目前已實現 22 個端點，規劃 118 個端點**)：

| 模組 | 端點範例 | 實現狀態 | 功能說明 |
|------|----------|----------|----------|
| **用戶管理** | `/user/*` | ✅ 7/9 | 註冊、登入、個人資料管理 |
| **角色系統** | `/character/*`, `/user/character` | ✅ 3/5 | 角色列表、選擇、統計數據 |
| **對話管理** | `/chat/*` | ✅ 10/11 | 會話創建、訊息發送、歷史記錄 |
| **系統管理** | `/health`, `/version`, `/status` | ✅ 2/3 | 服務監控、版本資訊 |
| **小說模式** | `/novel/*` | ❌ 0/7 | 互動小說、劇情分支、進度管理 |
| **情感系統** | `/emotion/*` | ❌ 0/5 | 好感度、關係狀態、事件觸發 |
| **記憶系統** | `/memory/*` | ❌ 0/8 | 記憶檢索、時間線、智能搜尋 |
| **語音合成** | `/tts/*` | ❌ 0/5 | 文字轉語音、語音預覽、批量生成 |
| **標籤系統** | `/tags/*` | ❌ 0/4 | 內容標籤、熱門標籤、標籤管理 |
| **檔案上傳** | `/upload/*` | ❌ 0/3 | 圖片、語音檔案上傳 |
| **通知系統** | `/notifications/*` | ❌ 0/4 | 通知管理、設定 |
| **分析統計** | `/analytics/*` | ❌ 0/3 | 用戶統計、對話分析 |

> ✅ = 已實現 API 雛形（但業務邏輯返回 NotImplemented）  
> ❌ = 尚未實現

### 🔐 認證方式
- **JWT Bearer Token** - 所有 API 請求需要認證
- **請求格式** - `Authorization: Bearer <token>`
- **內容類型** - `application/json`

## 安全與合規
- 面向成年用戶，完全開放 NSFW 內容
- 前端負責年齡驗證與內容分級
- 違法內容一律禁止
- 敏感資料加密儲存

## 開發狀態與路線圖

### 🚀 **當前狀態**
- ✅ **Phase 1**: HTTP API 雛形架構完成（22個端點，Swagger 自動生成）
- 🔄 **Phase 2**: 開始實現業務邏輯（用戶管理、對話管理）

### 📋 **接下來的開發順序**
1. **實現核心業務邏輯**（2-3週）
   - 完成現有 22 個端點的實際功能
   - 資料庫連接與 CRUD 操作
   - JWT 認證與授權機制

2. **AI 核心功能**（4-6週）  
   - OpenAI GPT-4o 整合
   - 情感系統（好感度、關係狀態）
   - 記憶系統（Redis + PostgreSQL + Qdrant）

3. **進階功能**（6-8週）
   - 小說模式與劇情分支
   - Grok 整合（NSFW 內容）
   - TTS 語音合成
   - 檔案上傳與處理

4. **完善與優化**（2-4週）
   - 分析統計功能
   - 效能優化與測試
   - Docker 容器化部署

> 詳細清單見 `SPEC.md` 的「開發計劃」。後續將以單一權威來源維護，避免重複與矛盾。

## 貢獻
- 請先閱讀 `SPEC.md` 後開發，提出 PR 前可在 Issue 討論設計
- 待建立 `CONTRIBUTING.md` 與程式碼規範（lint/formatter/test）

## 授權
- Copyright © 2024 clarencetw
- 專有軟體，保留所有權利
