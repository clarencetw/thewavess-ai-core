# 管理員系統使用指南

## 🏗️ 系統架構

本系統實現了**獨立的管理員系統**，與一般用戶完全分離：

- **用戶系統** (`users` 表): 服務的終端用戶  
- **管理員系統** (`admins` 表): 後台管理人員

## 🔐 管理員帳號

### 預設管理員帳號

系統提供以下預設管理員帳號（通過 fixtures.yml 載入）：

| 用戶名 | 密碼 | 角色 | 權限 |
|--------|------|------|------|
| `admin` | `admin123456` | 超級管理員 | 所有權限 (`*`) |
| `manager` | `manager123456` | 一般管理員 | 基本管理權限 (`basic`) |

### 安全建議

⚠️ **重要**: 生產環境使用前請務必修改預設密碼！

## 🎭 角色與權限

### 管理員角色

1. **super_admin** (超級管理員)
   - 權限: `*` (所有權限)
   - 用途: 系統管理、管理員管理、所有後台功能

2. **admin** (一般管理員)  
   - 權限: `basic` (基本管理權限)
   - 用途: 基本管理功能，不包括管理員管理

### 權限系統

權限系統：
- `*` - 所有權限 (僅超級管理員擁有)
- `basic` - 基本管理權限 (一般管理員)

## 🚀 快速開始

### 1. 初始化管理員系統

```bash
# 全新開始 (推薦)
make fresh-start

# 快速設置 (僅資料庫+fixtures)  
make quick-setup

# 僅重新載入 fixtures (包含管理員)
make fixtures-recreate

# 或手動執行
make migrate-reset
make migrate
make fixtures
```

### 2. 管理員登入

```bash
curl -X POST "http://localhost:8080/api/v1/admin/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }'
```

響應會包含管理員JWT令牌：

```json
{
  "success": true,
  "data": {
    "admin": {...},
    "access_token": "eyJ...",
    "token_type": "Bearer",
    "expires_in": 28800
  }
}
```

### 3. 使用管理員API

```bash
# 獲取管理員個人資料
curl -H "Authorization: Bearer <admin_token>" \
  "http://localhost:8080/api/v1/admin/profile"

# 獲取管理員列表（需要超級管理員權限）
curl -H "Authorization: Bearer <super_admin_token>" \
  "http://localhost:8080/api/v1/admin/admins"
```

## 📡 API 端點（9個管理系統端點）

### 認證相關
- `POST /admin/auth/login` - 管理員登入

### 管理員管理
- `GET /admin/profile` - 獲取個人資料 (所有管理員)
- `GET /admin/admins` - 管理員列表 (僅超級管理員)
- `POST /admin/admins` - 創建管理員 (僅超級管理員)

### 系統管理
- `GET /admin/stats` - 系統統計 (所有管理員)
- `GET /admin/logs` - 系統日誌 (所有管理員)

### 用戶管理  
- `GET /admin/users` - 用戶列表 (所有管理員)
- `PUT /admin/users/:id` - 更新用戶 (所有管理員)

## 🛡️ 安全特性

### JWT 令牌

- **有效期**: 8小時（比用戶令牌短）
- **簽名**: 專用簽名 (`thewavess-ai-core-admin`)  
- **包含**: 管理員ID、角色、權限列表

### 帳號安全

- **密碼加密**: bcrypt 加密存儲
- **失敗鎖定**: 5次失敗後鎖定30分鐘
- **權限檢查**: 每個API都有權限驗證

### 中間件

- `AdminMiddleware()` - 驗證管理員令牌  
- `RequireSuperAdmin()` - 檢查超級管理員權限

## 🧪 開發與測試

### 創建管理員

```bash
# 使用API創建（需要super_admin權限）
curl -X POST "http://localhost:8080/api/v1/admin/admins" \
  -H "Authorization: Bearer <super_admin_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "new_admin",
    "email": "new@example.com", 
    "password": "secure_password",
    "role": "admin"
  }'
```

### 測試

```bash
# 測試管理員登入
curl -X POST "http://localhost:8080/api/v1/admin/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123456"}'

# 測試管理員API
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8080/api/v1/admin/profile"
```

### 資料庫直接查詢

```sql
-- 查看所有管理員
SELECT username, email, role, status, permissions FROM admins;

-- 檢查權限
SELECT username, role, permissions FROM admins WHERE status = 'active';
```

## 📝 開發注意事項

### 路由設計

- 管理員認證路由無需中間件: `/admin/auth/*`
- 管理員API需要管理員中間件: `/admin/*` 
- 管理員管理功能需要 `RequireSuperAdmin()` 中間件

### 權限設計

- 超級管理員 (`super_admin`) 擁有所有權限
- 一般管理員 (`admin`) 擁有基本管理權限
- 只有超級管理員可以管理其他管理員

### 安全考慮

- 管理員令牌不可用於用戶API
- 用戶令牌不可用於管理員API  
- 敏感操作需要記錄日誌
- 生產環境請修改預設密碼

## 🔧 故障排除

### 常見問題

1. **管理員登入失敗**
   - 檢查用戶名/密碼是否正確
   - 檢查帳號是否被鎖定
   - 查看 `failed_attempts` 和 `locked_until` 欄位

2. **權限不足錯誤**
   - 檢查管理員角色和權限
   - 確認API端點需要的權限
   - 超級管理員可訪問所有端點

3. **令牌無效**
   - 檢查令牌是否過期（8小時）
   - 確認使用的是管理員令牌，不是用戶令牌
   - 檢查 `Authorization: Bearer <token>` 格式

### 日誌檢查

```bash
# 檢查管理員操作日誌
grep "管理員" server.log

# 檢查認證失敗
grep "登入失敗" server.log
```

## 🚀 未來擴展

- [ ] 管理員操作日誌審計
- [ ] 更細緻的權限控制（資源級別）  
- [ ] 管理員密碼策略設定
- [ ] 雙因子認證（2FA）
- [ ] 管理員會話管理