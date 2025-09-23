# 🏗️ 系統架構（Architecture）

本文件概述技術棧與高層架構，並對應到專案目錄結構。

> 📋 **相關文檔**: 完整文檔索引請參考 [DOCS_INDEX.md](./DOCS_INDEX.md)

---

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
┌─────────────────────────────────────────────────────────────────┐
│                        客戶端層 (Client Layer)                     │
├─────────────────┬─────────────────┬─────────────────────────────┤
│     Web UI      │   Mobile App    │     Third Party API         │
│   (public/)     │                 │                             │
└─────────────────┴─────────────────┴─────────────────────────────┘
                                │
┌─────────────────────────────────┴─────────────────────────────────┐
│                     API 層 (API Layer)                          │
│              Gin + JWT + Swagger + CORS                        │
└─────────────────────────────────┬─────────────────────────────────┘
                                │
┌─────────────────────────────────┴─────────────────────────────────┐
│                  業務邏輯層 (Business Logic)                      │
│   Chat Service │ Character Service │ Relationship Service        │
│   TTS Service  │ NSFW Classifier   │ Admin Service               │
└─────────────────────────────────┬─────────────────────────────────┘
                                │
┌─────────────────────────────────┴─────────────────────────────────┐
│                   AI 引擎路由層 (AI Router)                       │
│                                                                  │
│  ┌─────────────────┐           ┌─────────────────────────────┐   │
│  │  NSFW 分級器     │           │      Prompt Builders        │   │
│  │  (RAG Semantic) │           │   (OpenAI/Grok 專用)        │   │
│  │  L1-L5 分級     │  ────────▶│   繼承式架構                │   │
│  └─────────────────┘           └─────────────────────────────┘   │
└─────────────────────────────────┬─────────────────────────────────┘
                                │
        ┌───────────────────────────┴───────────────────────────┐
        │                                                       │
        ▼                                                       ▼
┌───────────────────┐                              ┌───────────────────┐
│   OpenAI GPT-4o   │                              │    Grok AI        │
│   (L1-L3 安全)    │                              │   (L4-L5 開放)   │
│   女性向優化      │                              │   創意表達         │
└───────────────────┘                              └───────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                     資料存儲層 (Data Layer)                       │
├─────────────────────────────────┬───────────────────────────────┤
│         PostgreSQL             │    Keyword Classification     │
│   • users, characters          │   • Built-in keyword rules    │
│   • chats, messages            │   • Zero-cost NSFW detection  │
│   • relationships              │   • Microsecond response      │
│   • admins                     │                               │
└─────────────────────────────────┴───────────────────────────────┘
```

—

## 對應到專案目錄

- `handlers/`：HTTP handlers（`relationship.go`、`search.go` 等已轉為型別化回應，直接回傳資料庫真實欄位）
- `services/`：核心服務邏輯（Chat、Character、Relationships、NSFW 分級器、AI Router、TTS、外部 AI 客戶端）
- `routes/`：路由註冊入口 (`routes.go`)
- `models/`：資料模型與 API 響應定義
  - `models/db/`：資料庫表模型（User、Character、Chat、Relationship、Admin…）
- `cmd/bun/`：資料庫連線、遷移、種子資料
- `services/keyword_classifier_*.go`：NSFW 關鍵字分類器（L1-L5 等級）
- `middleware/`：認證、日誌、CORS 等橫切關注
- `utils/`：日誌、錯誤、JWT、工具函式
- `public/`：靜態頁面與 Swagger UI
- `docs/`：由 `swag` 自動產出的 API 文件

—

## 系統特色

### 架構優化重構
- **API 響應簡化**：關係與搜尋 handlers 回傳嚴謹的 JSON 結構，移除舊有假資料欄位
- **資料真實性**：所有搜尋、關係與情感統計皆直接讀取資料庫欄位或 JSONB (`relationships.emotion_data`)
- **關鍵字分類升級**：NSFW 模組採用關鍵字匹配，零 API 成本微秒級響應
- **AI 引擎路由**：Chat Service 依 NSFW 分級與角色標籤切換 OpenAI / Grok 並驅動專屬 Prompt Builders
- **Prompt 架構**：繼承式 Prompt Builder，依照引擎 (OpenAI/Grok) 客製

### 性能與可維護性
- **內建關鍵字**：NSFW 分類器內建關鍵字規則，零外部依賴
- **記憶體運算**：NSFW 關鍵字純記憶體比對，平均 < 100μs
- **程式碼清理**：移除未使用 Helper，統一時間與分頁處理工具
- **Relationships 單一真實來源**：`relationships` 表同步保存好感度、情緒、親密度與歷史 (`emotion_data`)
- **JSONB 優化**：`emotion_data` 持久化歷史事件，篩選效率佳
- **資料庫一致性**：統一欄位命名，搜尋/關係 handler 使用相同查詢來源
- **維護工具**：`tools/keyword_*.go` 關鍵字分析與衝突檢測工具

## 其他說明

- **API 狀態**：57 個端點（100% 完成），詳見 `API_PROGRESS.md` 與 Swagger
- **配置管理**：完整環境變數說明請參考 `CONFIGURATION.md` 與 `.env.example`
- **測試覆蓋**：24/24 核心測試 + 6 組 Shell 測試腳本
- **持續重構**：Relationship 與 Search handler 皆已改為型別化結構，便於日後擴充
- **快取現況**：Redis 僅保留在 `docker-compose.yml` 作為預備服務，尚未在應用程式中啟用
