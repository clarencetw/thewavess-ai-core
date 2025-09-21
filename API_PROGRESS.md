# API 開發進度

## 📊 總體狀況
- **總計端點**: 57 個
- **已完成**: 57 個 (100%)
- **測試狀態**: 24/24 測試通過 (100%)
- **管理介面**: 完整實現，可進行角色管理、系統監控與帳號維護

| 類別 | 端點數 | 主要權限 | 備註 |
|------|--------|----------|------|
| 系統管理 | 2 | ⚪ 公開 | 版本與狀態查詢 |
| 系統監控 | 5 | ⚪ 公開 | 健康 / Ready / Live / Stats / Metrics |
| 認證系統 | 4 | ⚪ 公開 / 🟡 用戶 | 登入、註冊、Refresh、登出 |
| 用戶系統 | 4 | 🟡 用戶 | 個人資料、頭像、刪除帳號 |
| 對話系統 | 9 | 🟡 用戶 | 會話 CRUD、訊息、再生與匯出 |
| 角色系統 | 8 | ⚪ 公開 / 🟡 用戶 | 列表、搜尋、CRUD、Profile |
| 情感系統 | 3 | 🟡 用戶 | 關係狀態 / 好感度 / 歷史 |
| 搜尋系統 | 2 | 🟡 用戶 | 對話搜尋、全域搜尋 |
| TTS 系統 | 2 | ⚪ 公開 / 🟡 用戶 | 語音生成、語音列表 |
| 管理系統 | 18 | 🔴 管理員 / 🟣 超管 | 用戶、角色、聊天、管理員管理 |

## 🎯 權限標誌
- ⚪ **公開**: 無需認證
- 🟡 **用戶**: 需要用戶 JWT Token
- 🔴 **管理員**: 需要管理員 JWT Token
- 🟣 **超級管理員**: 需要超級管理員權限

## ✅ 已實現端點

### 系統管理 (2)
- `GET /api/v1/version` - API版本 ⚪
- `GET /api/v1/status` - 系統狀態 ⚪

### 系統監控 (5)
- `GET /api/v1/monitor/health` - 健康檢查 ⚪
- `GET /api/v1/monitor/ready` - 就緒檢查 ⚪
- `GET /api/v1/monitor/live` - 存活檢查 ⚪
- `GET /api/v1/monitor/stats` - 系統統計 ⚪
- `GET /api/v1/monitor/metrics` - Prometheus 指標 ⚪

### 認證系統 (4)
- `POST /api/v1/auth/register` - 用戶註冊 ⚪
- `POST /api/v1/auth/login` - 用戶登入 ⚪
- `POST /api/v1/auth/refresh` - 刷新 Token ⚪
- `POST /api/v1/auth/logout` - 用戶登出 🟡

### 用戶系統 (4)
- `GET /api/v1/user/profile` - 個人資料 🟡
- `PUT /api/v1/user/profile` - 更新資料 🟡
- `POST /api/v1/user/avatar` - 上傳頭像 🟡
- `DELETE /api/v1/user/account` - 刪除帳號 🟡

### 對話系統 (9)
- `POST /api/v1/chats` - 創建會話 🟡
- `GET /api/v1/chats/{chat_id}` - 會話詳情 🟡
- `GET /api/v1/chats` - 會話列表 🟡
- `POST /api/v1/chats/{chat_id}/messages` - 發送訊息 🟡
- `GET /api/v1/chats/{chat_id}/history` - 對話歷史 🟡
- `PUT /api/v1/chats/{chat_id}/mode` - 更新會話模式 🟡
- `DELETE /api/v1/chats/{chat_id}` - 刪除會話 🟡
- `GET /api/v1/chats/{chat_id}/export` - 匯出對話 🟡
- `POST /api/v1/chats/{chat_id}/messages/{message_id}/regenerate` - 重新生成 🟡

### 角色系統 (8)
- `GET /api/v1/character/list` - 角色列表 ⚪
- `GET /api/v1/character/search` - 角色搜尋 ⚪
- `GET /api/v1/character/{id}` - 角色詳情 ⚪
- `GET /api/v1/character/{id}/stats` - 角色統計 ⚪
- `GET /api/v1/character/{id}/profile` - 角色檔案 🟡
- `POST /api/v1/character` - 創建角色 🟡
- `PUT /api/v1/character/{id}` - 更新角色 🟡
- `DELETE /api/v1/character/{id}` - 刪除角色 🟡

### 情感系統 (3)
- `GET /api/v1/relationships/chat/{chat_id}/status` - 關係狀態 🟡
- `GET /api/v1/relationships/chat/{chat_id}/affection` - 好感度查詢 🟡
- `GET /api/v1/relationships/chat/{chat_id}/history` - 關係歷史 🟡

### 搜尋功能 (2)
- `GET /api/v1/search/chats` - 搜尋對話 🟡
- `GET /api/v1/search/global` - 全局搜尋 🟡  
  - 回傳型別化資料：聊天結果包含 `chat_id`, `dialogue`, `character`, `nsfw_level`, `relevance`；分面資訊提供角色與 NSFW 等級統計

### TTS 語音系統 (2)
- `POST /api/v1/tts/generate` - 生成語音 🟡
- `GET /api/v1/tts/voices` - 語音列表 ⚪

### 管理系統 (18)
- `POST /api/v1/admin/auth/login` - 管理員登入 ⚪
- `GET /api/v1/admin/stats` - 系統統計 🔴
- `GET /api/v1/admin/logs` - 系統日誌 🔴
- `GET /api/v1/admin/users` - 用戶列表 🔴
- `GET /api/v1/admin/users/{id}` - 特定用戶 🔴
- `PUT /api/v1/admin/users/{id}` - 修改用戶 🔴
- `PUT /api/v1/admin/users/{id}/password` - 重置密碼 🔴
- `PUT /api/v1/admin/users/{id}/status` - 更新用戶狀態 🔴
- `GET /api/v1/admin/chats` - 搜尋聊天記錄 🔴
- `GET /api/v1/admin/chats/{chat_id}/history` - 查看聊天歷史 🔴
- `GET /api/v1/admin/characters` - 角色列表 🔴
- `GET /api/v1/admin/characters/{id}` - 角色詳情 🔴
- `PUT /api/v1/admin/characters/{id}` - 更新角色 🔴
- `POST /api/v1/admin/characters/{id}/restore` - 還原角色 🔴
- `DELETE /api/v1/admin/characters/{id}/permanent` - 永久刪除角色 🔴
- `PUT /api/v1/admin/character/{id}/status` - 調整角色狀態 🔴
- `GET /api/v1/admin/admins` - 管理員列表 🟣
- `POST /api/v1/admin/admins` - 創建管理員 🟣

## 🚀 系統特色

### AI 引擎
- **OpenAI GPT-4o**: Level 1-4 內容
- **Grok**: Level 5 極度內容
- **智能路由**: 自動選擇合適引擎並具備 fallback 機制

### NSFW 分級
- **5級分類系統**: 準確率 95%+
- **女性向優化**: 優雅表達，重視氛圍
- **年齡驗證**: 成人內容需 18+ 驗證

### 情感系統
- **AI 驅動**: 智能分析好感度與情緒
- **關係追蹤**: 0-100 好感度系統 + JSONB 歷史記錄
- **歷史回溯**: 支援查詢情感事件與統計

## 📖 文檔
- **即時 API 文檔**: http://localhost:8080/swagger/index.html
- **管理後台**: http://localhost:8080/admin/
- **測試工具**: `./tests/test-all.sh`
