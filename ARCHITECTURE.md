# 🏗️ 系統架構（Architecture）

本文件概述技術棧與高層架構，並對應到專案目錄結構。

—

## 技術棧

- 語言與框架：Go 1.23+、Gin
- 認證：JWT（Access/Refresh）
- 資料庫：PostgreSQL、Bun ORM、SQL 遷移
- AI：OpenAI（L1–L4）、Grok（L5）
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
          │           Business Logic                   │
          │   Chat / Character / Memory / TTS         │
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

- handlers/：HTTP handlers（auth、user、chat、character、monitor 等）
- services/：核心服務邏輯（chat、nsfw、memory、tts、openai/grok 客戶端）
- routes/：路由註冊入口（routes.go）
- models/：資料模型與響應格式
- database/：連線、遷移、種子資料（cmd/bun）
- middleware/：認證、日誌、CORS 等橫切關注
- utils/：日誌、錯誤、JWT、輔助工具
- public/：靜態頁面與 Swagger UI 入口
- docs/：由 swag 自動產生的 API 文件

—

## 其他說明

- 端點清單與可用性以 API_PROGRESS.md 與 Swagger 為準。
- 設定與參數請參考 CONFIGURATION.md 與 .env.example。

