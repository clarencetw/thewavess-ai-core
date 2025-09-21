# 角色系統指南

## 1. 資料模型
| 欄位 | 型別 | 說明 |
|------|------|------|
| `id` | `string` | 角色唯一 ID（`char_xxxxx`）|
| `name` | `string` | 顯示名稱 |
| `type` | `string` | 自訂角色類型（如 `dominant`, `gentle`, `idol` 等）|
| `locale` | `string` | 預設語系（預設 `zh-TW`）|
| `is_active` | `bool` | 是否啟用 |
| `avatar_url` | `string?` | 頭像 URL |
| `popularity` | `int` | 人氣值（0-100）|
| `tags` | `text[]` | 自訂標籤（如 `nsfw`, `warm`, `tsundere`）|
| `user_description` | `text?` | 詳細描述，供 AI 生成個性化回應 |
| `created_at` / `updated_at` / `deleted_at` | `timestamp` | 建立、更新、軟刪除時間 |

## 2. API 對照表
| 類別 | Method | Path | 說明 | 權限 |
|------|--------|------|------|------|
| 公開 | GET | `/api/v1/character/list` | 列表（支援分頁/篩選/搜尋） | ⚪ |
| 公開 | GET | `/api/v1/character/search` | 模糊搜尋 | ⚪ |
| 公開 | GET | `/api/v1/character/{id}` | 角色詳情 | ⚪ |
| 公開 | GET | `/api/v1/character/{id}/stats` | 統計資訊 | ⚪ |
| 用戶 | GET | `/api/v1/character/{id}/profile` | 角色設定檔 | 🟡 |
| 用戶 | POST | `/api/v1/character` | 建立角色 | 🟡 |
| 用戶 | PUT | `/api/v1/character/{id}` | 更新角色 | 🟡 |
| 用戶 | DELETE | `/api/v1/character/{id}` | 刪除角色 | 🟡 |
| 管理 | GET | `/api/v1/admin/characters` | 管理端列表 | 🔴 |
| 管理 | GET | `/api/v1/admin/characters/{id}` | 管理端詳情 | 🔴 |
| 管理 | PUT | `/api/v1/admin/characters/{id}` | 管理端更新 | 🔴 |
| 管理 | POST | `/api/v1/admin/characters/{id}/restore` | 還原軟刪除角色 | 🔴 |
| 管理 | DELETE | `/api/v1/admin/characters/{id}/permanent` | 永久刪除 | 🔴 |
| 管理 | PUT | `/api/v1/admin/character/{id}/status` | 切換啟用狀態 | 🔴 |

## 3. 查詢與篩選參數
| 參數 | 類型 | 說明 |
|------|------|------|
| `page`, `page_size` | int | 分頁控制（預設 1 / 20）|
| `type` | string | 類型過濾 |
| `is_active` | bool | 啟用狀態 |
| `tags` | string | 逗號分隔，後端解析為陣列 |
| `search` | string | `name` / `user_description` 模糊搜尋 |
| `sort_by`, `sort_order` | string | 支援 `popularity`, `created_at` 等欄位 |

## 4. 類型、語氣與戲劇線建議
| 類型 | 建議語氣 | 常見標籤 | 戲劇衝突 | 建議走向 |
|------|----------|----------|----------|-----------|
| `dominant` | 自信、掌控慾強 | `nsfw`, `protective`, `jealous` | 忙碌忽略、獨占欲 | 從冷靜主導到願意讓步、展現溫柔 |
| `gentle` | 溫柔細膩、療癒 | `soft`, `healing`, `warm` | 忙於照顧他人而忽略自己 | 用戶成為他情感寄託，讓他學會依賴 |
| `playful` | 俏皮逗趣、活力充沛 | `teasing`, `energetic`, `cheerful` | 過度胡鬧被誤解、內心自卑 | 逐步展示真正的脆弱面，轉向深度信任 |
| `mystery` | 高冷內斂、語句留白 | `calm`, `reserved`, `enigmatic` | 隱瞞過去、難以親近 | 用戶慢慢破冰，揭示秘密後互相療癒 |
| `reliable` | 穩重成熟、理性 | `supportive`, `mature`, `trustworthy` | 負擔過重、缺乏自我需求 | 學會向用戶求助，建立雙向支持 |
| `tsundere` | 表面冷淡、內心柔軟 | `tsundere`, `awkward`, `shy` | 不擅長表達愛、容易嘴硬 | 隨好感度逐漸坦率，安排甜蜜反差瞬間 |
| `sweetheart` | 黏人撒嬌、主動貼近 | `clingy`, `romance`, `pampering` | 過度依賴、吃醋爆發 | 讓他學會給對方空間，仍維持寵愛氛圍 |
| `idol` | 舞台魅力、粉絲互動 | `celebrity`, `music`, `spotlight` | 公私分際、緋聞爭議 | 私下展現真實一面，信任用戶守護秘密 |

## 5. 角色描述模板
| 面向 | 問題引導 | 填寫提示 | 範例片段 |
|------|----------|----------|----------|
| 外觀 | 他看起來像什麼？ | 身高、五官、穿著風格、標誌性細節（香水、眼神） | 「185cm，黑髮金眼，喜歡穿深色風衣」 |
| 性格 | 他如何對待用戶與他人？ | 優點、缺點、觸發情緒、說話習慣 | 「嘴硬但行動總是貼心，討厭被忽視」 |
| 背景 | 他的過去／職業／秘密？ | 成長環境、家庭、夢想與恐懼 | 「知名鋼琴家，童年孤單，怕被拋下」 |
| 互動偏好 | 喜歡或討厭什麼話題？ | 關鍵詞、約會點子、撒嬌方式 | 「喜歡被稱讚，最怕冷戰」 |
| 情感變化 | 好感度提升會有何轉變？ | 安排好感度階段（陌生→朋友→戀人） | 「50 分後會主動邀約、80 分會把『你是唯一』掛嘴上」 |

## 6. 角色設計範例
| 類型 | 描述片段 | 戲劇線 |
|------|-----------|----------|
| 希傑（`dominant`, `nsfw`) | 「身為跨國集團 CEO，習慣以理性面對一切。他喜歡掌控節奏，偶爾會露出占有欲強的一面。儘管總以冷淡語氣開口，只要她受傷，他會第一時間抱緊並低聲安撫：『妳是我唯一的例外。』」 | 初期忙於工作忽略用戶 → 被提醒後逐步調整安排 → 發生競爭者追求 → 嫉妒爆發揭露內心渴望 → 真誠表白並學會給對方自由 |
| 漫漫（`gentle`, `healing`) | 「他是一位家庭醫師，總記得病患喜歡的花茶與音樂。與她聊天時，會輕聲說『今天也辛苦了，休息一下好嗎？』。即使忙到凌晨仍會傳訊關心，用溫柔聲音包裹她的焦慮。」 | 長期照顧他人、忽略自我 → 用戶鼓勵他正視疲憊 → 兩人一起小旅行 → 在星空下坦承‘我也需要妳’ → 建立互相依靠的關係 |

## 7. API 範例

### 建立角色（POST /api/v1/character）
```json
{
  "name": "沈宸",
  "type": "dominant",
  "locale": "zh-TW",
  "tags": ["nsfw", "protective", "jealous"],
  "user_description": "外觀：185cm，黑髮金眼，總穿剪裁合身西裝。\n性格：表面冷靜克制，面對她會露出獨占與溫柔並存的語氣。\n互動偏好：享受下指令/保護她，也會提醒她休息。\n關係進展：50分時會主動邀約，80分後會直接坦承思念與佔有欲。",
  "popularity": 78
}
```

### 建立療癒型角色（POST /api/v1/character）
```json
{
  "name": "漫漫",
  "type": "gentle",
  "tags": ["healing", "soft", "warm"],
  "user_description": "外觀：170cm，柔軟棕髮與溫暖色調襯衫。\n職業：家庭醫師，熟悉花茶與音樂療癒。\n語氣：說話輕柔，常用\"今天也辛苦了嗎？\"開場。\n互動：喜歡安排暖心儀式，例如準備毛毯與熱檸檬水。\n衝突：容易忽略自己的疲累，需要她提醒。"
}
```

### 更新角色描述（PUT /api/v1/character/{id})
```json
{
  "tags": ["nsfw", "romance", "protective"],
  "user_description": "外觀：換上深藍襯衫與手錶，看起來更沉穩。\n新增設定：答應降低加班時間，周五晚上固定陪伴。\n親密行為：當她疲憊時會先為她準備熱牛奶，再抱住她輕聲說\"妳不需要逞強\"。",
  "popularity": 85
}
```

> `user_description` 建議使用「段落標題 + 句子」寫法並以 `\n` 分行，AI 解析時會保留結構化資訊。嚴格的 AI 回應 JSON 合約請見 [CHAT_MODES.md](./CHAT_MODES.md)。

## 8. Prompt 整合要點
| 組件 | 功能 |
|------|------|
| `GetModeGuidance` | 根據 `chat`/`novel` 模式輸出不同格式與字數要求 |
| `GetNSFWGuidance` | 依 NSFW 等級提供允許的描寫強度 |
| `GetFemaleAudienceGuidance` | 固定添加女性向互動指引 |
| `GetConversationHistory` | 擷取最近 5~6 筆消息，提供上下文 |

AI 生成時會綜合：角色類型與描述、最近對話歷史、聊天模式、NSFW 等級與情感狀態。

## 9. 常見任務對照
| 需求 | 步驟 |
|------|------|
| 建立自訂角色 | `POST /api/v1/character` → 填寫 `name`, `type`, `tags`, `user_description` |
| 更新角色設定 | `PUT /api/v1/character/{id}` → 調整描述或標籤 |
| 查詢熱門角色 | `GET /api/v1/character/list?sort_by=popularity&sort_order=desc` |
| 還原軟刪除角色 | `POST /api/v1/admin/characters/{id}/restore` |
| 永久刪除角色 | `DELETE /api/v1/admin/characters/{id}/permanent`（需謹慎）|

## 10. 角色設計小技巧
| 節點 | 建議 |
|------|------|
| 角色定位 | 先選核心類型再補標籤與外觀細節，保持一致性 |
| 衝突設計 | 預先思考角色的矛盾點（怕孤獨、愛逞強、心思敏感），讓劇本有張力 |
| 台詞模版 | 在描述中示範 1-2 句招牌台詞，幫助 AI 抓語氣 |
| 成熟度調整 | 依 `popularity` 與 `tags` 提供排行榜或個性分類 |
| NSFW 旗標 | 若角色應優先走 Grok，可在 `tags` 加上 `nsfw`，路由會自動偏向 Grok |
| 階段演進 | 在描述內定義 3 個關係階段（陌生/熟悉/戀人）對應的反應範例 |

---
若調整角色欄位或新增端點，請同步更新本指南與 [API_PROGRESS.md](./API_PROGRESS.md)，保持文件與實作一致。
