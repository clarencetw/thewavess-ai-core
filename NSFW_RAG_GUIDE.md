# NSFW RAG 指南

本指南說明目前系統採用的 NSFW 語意檢索流程、設定要點，以及語料擴充與 fallback 行為，供維護與擴充時參考。

## 架構總覽
- **分類器**：`services/nsfw_classifier.go` 使用 OpenAI `text-embedding-*` 模型產生向量，於記憶體內用餘弦相似度比對 `configs/nsfw/rag_corpus.json` 語料，輸出 L1~L5 等級與命中 chunk。
- **聊天服務**：`services/chat_service.go` 在 `analyzeContent` 中取得 RAG 判定，L4/L5 時標記 `ShouldUseGrok` 並觸發 sticky；同時記錄 `rag_chunk:<id>` 與 `reason` 供審查。`IsNSFW` 以 L4 為門檻（只有 Grok 處理高強度內容）。
- **Fallback**：若 OpenAI 回傳拒絕錯誤或拒絕句（例：「抱歉，我無法協助處理此請求。」），自動切換 Grok 生成並刷新 sticky。

## 語料格式（`configs/nsfw/rag_corpus.json`）
每筆語料須包含：
- `id`：唯一識別碼（建議以用途命名，如 `explicit_penetration`）。
- `level`：整數 1~5。
- `tags`：標籤陣列（如 `explicit`, `dirty_talk`, `block`）。
- `locale`：目前支援 `zh-Hant`、`en`。
- `text`：核心片段，可含多個詞彙或情境描述。
- `reason`：分類器輸出的代號，需與程式中使用的阻擋邏輯（如 `illegal_underage`）一致。
- `version`：版本字串，方便追蹤語料更新。

> ⚠️ 語料更新須人工審核，並記錄 reviewer / 影響範圍；若新增 `reason`，需確認 `services/chat_service.go` 中的非法判斷是否要同步擴充。

## 分級與路由
- L1~L3：視為安全或中等強度對話，預設仍交給 OpenAI（只在 prompt 中保留語料，但不視為 `IsNSFW=true`）。
- L4~L5：明確露骨內容，`IsNSFW=true`、`ShouldUseGrok=true`，並寫入 sticky map 以維持 Grok 連線 5 分鐘（期間若再次出現 L4/L5 會刷新 TTL）。
- `illegal_content`：當 `reason` 落在未成年／性暴力／亂倫／獸交等代碼時，自動加入標籤，後續生成層會直接拒絕。

## OpenAI Fallback 行為
- `isOpenAIContentRejection(err)`：以錯誤訊息關鍵字判斷 OpenAI 拒絕，立即切到 Grok。
- `isOpenAIRefusalContent(responseText)`：成功回應但內容是拒絕語（中英文），也會切換 Grok。可持續更新該列表並加註註解標記已驗證字串。

## 配置與環境變數
- `NSFW_RAG_CORPUS_PATH`（預設 `configs/nsfw/rag_corpus.json`）
- `NSFW_RAG_TOP_K`（預設 4）
- `NSFW_RAG_LEVEL_THRESHOLDS`（預設 `5:0.55,4:0.42,3:0.30,2:0.18,1:0.10`）
- `NSFW_EMBED_MODEL`（預設 `text-embedding-3-small`）
- `NSFW_EMBED_TIMEOUT_MS`（預設 2000）

> 系統僅支援 OpenAI 嵌入，若缺少 `OPENAI_API_KEY` 會直接 fatal。語料調整請於 PR 記錄版本與測試內容，方便追溯。
