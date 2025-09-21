# 關係系統指南

## 系統概述

本系統採用**AI驅動的關係管理**，透過分析對話內容即時更新用戶與角色之間的互動狀態。

### 🎯 核心目標
- **智能化**：AI 自動評估情感變化，無需硬式規則
- **個性化**：依據對話歷史與角色特性調整反應
- **持久化**：完整記錄關係發展與情感事件
- **可追溯**：提供清楚的歷史來源與統計資料

## 關係模型

### 📊 核心維度
- **好感度 (Affection)**：0-100 數值，反映情感親密程度
- **心情狀態 (Mood)**：當前情緒類型（happy、sad、excited…）
- **關係類型 (Relationship)**：關係的語意描述
- **親密等級 (Intimacy Level)**：互動的深度層級

### 🎭 關係等級系統
| 好感度範圍 | 關係類型 | 親密等級 | 特徵描述 |
| ---------- | -------- | -------- | -------- |
| 0-19       | stranger | distant  | 初次見面，維持禮貌距離 |
| 20-39      | friend   | casual   | 普通朋友，輕鬆對話 |
| 40-59      | friend   | casual   | 好朋友，相互信任 |
| 60-79      | close    | close    | 親密朋友，能進行深度交流 |
| 80-89      | intimate | intimate | 戀人關係，情感深厚 |
| 90-100     | soulmate | intimate | 靈魂伴侶，心靈相通 |

## 數據庫架構

### 🗄️ Relationships 表結構
```sql
relationships:
├── id (string, PK)
├── user_id (string, FK)
├── character_id (string, FK)
├── chat_id (string, FK, 可選)
├── affection (int)
├── mood (string)
├── relationship (string)
├── intimacy_level (string)
├── total_interactions (int)
├── last_interaction (timestamp)
├── emotion_data (jsonb)
├── created_at (timestamp)
└── updated_at (timestamp)
```

### 📝 情感歷史 (emotion_data)
```json
{
  "history": [
    {
      "timestamp": "2025-09-01T18:00:00Z",
      "trigger_type": "user_message",
      "trigger_content": "用戶表達關心",
      "old_affection": 45,
      "new_affection": 48,
      "affection_change": 3,
      "old_mood": "neutral",
      "new_mood": "happy"
    }
  ]
}
```

## API 端點

### `GET /api/v1/relationships/chat/{chat_id}/status`
- **用途**：取得指定對話的即時關係狀態
- **主要欄位**：`user_id`, `character_id`, `chat_id`, `affection`, `mood`, `mood_intensity`, `mood_description`, `relationship`, `intimacy_level`, `total_interactions`, `last_interaction_at`, `updated_at`
- **資料來源**：直接使用 `relationships` 表中的即時數據

### `GET /api/v1/relationships/chat/{chat_id}/affection`
- **用途**：檢視好感度細節與升級進度
- **主要欄位**：`current`, `level_name`, `level_tier`, `description`, `next_level_threshold`, `points_to_next`, `updated_at`
- **計算方式**：依據 affection 數值計算等級與下一門檻，移除任何硬編碼或假資料

### `GET /api/v1/relationships/chat/{chat_id}/history`
- **用途**：回顧情感變化軌跡
- **主要欄位**：`current_affection`, `total_interactions`, `history[]`
  - 每筆歷史紀錄包含：`timestamp`, `trigger_type`, `trigger_content`, `affection_before`, `affection_after`, `affection_change`, `mood_before`, `mood_after`
- **數據來源**：從 `emotion_data.history` 讀取真實 JSONB 紀錄；若無歷史資料則回傳空陣列

## 系統特色

### ✨ 智能特性
1. **動態適應**：結合 AI 分析與歷史資料調整反應
2. **個人化反應**：不同角色具備專屬情緒與好感度曲線
3. **自然進展**：避免固定腳本，追求貼近真人互動
4. **情感記憶**：重要事件會被寫入 `emotion_data` 以供後續參考

### 🔄 多會話支援
- 每個 `chat_id` 擁有獨立的關係狀態
- 如需跨會話共享，可使用 `chat_id IS NULL` 的全局紀錄

### 📊 數據洞察
- 透過歷史紀錄掌握好感度變化與觸發因素
- 利用 `total_interactions` 與 `last_interaction_at` 追蹤互動頻率

## 性能與監控

- **JSONB 索引**：確保 emotion_data 查詢效率
- **批次更新**：在訊息處理流程中一次更新 affection、mood、relationship 等欄位
- **快照式回應**：所有 handler 立即返回資料庫中的真實數據

## 實作重點

- 改採**型別化的回應結構**，移除原先的動態 Map 與假資料欄位
- 直接使用資料庫欄位，避免硬編碼描述或臨時統計
- 歷史端點僅在 emotion_data 有資料時回傳事件，維持資料可信度
- 所有時間欄位統一輸出 ISO 8601（RFC3339）格式

---

*關係系統透過 AI 與資料庫的緊密整合，提供自然且可追蹤的互動體驗。*
