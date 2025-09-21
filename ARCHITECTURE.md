# 🏗️ 系統架構（Architecture）

本文件概述技術棧與高層架構，並對應到專案目錄結構。

—

## 技術棧

- 語言與框架：Go 1.23+、Gin
- 認證：JWT（Access/Refresh）
- 資料庫：PostgreSQL、Bun ORM、Go 遷移工具
- AI：OpenAI GPT-4o（L1–L3）、Grok（L4–L5）、智能路由選擇
- NSFW 分級：語意檢索 RAG 系統（預計算向量）
- 文件：Swagger/OpenAPI 3
- 日誌：結構化日誌
- 容器：Docker、Docker Compose

—

## 高層架構

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web UI        │    │  Mobile App     │    │   Third Party   │
│   (public/)     │    │                 │    │      API        │
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
          │             Business Logic                 │
          │ Chat / Character / Emotion / Memory / TTS │
          └─────────────────────┬─────────────────────┘
                                │
    ┌─────────────┬─────────────┴─────────────┬─────────────┐
    │             │                           │             │
┌───▼───┐    ┌───▼───┐    ┌─────▼─────┐  ┌───▼────┐    ┌───▼───┐
│OpenAI │    │ Grok  │    │NSFW RAG   │  │Postgres│    │(Cache)│
│L1-L3  │    │L4-L5  │    │語意檢索   │  │  Bun   │    │ Redis │
│GPT-4o │    │       │    │向量比對   │  │        │    │       │
└───────┘    └───────┘    └───────────┘  └────────┘    └───────┘
```

—

## 對應到專案目錄

- `handlers/`：HTTP handlers（`relationship.go`、`search.go` 等已轉為型別化回應，直接回傳資料庫真實欄位）
- `services/`：核心服務邏輯（NSFW、Chat、Character、AI 客戶端、記憶體管理）
- `routes/`：路由註冊入口 (`routes.go`)
- `models/`：資料模型與 API 響應定義
  - `models/db/`：資料庫表模型（User、Character、Chat、Relationship、Admin…）
- `cmd/bun/`：資料庫連線、遷移、種子資料
- `configs/nsfw/`：NSFW RAG 語料與向量（`corpus.json` + `embeddings.json`）
- `middleware/`：認證、日誌、CORS 等橫切關注
- `utils/`：日誌、錯誤、JWT、工具函式
- `public/`：靜態頁面與 Swagger UI
- `docs/`：由 `swag` 自動產出的 API 文件

—

## 系統特色

### 架構優化重構
- **API 響應簡化**：關係與搜尋 handlers 回傳嚴謹的 JSON 結構，移除舊有假資料欄位
- **資料真實性**：所有搜尋結果與情感統計都直接讀取資料庫欄位或 JSONB
- **語意檢索升級**：NSFW 模組採用預計算向量，查詢零額外 API 成本
- **Prompt 架構**：繼承式 Prompt Builder，依照引擎 (OpenAI/Grok) 客製

### 性能與可維護性
- **分離式文件**：`corpus.json` + `embeddings.json`，降低啟動成本
- **記憶體運算**：NSFW RAG 純記憶體比對，平均 8.5ms
- **程式碼清理**：移除未使用 Helper，統一時間與分頁處理工具
- **JSONB 優化**：`emotion_data` 持久化歷史事件，篩選效率佳
- **資料庫一致性**：統一欄位命名，搜尋/關係 handler 使用相同查詢來源
- **維護工具**：`make nsfw-embeddings`、`make nsfw-check` 確保語料與向量同步

## 其他說明

- **API 狀態**：57 個端點（100% 完成），詳見 `API_PROGRESS.md` 與 Swagger
- **配置管理**：完整環境變數說明請參考 `CONFIGURATION.md` 與 `.env.example`
- **測試覆蓋**：24/24 核心測試 + 6 組 Shell 測試腳本
- **持續重構**：Relationship 與 Search handler 皆已改為型別化結構，便於日後擴充
