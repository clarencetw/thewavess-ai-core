# 🏗️ 系統架構（Architecture）

本文件概述技術棧與高層架構，並對應到專案目錄結構。

—

## 技術棧

- 語言與框架：Go 1.23+、Gin
- 認證：JWT（Access/Refresh）
- 資料庫：PostgreSQL、Bun ORM、SQL 遷移（5個數據表）
- AI：OpenAI（L1–L4）、Grok（L5）、智能路由選擇
- 文件：Swagger/OpenAPI 3（49個API端點）
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
          │           Business Logic                   │
          │ Chat / Character / Emotion / Memory / TTS │
          └─────────────────────┬─────────────────────┘
                                │
    ┌─────────────┬─────────────┴─────────────┬─────────────┐
    │             │                           │             │
┌───▼───┐    ┌───▼───┐                  ┌───▼────┐    ┌───▼───┐
│OpenAI │    │ Grok  │                  │Postgres│    │(Cache)│
│ L1-L4 │    │  L5   │                  │  Bun   │    │ Redis │
└───────┘    └───────┘                  └────────┘    └───────┘
```

—

## 對應到專案目錄

- handlers/：HTTP handlers（11個文件：admin、user、chat、character、relationship、monitor 等）
- services/：核心服務邏輯（11個文件：chat、character、admin、AI客戶端、prompt等）
- routes/：路由註冊入口（routes.go）
- models/：資料模型與響應格式（整合JSONB欄位儲存複雜數據）
  - db/：5個資料表模型（User、Character、Chat、Relationship、Admin）
- cmd/bun/：資料庫連線、遷移工具、種子資料
- middleware/：認證、日誌、CORS 等橫切關注
- utils/：日誌、錯誤、JWT、輔助工具
- public/：靜態頁面與 Swagger UI 入口
- docs/：由 swag 自動產生的 API 文件

—

## 系統特色

### 簡化架構
- **數據整合**：Memory功能整合到 relationships.emotion_data JSONB字段
- **標籤系統**：Tags功能整合到 character.metadata.tags 字段
- **智能路由**：自動選擇最適合的AI引擎處理請求
- **NSFW分級**：5級內容分類系統，自動路由到合適的AI服務

### 性能優化
- **JSONB存儲**：使用PostgreSQL JSONB提高查詢效率
- **智能緩存**：基於內容類型的智能緩存策略
- **異步處理**：支援背景任務和長時間運行的處理

## 其他說明

- **API狀態**：49個端點，詳見 API_PROGRESS.md 與 Swagger
- **配置管理**：完整的環境變數支援，參考 CONFIGURATION.md 與 .env.example
- **測試覆蓋**：完整的單元測試、集成測試、API測試套件

