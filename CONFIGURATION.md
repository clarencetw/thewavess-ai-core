# ⚙️ 環境設定

權威來源：[.env.example](./.env.example)

## 最小設定（必填）

複製 `.env.example` 為 `.env`，至少設定：

```env
OPENAI_API_KEY=sk-your-openai-api-key-here
```

## 完整配置

### 資料庫（PostgreSQL）
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=thewavess
DB_PASSWORD=password
DB_NAME=thewavess_ai_core
```

### AI 服務
```env
# OpenAI (必填) - 處理 Level 1-4
OPENAI_API_KEY=sk-your-key
OPENAI_MODEL=gpt-4o
OPENAI_MAX_TOKENS=1200

# Grok (可選) - 處理 Level 5
GROK_API_KEY=your-key
GROK_MODEL=grok-3
GROK_MAX_TOKENS=2000

# TTS (可選) - 語音合成
TTS_API_KEY=sk-your-key  # 未設定時使用 OPENAI_API_KEY
```

### 伺服器設定
```env
PORT=8080
GIN_MODE=debug          # debug/release
ENVIRONMENT=development # development/production
LOG_LEVEL=debug        # debug/info/warn/error
```

### 安全設定
```env
JWT_SECRET=your-super-secret-key
CORS_ALLOWED_ORIGINS=*  # 生產環境應限制
```

### NSFW 分級閾值
```env
NSFW_DETECTION_THRESHOLD=0.5
NSFW_L2_THRESHOLD=2
NSFW_L3_THRESHOLD=2
NSFW_L4_THRESHOLD=2
NSFW_L5_THRESHOLD=1
```

## 系統特色

- **智能路由**: 自動選擇 OpenAI/Grok 引擎
- **5級NSFW分類**: 女性向優化，準確率95%+
- **精簡架構**: 5張表，JSONB整合複雜數據
- **JWT認證**: Access + Refresh Token 機制