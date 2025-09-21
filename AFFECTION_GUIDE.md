# 好感度系統指南

## 1. 資料欄位對照
| 欄位 | 型別 | 預設值 | 說明 |
|------|------|--------|------|
| `affection` | `int` | 50 | 0–100 好感度分數 |
| `mood` | `text` | `neutral` | 當前心情描述，直接覆蓋於 AI JSON |
| `relationship` | `text` | `stranger` | 角色與使用者的關係語意 |
| `intimacy_level` | `text` | `distant` | 親密等級描述 |
| `total_interactions` | `int` | 0 | 每次 AI 回應後自動 +1 |
| `last_interaction` / `updated_at` | `timestamp` | `now()` | 互動時間戳 |
| `emotion_data` | `jsonb` | `{}` | 供自訂歷史事件使用（預設流程未寫入） |

## 2. AI JSON 欄位
| 欄位 | 是否必填 | 解析來源 | 作用 | 回退行為 |
|------|-----------|----------|------|-----------|
| `content` | 是 | AI 回覆本文 | 顯示給用戶的訊息 | 無內容時視為解析失敗 |
| `emotion_delta.affection_change` | 否 | AI JSON / 混合格式 | 直接加總至 `affection` | 缺失時預設為 `0`（JSON 解析成功但缺欄位則為 `1`） |
| `mood` | 否 | 同上 | 覆蓋 `relationships.mood` | 缺值保持原樣 |
| `relationship` | 否 | 同上 | 覆蓋 `relationships.relationship` | 缺值保持原樣 |
| `intimacy_level` | 否 | 同上 | 覆蓋 `relationships.intimacy_level` | 缺值保持原樣 |
| `metadata` | 否 | 同上 | 附加資訊，僅寫入訊息 JSON | 不影響資料庫 |

> 支援兩種格式：純 JSON 或「對話 --- metadata」混合格式。解析失敗時會回退為純文字回覆（`EmotionDelta` = 0）。

## 3. 更新流程
| 階段 | 函式 | 行為 | 備註 |
|------|------|------|------|
| 取得目前狀態 | `getAffectionFromDB` | 查詢 `relationships`。若找不到會新建一筆，預設好感度 50 | 保證後續流程有資料 |
| 解析 AI 回覆 | `parseJSONResponse` → `updateAffection` | 計算 `current + affection_change`，範圍限制 0–100 | Prompt 的變化幅度需自行管控 |
| 寫入訊息 | `saveAssistantMessageToDB` | 儲存 AI 回覆，並更新 `affection`、`mood`、`relationship`、`intimacy_level`、`total_interactions`、時間戳 | `EmotionDelta` 不存在時預設為 0 |
| 歷史記錄（可選） | `RelationshipDB.AddEmotionHistory` | 追加至 `emotion_data.history`（需自行呼叫並更新欄位） | 預設流程未啟用 |

## 4. API 對照表
| 端點 | 回傳欄位 | 來源 | 備註 |
|------|----------|------|------|
| `GET /relationships/chat/{chat_id}/status` | `affection`, `mood`, `relationship`, `intimacy_level`, `total_interactions` | 直接讀取 `relationships` | `mood_intensity`, `mood_description` 為 handler 計算欄位 |
| `GET /relationships/chat/{chat_id}/affection` | `current`, `level_name`, `next_level_threshold` 等 | `affection` + handler 對照表 | Max 固定 100，門檻映射寫在 handler |
| `GET /relationships/chat/{chat_id}/history` | `history[]`, `statistics` | `emotion_data.history` | 若沒有外部流程寫入，回傳空陣列 |

## 5. 歷史紀錄欄位（若啟用）
| 欄位 | 說明 |
|------|------|
| `timestamp` | 事件時間 |
| `trigger_type` | 觸發來源（自訂字串，例如 `user_message`） |
| `trigger_content` | 觸發內容摘要 |
| `old_affection` / `new_affection` | 更新前後好感度 |
| `affection_change` | 差值（`new - old`） |
| `old_mood` / `new_mood` | 心情變化 |

> `AddEmotionHistory` 會維持最多 50 筆 history。若要存盤，請在 `saveAssistantMessageToDB` 中將 `EmotionData` 一併更新。

## 6. 常見調整參考
| 需求 | 建議調整位置 | 說明 |
|------|--------------|------|
| 修改初始好感度 | `handlers/chat.go` 新增聊天時的 `relationshipDB` | 目前預設 50 |
| 限制好感度波動 | Prompt / JSON Schema | 程式側只做 0–100 夾取 |
| 啟用歷史記錄 | `saveAssistantMessageToDB` 內部 | 呼叫 `AddEmotionHistory` 並更新 `emotion_data` 欄位 |
| 顯示更多統計 | 調整 `handlers/relationship.go` | 可根據 `total_interactions`、`updated_at` 計算 |

---
所有資料均以 `relationships` 表為唯一真實來源；若實作有變動，請同步確認 `services/chat_service.go` 與本指南內容。
