# 管理員系統指南

> 📋 **相關文檔**: 完整文檔索引請參考 [DOCS_INDEX.md](./DOCS_INDEX.md)

## 1. 架構與角色
| 項目 | 說明 |
|------|------|
| 資料表 | `admins`（獨立於使用者 `users`）|
| 認證 | JWT（專用簽名與有效期，與一般用戶分離）|
| Middleware | `AdminMiddleware()` 驗證管理員、`RequireSuperAdmin()` 限制超管專屬端點 |
| Sticky 權限 | 依角色決定可存取的管理 API |

### 角色與預設帳號
| 角色 | 權限代號 | 可執行項目 | 預設帳號 (fixtures) |
|------|-----------|------------|----------------------|
| `super_admin` | `*` | 全部管理 API、管理員管理 | `admin / admin123456` |
| `admin` | `basic` | 系統、用戶、角色、聊天管理 | `manager / manager123456` |

> 生產環境啟動前務必修改預設密碼或直接刪除預設帳號。

## 2. 快速上手
| 步驟 | 指令 | 說明 |
|------|------|------|
| 初始化 | `make fresh-start` | 清理、遷移、載入 fixtures（含管理員）|
| 快速重建 | `make fixtures-recreate` | 在現有資料庫重新載入預設管理員 |
| 登入 | `POST /api/v1/admin/auth/login` | 取得 `access_token`、`token_type`、`expires_in` |
| 呼叫 API | `curl -H "Authorization: Bearer <token>" ...` | 使用管理員 Token 存取後台端點 |

登入範例：
```bash
curl -sS -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123456"}'
```

## 3. API 清單與權限
| 類別 | Method | Path | 權限 |
|------|--------|------|------|
| 認證 | POST | `/api/v1/admin/auth/login` | ⚪ 公開 |
| 系統 | GET | `/api/v1/admin/stats` | 🔴 管理員 |
| 系統 | GET | `/api/v1/admin/logs` | 🔴 管理員 |
| 用戶 | GET | `/api/v1/admin/users` | 🔴 管理員 |
| 用戶 | GET | `/api/v1/admin/users/{id}` | 🔴 管理員 |
| 用戶 | PUT | `/api/v1/admin/users/{id}` | 🔴 管理員 |
| 用戶 | PUT | `/api/v1/admin/users/{id}/password` | 🔴 管理員 |
| 用戶 | PUT | `/api/v1/admin/users/{id}/status` | 🔴 管理員 |
| 聊天 | GET | `/api/v1/admin/chats` | 🔴 管理員 |
| 聊天 | GET | `/api/v1/admin/chats/{chat_id}/history` | 🔴 管理員 |
| 角色 | GET | `/api/v1/admin/characters` | 🔴 管理員 |
| 角色 | GET | `/api/v1/admin/characters/{id}` | 🔴 管理員 |
| 角色 | PUT | `/api/v1/admin/characters/{id}` | 🔴 管理員 |
| 角色 | POST | `/api/v1/admin/characters/{id}/restore` | 🔴 管理員 |
| 角色 | DELETE | `/api/v1/admin/characters/{id}/permanent` | 🔴 管理員 |
| 角色 | PUT | `/api/v1/admin/character/{id}/status` | 🔴 管理員 |
| 管理員 | GET | `/api/v1/admin/admins` | 🟣 超管 |
| 管理員 | POST | `/api/v1/admin/admins` | 🟣 超管 |

> 所有 `/api/v1/admin/*` 端點都需要管理員 JWT；超管端點額外傳入 `RequireSuperAdmin()` 驗證。

## 4. Token 與安全設定
| 項目 | 值 |
|------|------|
| Token Type | JWT (`Bearer`) |
| 有效時間 | 8 小時（`expires_in = 28800`）|
| Payload | `admin_id`, `username`, `role`, `permissions` |
| 秘鑰 | `.env` 中的 `ADMIN_JWT_SECRET`（若未配置則 fallback 到 `JWT_SECRET`）|
| 鎖定策略 | 登入失敗 5 次 → 帳號鎖定 30 分鐘 |

建議：
- 生產環境使用獨立的 `ADMIN_JWT_SECRET`
- 透過 `LOG_LEVEL=info` 以上避免敏感資訊洩漏
- 監控登入失敗、敏感操作（更新密碼/刪除角色等）

## 5. 常見操作
| 操作 | 指令 / API | 權限 |
|------|-------------|------|
| 查看目前管理員 | `GET /api/v1/admin/admins` | 🟣 超管 |
| 新增管理員 | `POST /api/v1/admin/admins` | 🟣 超管 |
| 重置用戶密碼 | `PUT /api/v1/admin/users/{id}/password` | 🔴 管理員 |
| 停用使用者 | `PUT /api/v1/admin/users/{id}/status` | 🔴 管理員 |
| 還原角色 | `POST /api/v1/admin/characters/{id}/restore` | 🔴 管理員 |
| 永久刪除角色 | `DELETE /api/v1/admin/characters/{id}/permanent` | 🔴 管理員 |

## 6. 疑難排解
| 問題 | 檢查項目 | 解決方式 |
|------|----------|-----------|
| `401 Unauthorized` | Token 是否過期／缺少 `Bearer` 前綴 | 重新登入，確認 Header 格式 |
| 呼叫 `/admin/admins` 失敗 | 是否使用超管 Token | 使用 `admin` 帳號或任何 `role=super_admin` 的 Token |
| 端點回傳 404 | 路徑是否包含 `/api/v1` BasePath | 對照 `routes/routes.go` 或 Swagger |
| 無法登入 | DB 中是否已有管理員 | `SELECT username, status FROM admins;` 檢查狀態 |

---
若新增或調整管理路由，請同步更新上方表格與 `API_PROGRESS.md`，確保文件與程式一致。
