# ⚙️ 環境設定（Configuration）

本文件僅提供摘要與最小設定示例，「權威來源」為專案根目錄的 [.env.example](./.env.example)。任何變更與預設值請以該檔案為準。

—

## 最小設定（必填）

將 `.env.example` 複製為 `.env`，至少設定以下變數即可啟動：

```env
# 必填：OpenAI 金鑰（處理 L1–L4）
OPENAI_API_KEY=sk-your-openai-api-key-here

# 建議：本機埠與模式（也可維持預設）
PORT=8080
GIN_MODE=debug
ENVIRONMENT=development
LOG_LEVEL=debug
```

—

## 資料庫（PostgreSQL）

- DB_HOST、DB_PORT、DB_USER、DB_PASSWORD、DB_NAME、DB_SSLMODE

—

## AI 金鑰（以 .env.example 為準）

- OPENAI_API_KEY（必填）：處理 NSFW Level 1–4
- GROK_API_KEY（可選）：處理 NSFW Level 5
- TTS_API_KEY（可選）：未設定時預設使用 OPENAI_API_KEY

—

## 伺服器與日誌（以 .env.example 為準）

- PORT（預設 8080）
- GIN_MODE（debug/release）
- ENVIRONMENT（development/production）
- LOG_LEVEL（debug/info/warn/error）

—

## CORS（跨域，詳見 .env.example）

- CORS_ALLOWED_ORIGINS（預設 *）
- CORS_ALLOWED_METHODS、CORS_ALLOWED_HEADERS、CORS_EXPOSED_HEADERS

—

## NSFW 設定（詳見 .env.example）

- NSFW_DETECTION_THRESHOLD 以及各級觸發門檻（參考 .env.example）

—

## OpenAI 與 Grok 模型（詳見 .env.example）

- OPENAI_API_URL、OPENAI_MODEL、OPENAI_MAX_TOKENS、OPENAI_TEMPERATURE
- GROK_API_URL、GROK_MODEL、GROK_MAX_TOKENS、GROK_TEMPERATURE

—

## 場景描述（詳見 .env.example）

- SCENE_DESCRIPTIONS_ENABLED、SCENE_MAX_LENGTH、SCENE_UPDATE_FREQUENCY

—

## JWT（以 .env.example 為準）

- JWT_SECRET（供未來使用）
