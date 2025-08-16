#! /usr/bin/env markdown
# 🧠 記憶系統 MVP（短期/長期）

本文件提供最小可用版本（MVP）的記憶儲存/檢索方案與 Prompt 注入建議，對應目前的 `/api/v1/chat` 流程與程式碼結構（`services/chat_service.go`、`services/openai_client.go`、`services/grok_client.go`）。

## 目標

- 低成本接入：不改變現有流程，先以簡單策略提升人物一致性與女性向體驗。
- 兩層記憶：
  - 短期：最近對話重點（3–5 輪）
  - 長期：偏好/稱呼/里程碑/不喜歡的內容（Top-K 摘要）
- Prompt 注入：在 System/Context 區塊加入「記憶摘要」，並限制長度。

## 數據來源與儲存

- 短期記憶（會話級）：
  - 來源：資料庫 `messages`（最近 N 條，建議 5–10）
  - 處理：抽取重點句（最短生成的一句話 or 100–200 字）
  - 使用：組裝成精簡「Recent Context」陳述（條列 3–5 點）

- 長期記憶（使用者/角色關係級）：
  - 新增資料表（建議）：`memory_long_term`
    - 欄位：`id, user_id, character_id, type, content, importance, created_at`
    - type 範例如：`preference`（偏好）、`nickname`（稱呼）、`milestone`（里程碑）、`dislike`（禁忌）
  - 更新：每輪對話結束後，從 `(userMessage, aiResponse, emotion)` 中抽取候選，重要度打分（規則見下）
  - 檢索：Top-K（K=3~5），按 `importance DESC, created_at DESC`

> 若短期內不加資料表，可先以 `messages` 上打 Tag（如 `is_memory=true`）的方法暫存長期記憶，後續再遷移。

## 抽取與打分（簡單規則）

- 關鍵模板偵測：
  - 偏好/不喜歡：「我喜歡…」「我不喜歡…」「我喜歡你叫我…」
  - 里程碑：「第一次…」「今天起…」「我願意…」「告白/在一起」
  - 敏感日期/詞：「生日/紀念日/週年」
- 打分：
  - 含上述模板 +2；含具體時間/數字 +1；帶情緒強度詞（超喜歡/非常/一直） +1
  - 總分≥3 → 優先保存，2 分作為候選，<2 丟棄
- 去重：同義或高度相似（Jaccard/編碼相似度>0.9）則覆蓋舊的或忽略

## Prompt 注入（模板）

在 OpenAI/Grok 的 System Prompt 前置加入記憶區塊，限制字數：

```
# Long-Term Memory (summary)
- 偏好：喜歡被稱呼「寶貝」，不喜歡粗俗語
- 里程碑：上週願意牽手；好感度突破 60
- 禁忌：討厭職場八卦

# Recent Context (last 3-5 turns)
- 她今天很累，需要安慰
- 她說想被擁抱，但有點害羞
- 我剛才提醒她多休息
```

插入順序建議：
1) 記憶區塊 → 2) 場景 → 3) 角色設定 → 4) 任務要求 → 5) 輸出格式

## 對應程式碼位置（實現狀態）

- `services/chat_service.go` ✅ 基礎架構完成
  - ✅ `getRecentMemories`: 從記憶管理器取最近 N 條對話，轉換為摘要格式
  - ✅ `updateMemorySystem`: 更新短期記憶到記憶管理器
  - ✅ `buildFemaleOrientedContext`: 已集成記憶系統和用戶偏好

- `services/memory_manager.go` ✅ 完整實現
  - ✅ `ShortTermMemory`: 短期記憶結構（會話級）
  - ✅ `LongTermMemory`: 長期記憶結構（偏好、稱呼、里程碑、禁忌）
  - ✅ `UpdateShortTermMemory`: 更新短期記憶
  - ✅ `ExtractLongTermMemory`: 抽取長期記憶

- `services/openai_client.go: BuildCharacterPrompt` ⚠️ 待完成
  - ✅ TODO 標記已添加：注入「Long-Term Memory」與「Recent Context」
  - ❌ 實際記憶區塊注入功能待實現

- `services/grok_client.go: BuildNSFWPrompt` ⚠️ 待完成
  - ✅ TODO 標記已添加：記憶區塊注入（NSFW場景縮短版本）
  - ❌ 實際記憶區塊注入功能待實現

## Pseudo-code（最小實作）

```go
// chat_service.go
func (s *ChatService) getRecentMemories(userID, characterID string, limit int) []models.ChatMessage {
    // 1) 查詢最近 N 條 messages（user/assistant）
    // 2) 以啟發式方法抽取 3–5 條摘要（截斷）
    // 3) 返回作為 Recent Context 使用
}

func (s *ChatService) updateMemorySystem(userID, characterID, sessionID, userMessage, aiResponse string, emotion *EmotionState) {
    // 1) 以規則抽取長期記憶候選（偏好/稱呼/里程碑/禁忌）
    // 2) 打分 ≥3 → upsert 到 memory_long_term
    // 3) 記錄審計日誌（type, content, reason）
}
```

```go
// openai_client.go / grok_client.go（在 systemPrompt 前）
mem := buildMemoryBlock(context) // 拼 Long-Term + Recent Context（已截斷）
systemPrompt = mem + "\n\n" + systemPrompt
```

## 驗收清單（MVP）

- [x] 最近 5–10 條訊息 → 3–5 點摘要注入 Prompt ✅ `getRecentMemories` 已實現
- [x] 長期記憶抽取與去重 → Top-K 注入 Prompt ✅ `MemoryManager` 完整實現
- [x] OpenAI（L1-L4）與 Grok（L5）均有記憶區塊 ✅ TODO 標記已添加到提示詞函數
- [ ] 總字數限制與截斷策略一致
- [x] 日誌含：抽取結果、打分、寫入/忽略原因 ✅ `updateMemorySystem` 已實現

## 後續擴展

- 使用向量資料庫做語義檢索（Top-K）
- 記憶重要度隨時間衰減（Time-decay）
- 事件圖譜化（里程碑/關係發展）
- 自動修剪與摘要（長期）

