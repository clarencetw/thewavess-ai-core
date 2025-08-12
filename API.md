# Thewavess AI Core - API 文檔

## 🚀 快速開始

### 環境要求
- Go 1.22+
- Docker & Docker Compose (推薦)
- Make (可選，但推薦)

### 快速啟動

#### 使用 Docker Compose (推薦)
```bash
# 1. Clone 專案
git clone https://github.com/clarencetw/thewavess-ai-core.git
cd thewavess-ai-core

# 2. 複製環境變數檔案
cp .env.example .env

# 3. 編輯 .env 檔案，填入你的 API Keys
nano .env

# 4. 啟動所有服務
docker-compose up -d

# 5. 檢查服務狀態
docker-compose ps
```

#### 本地開發模式
```bash
# 1. 安裝依賴
make install

# 2. 生成 API 文檔並啟動服務
make dev

# 或者分步驟執行
make docs    # 生成 Swagger 文檔
make run     # 啟動服務器
```

### API 文檔訪問

啟動後，你可以透過以下方式查看 API 文檔：

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **健康檢查**: http://localhost:8080/health
- **系統狀態**: http://localhost:8080/api/v1/status

### 🛠️ 開發指令

```bash
# 查看所有可用指令
make help

# 常用開發指令
make install      # 安裝依賴
make docs         # 生成 API 文檔
make run          # 啟動服務器
make test         # 運行測試
make build        # 編譯應用
make clean        # 清理構建檔案

# Docker 相關
make docker-build # 建立 Docker 映像
make docker-run   # 運行 Docker 容器
```

---

## 📚 API 規格文檔

### 基本信息
- **Base URL**: `https://api.thewavess.ai/api/v1` (生產環境)
- **本地開發**: `http://localhost:8080/api/v1`
- **認證方式**: JWT Bearer Token
- **內容類型**: `application/json`
- **API 版本**: v1

### 認證
所有 API 請求都需要在 Header 中包含 JWT Token：
```
Authorization: Bearer <your_jwt_token>
```

### 快速測試範例

#### 1. 用戶註冊
```bash
curl -X POST http://localhost:8080/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice123",
    "email": "alice@example.com",
    "password": "password123",
    "birth_date": "2000-01-01",
    "gender": "female",
    "nickname": "小愛"
  }'
```

#### 2. 用戶登入
```bash
curl -X POST http://localhost:8080/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice123",
    "password": "password123"
  }'
```

#### 3. 使用 JWT Token
```bash
# 將從登入回應中獲得的 access_token 用於後續請求
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN_HERE"
```

### 基本對話流程

#### 1. 選擇角色
```bash
curl -X PUT http://localhost:8080/api/v1/user/character \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "character_id": "char_001"
  }'
```

#### 2. 創建對話會話
```bash
curl -X POST http://localhost:8080/api/v1/chat/session \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "character_id": "char_001",
    "title": "與陸寒淵的對話",
    "mode": "normal"
  }'
```

#### 3. 發送訊息
```bash
curl -X POST http://localhost:8080/api/v1/chat/message \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "YOUR_SESSION_ID",
    "message": "你好"
  }'
```

## API 端點總覽

### 1. 用戶管理 (User Management)

#### 1.1 用戶註冊
```
POST /user/register
```

**請求體**:
```json
{
  "username": "string",
  "email": "string", 
  "password": "string",
  "birth_date": "2000-01-01",
  "gender": "female|male|other",
  "nickname": "string"
}
```

**回應**:
```json
{
  "success": true,
  "message": "用戶註冊成功",
  "data": {
    "user_id": "uuid",
    "access_token": "jwt_token",
    "refresh_token": "jwt_token",
    "expires_in": 3600
  }
}
```

#### 1.2 用戶登入
```
POST /user/login
```

**請求體**:
```json
{
  "email": "string",
  "password": "string"
}
```

**回應**:
```json
{
  "success": true,
  "data": {
    "user_id": "uuid",
    "access_token": "jwt_token", 
    "refresh_token": "jwt_token",
    "expires_in": 3600,
    "user": {
      "id": "uuid",
      "username": "string",
      "email": "string",
      "nickname": "string",
      "avatar_url": "string"
    }
  }
}
```

#### 1.3 用戶登出
```
POST /user/logout
```

**回應**:
```json
{
  "success": true,
  "message": "登出成功"
}
```

#### 1.4 刷新 Token
```
POST /user/refresh
```

**請求體**:
```json
{
  "refresh_token": "string"
}
```

#### 1.5 獲取個人資料
```
GET /user/profile
```

**回應**:
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "username": "string",
    "email": "string",
    "nickname": "string",
    "birth_date": "2000-01-01",
    "gender": "female",
    "avatar_url": "string",
    "created_at": "2024-01-01T00:00:00Z",
    "preferences": {
      "default_character": "character_id",
      "nsfw_enabled": true,
      "voice_enabled": true,
      "notification_enabled": true
    }
  }
}
```

#### 1.6 更新個人資料
```
PUT /user/profile
```

**請求體**:
```json
{
  "nickname": "string",
  "birth_date": "2000-01-01",
  "gender": "female|male|other",
  "avatar_url": "string"
}
```

#### 1.7 更新偏好設定
```
PUT /user/preferences
```

**請求體**:
```json
{
  "default_character": "character_id",
  "nsfw_enabled": true,
  "voice_enabled": true,
  "notification_enabled": true,
  "preferred_voice": "voice_id"
}
```

#### 1.8 上傳頭像
```
POST /user/avatar
```

**請求**: multipart/form-data
- `file`: 圖片檔案 (max 5MB, jpg/png)

**回應**:
```json
{
  "success": true,
  "data": {
    "avatar_url": "https://cdn.thewavess.ai/avatars/uuid.jpg"
  }
}
```

#### 1.9 刪除帳號
```
DELETE /user/account
```

**請求體**:
```json
{
  "password": "string",
  "confirmation": "DELETE_MY_ACCOUNT"
}
```

### 2. 角色管理 (Character Management)

#### 2.1 獲取角色列表
```
GET /character/list
```

**查詢參數**:
- `page`: 頁碼 (default: 1)
- `limit`: 每頁數量 (default: 10)
- `type`: 角色類型過濾

**回應**:
```json
{
  "success": true,
  "data": {
    "characters": [
      {
        "id": "char_001",
        "name": "陸寒淵",
        "type": "dominant",
        "description": "霸道總裁，冷峻外表下隱藏深情",
        "avatar_url": "string",
        "voice_id": "voice_001",
        "popularity": 95,
        "tags": ["霸道總裁", "深情", "禁慾系"],
        "appearance": {
          "height": "185cm",
          "hair_color": "黑髮",
          "eye_color": "深邃黑眸",
          "description": "俊朗五官，總是穿著剪裁合身的西裝"
        },
        "personality": {
          "traits": ["冷酷", "強勢", "專一", "佔有欲"],
          "likes": ["工作", "掌控", "用戶"],
          "dislikes": ["被違抗", "失去控制"]
        }
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 3,
      "total_count": 25,
      "has_next": true
    }
  }
}
```

#### 2.2 獲取角色詳情
```
GET /character/{character_id}
```

#### 2.3 獲取當前選擇角色
```
GET /user/character
```

#### 2.4 選擇當前角色
```
PUT /user/character
```

**請求體**:
```json
{
  "character_id": "char_001"
}
```

#### 2.5 獲取角色統計數據
```
GET /character/{character_id}/stats
```

**回應**:
```json
{
  "success": true,
  "data": {
    "total_conversations": 1523,
    "average_rating": 4.8,
    "total_users": 892,
    "popular_tags": ["溫柔", "體貼", "浪漫"]
  }
}
```

### 3. 對話管理 (Chat Management)

#### 3.1 創建新會話
```
POST /chat/session
```

**請求體**:
```json
{
  "character_id": "char_001",
  "mode": "normal|novel|nsfw",
  "title": "string",
  "tags": ["溫柔", "日常"]
}
```

**回應**:
```json
{
  "success": true,
  "data": {
    "session_id": "session_uuid",
    "character_id": "char_001",
    "mode": "normal",
    "title": "與陸寒淵的對話",
    "created_at": "2024-01-01T00:00:00Z",
    "last_message_at": null,
    "message_count": 0,
    "emotional_state": {
      "affection": 50,
      "mood": "neutral",
      "relationship": "stranger"
    }
  }
}
```

#### 3.2 獲取會話資訊
```
GET /chat/session/{session_id}
```

#### 3.3 獲取用戶會話列表
```
GET /chat/sessions
```

**查詢參數**:
- `page`: 頁碼
- `limit`: 每頁數量
- `character_id`: 角色過濾
- `mode`: 模式過濾

#### 3.4 發送訊息
```
POST /chat/message
```

**請求體**:
```json
{
  "session_id": "session_uuid",
  "message": "嗨，你好！",
  "message_type": "text|image|voice",
  "metadata": {
    "image_url": "string",
    "voice_duration": 5.2
  }
}
```

**回應**:
```json
{
  "success": true,
  "data": {
    "message_id": "msg_uuid",
    "session_id": "session_uuid",
    "character_response": {
      "message": "你好，很高興見到你。",
      "emotion": "happy",
      "affection_change": 2,
      "engine_used": "openai",
      "response_time_ms": 1250,
      "tts_url": "https://cdn.thewavess.ai/tts/msg_uuid.mp3"
    },
    "emotional_state": {
      "affection": 52,
      "mood": "happy", 
      "relationship": "stranger"
    },
    "novel_choices": [],
    "special_event": null
  }
}
```

#### 3.5 重新生成回應
```
POST /chat/regenerate
```

**請求體**:
```json
{
  "message_id": "msg_uuid",
  "regeneration_reason": "tone|content|length"
}
```

#### 3.6 切換對話模式
```
PUT /chat/session/{session_id}/mode
```

**請求體**:
```json
{
  "mode": "normal|novel|nsfw",
  "transition_message": "我們來玩個遊戲吧..."
}
```

#### 3.7 獲取會話對話歷史
```
GET /chat/session/{session_id}/history
```

**查詢參數**:
- `page`: 頁碼
- `limit`: 每頁數量 (max 50)
- `before`: 訊息 ID，獲取該訊息之前的歷史
- `after`: 訊息 ID，獲取該訊息之後的歷史

#### 3.8 為會話添加標籤
```
POST /chat/session/{session_id}/tag
```

**請求體**:
```json
{
  "tags": ["浪漫", "甜蜜", "日常"]
}
```

#### 3.9 結束對話會話
```
DELETE /chat/session/{session_id}
```

#### 3.10 匯出對話記錄
```
GET /chat/session/{session_id}/export
```

**查詢參數**:
- `format`: `json|txt|pdf`

#### 3.11 搜尋對話內容
```
GET /chat/search
```

**查詢參數**:
- `q`: 搜尋關鍵字
- `character_id`: 角色過濾
- `date_from`: 開始日期
- `date_to`: 結束日期

### 4. 小說模式 (Novel Mode)

#### 4.1 開始小說模式
```
POST /novel/start
```

**請求體**:
```json
{
  "session_id": "session_uuid",
  "scenario": "office|school|historical|modern",
  "character_role": "boss|classmate|emperor|ceo",
  "user_role": "employee|student|concubine|secretary",
  "tags": ["甜寵", "霸道總裁"]
}
```

**回應**:
```json
{
  "success": true,
  "data": {
    "novel_id": "novel_uuid",
    "session_id": "session_uuid",
    "opening_scene": "辦公室的燈光依然亮著...",
    "character_introduction": "陸寒淵正專注地看著文件...",
    "initial_choices": [
      {
        "id": "choice_001",
        "text": "敲門進入辦公室",
        "consequence": "主動接觸，可能增加好感"
      },
      {
        "id": "choice_002", 
        "text": "在門外等待",
        "consequence": "保持距離，展現禮貌"
      }
    ]
  }
}
```

#### 4.2 選擇劇情分支
```
POST /novel/choice
```

**請求體**:
```json
{
  "novel_id": "novel_uuid",
  "choice_id": "choice_001",
  "user_action": "我輕敲辦公室的門"
}
```

#### 4.3 保存進度
```
POST /novel/progress/save
```

**請求體**:
```json
{
  "novel_id": "novel_uuid",
  "save_name": "辦公室邂逅 - 第一章",
  "description": "剛剛敲門進入總裁辦公室"
}
```

#### 4.4 載入進度
```
GET /novel/progress/{progress_id}
```

#### 4.5 獲取存檔列表
```
GET /novel/progress/list
```

#### 4.6 獲取小說統計
```
GET /novel/{novel_id}/stats
```

#### 4.7 刪除存檔
```
DELETE /novel/progress/{progress_id}
```

### 5. 情感系統 (Emotion System)

#### 5.1 獲取情感狀態
```
GET /emotion/status
```

**查詢參數**:
- `session_id`: 特定會話
- `character_id`: 特定角色

**回應**:
```json
{
  "success": true,
  "data": {
    "session_id": "session_uuid",
    "character_id": "char_001",
    "affection": 75,
    "mood": "happy",
    "relationship": "lover",
    "trust_level": 68,
    "intimacy_level": 45,
    "last_interaction": "2024-01-01T12:00:00Z",
    "milestone_progress": {
      "next_milestone": "深度交流",
      "progress_percentage": 80,
      "required_affection": 80
    }
  }
}
```

#### 5.2 獲取好感度歷史
```
GET /emotion/affection/history
```

#### 5.3 觸發特殊事件
```
POST /emotion/event
```

**請求體**:
```json
{
  "session_id": "session_uuid",
  "event_type": "anniversary|birthday|valentine|special_date",
  "event_data": {
    "date": "2024-02-14",
    "message": "今天是情人節..."
  }
}
```

#### 5.4 獲取關係里程碑
```
GET /emotion/milestones
```

#### 5.5 重置情感狀態
```
POST /emotion/reset
```

### 6. TTS 語音功能 (Text-to-Speech)

#### 6.1 生成語音
```
POST /tts/generate
```

**請求體**:
```json
{
  "text": "你好，很高興見到你",
  "voice_id": "voice_001",
  "speed": 1.0,
  "emotion": "happy|sad|angry|neutral",
  "format": "mp3|wav|ogg"
}
```

**回應**:
```json
{
  "success": true,
  "data": {
    "audio_url": "https://cdn.thewavess.ai/tts/audio_uuid.mp3",
    "duration": 3.5,
    "file_size": 125840,
    "expires_at": "2024-01-02T00:00:00Z"
  }
}
```

#### 6.2 獲取語音列表
```
GET /tts/voices
```

**回應**:
```json
{
  "success": true,
  "data": {
    "voices": [
      {
        "voice_id": "voice_001",
        "name": "磁性低音",
        "description": "成熟男性聲音，適合霸道總裁",
        "character_ids": ["char_001"],
        "language": "zh-CN",
        "gender": "male",
        "preview_url": "https://cdn.thewavess.ai/previews/voice_001.mp3"
      }
    ]
  }
}
```

#### 6.3 預覽語音
```
POST /tts/preview
```

**請求體**:
```json
{
  "voice_id": "voice_001",
  "preview_text": "這是語音預覽"
}
```

#### 6.4 獲取語音配置
```
GET /tts/config
```

#### 6.5 批量生成語音
```
POST /tts/batch
```

### 7. 記憶系統 (Memory System)

#### 7.1 獲取用戶記憶
```
GET /memory/user/{user_id}
```

**查詢參數**:
- `type`: `short_term|long_term|emotional`
- `character_id`: 特定角色記憶
- `limit`: 返回數量

**回應**:
```json
{
  "success": true,
  "data": {
    "user_id": "user_uuid",
    "memories": [
      {
        "id": "memory_uuid",
        "type": "long_term",
        "content": "用戶喜歡吃提拉米蘇",
        "importance": 8,
        "created_at": "2024-01-01T00:00:00Z",
        "last_accessed": "2024-01-05T12:00:00Z",
        "access_count": 5,
        "character_id": "char_001",
        "session_id": "session_uuid",
        "tags": ["喜好", "食物"]
      }
    ],
    "memory_stats": {
      "total_memories": 156,
      "short_term": 12,
      "long_term": 134,
      "emotional": 10
    }
  }
}
```

#### 7.2 手動保存記憶
```
POST /memory/save
```

**請求體**:
```json
{
  "user_id": "user_uuid",
  "character_id": "char_001",
  "content": "用戶今天心情不好，需要安慰",
  "type": "emotional",
  "importance": 7,
  "tags": ["情緒", "當日狀態"]
}
```

#### 7.3 選擇性遺忘
```
DELETE /memory/forget
```

**請求體**:
```json
{
  "memory_ids": ["memory_uuid1", "memory_uuid2"],
  "reason": "用戶要求刪除"
}
```

#### 7.4 記憶時間線
```
GET /memory/timeline
```

**查詢參數**:
- `date_from`: 開始日期
- `date_to`: 結束日期
- `character_id`: 角色過濾

#### 7.5 搜尋記憶
```
POST /memory/search
```

**請求體**:
```json
{
  "query": "提拉米蘇",
  "user_id": "user_uuid",
  "character_id": "char_001",
  "search_type": "keyword|semantic",
  "limit": 10
}
```

#### 7.6 記憶統計
```
GET /memory/stats
```

#### 7.7 記憶備份
```
POST /memory/backup
```

#### 7.8 記憶還原
```
POST /memory/restore
```

### 8. 標籤系統 (Tag System)

#### 8.1 獲取所有可用標籤
```
GET /tags
```

**查詢參數**:
- `category`: `content|nsfw|scene|emotion`
- `language`: `zh-CN|en`

**回應**:
```json
{
  "success": true,
  "data": {
    "tags": [
      {
        "id": "tag_001",
        "name": "溫柔",
        "category": "content",
        "description": "溫和體貼的互動風格",
        "usage_count": 1523,
        "nsfw": false,
        "related_tags": ["體貼", "細心", "關懷"]
      }
    ],
    "categories": {
      "content": ["溫柔", "霸道", "禁慾"],
      "scene": ["辦公室", "校園", "古風"],
      "nsfw": ["親密", "激情", "深度互動"],
      "emotion": ["開心", "害羞", "興奮"]
    }
  }
}
```

#### 8.2 獲取熱門標籤
```
GET /tags/popular
```

#### 8.3 創建自定義標籤
```
POST /tags/custom
```

#### 8.4 獲取標籤建議
```
GET /tags/suggestions
```

### 9. 檔案上傳 (File Upload)

#### 9.1 上傳圖片
```
POST /upload/image
```

**請求**: multipart/form-data
- `file`: 圖片檔案 (max 10MB)
- `type`: `avatar|chat|background`

#### 9.2 上傳語音
```
POST /upload/voice
```

#### 9.3 獲取上傳記錄
```
GET /upload/history
```

### 10. 通知系統 (Notification System)

#### 10.1 獲取通知列表
```
GET /notifications
```

#### 10.2 標記通知已讀
```
PUT /notifications/{notification_id}/read
```

#### 10.3 獲取通知設定
```
GET /notifications/settings
```

#### 10.4 更新通知設定
```
PUT /notifications/settings
```

### 11. 統計分析 (Analytics)

#### 11.1 獲取用戶統計
```
GET /analytics/user
```

#### 11.2 獲取對話統計
```
GET /analytics/conversations
```

#### 11.3 獲取角色人氣統計
```
GET /analytics/characters
```

### 12. 系統管理 (System Management)

#### 12.1 健康檢查
```
GET /health
```

**回應**:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T00:00:00Z",
    "version": "1.0.0",
    "services": {
      "database": "healthy",
      "redis": "healthy", 
      "vector_db": "healthy",
      "openai_api": "healthy",
      "grok_api": "healthy"
    },
    "uptime": "72h30m15s"
  }
}
```

#### 12.2 API 版本
```
GET /version
```

#### 12.3 系統狀態
```
GET /status
```

## 錯誤處理

### 標準錯誤格式
```json
{
  "success": false,
  "error": {
    "code": "AUTH_TOKEN_EXPIRED",
    "message": "認證令牌已過期",
    "details": "JWT token expired at 2024-01-01T00:00:00Z",
    "timestamp": "2024-01-01T01:00:00Z",
    "request_id": "req_uuid"
  }
}
```

### 常見錯誤代碼

| 錯誤代碼 | HTTP 狀態碼 | 描述 |
|---------|------------|------|
| `INVALID_TOKEN` | 401 | 無效的認證令牌 |
| `TOKEN_EXPIRED` | 401 | 認證令牌已過期 |
| `INSUFFICIENT_PERMISSIONS` | 403 | 權限不足 |
| `RESOURCE_NOT_FOUND` | 404 | 資源不存在 |
| `VALIDATION_ERROR` | 400 | 請求參數驗證失敗 |
| `RATE_LIMIT_EXCEEDED` | 429 | 請求頻率超限 |
| `AI_SERVICE_UNAVAILABLE` | 503 | AI 服務不可用 |
| `NSFW_CONTENT_BLOCKED` | 451 | NSFW 內容被阻擋 |

## 請求頻率限制

| 端點類型 | 限制 |
|---------|------|
| 一般 API | 100 請求/分鐘 |
| 對話 API | 30 請求/分鐘 |
| TTS API | 20 請求/分鐘 |
| 檔案上傳 | 10 請求/分鐘 |

## WebSocket 支援

### 即時對話連接
```
ws://api.thewavess.ai/ws/chat/{session_id}?token={jwt_token}
```

### 訊息格式
```json
{
  "type": "message|typing|emotion_update|system",
  "data": {
    "content": "string",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```