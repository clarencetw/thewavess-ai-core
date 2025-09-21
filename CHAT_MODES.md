# Chat Modes 指南

聊天會話目前支援兩種模式：`chat`（預設）與 `novel`。模式控制 Prompt 指引與輸出風格，但 NSFW 分級與引擎路由另由 `ChatService` 處理。

## 1. 模式總覽
| 模式 | 特色 | 典型用途 | 預設字數 | Prompt 指引來源 |
|------|------|----------|-----------|------------------|
| `chat` | 輕鬆互動、句子精簡、溫暖陪伴 | 日常聊天、快速回應 | 約 100 字 | `BasePromptBuilder.GetModeGuidance`（chat 分支）|
| `novel` | 文學化敘事、段落豐富、沉浸式體驗 | 敘事型互動、角色扮演 | 約 300 字 | `BasePromptBuilder.GetModeGuidance`（novel 分支）|

## 2. Prompt 差異
| 元素 | `chat` 模式 | `novel` 模式 |
|------|-------------|--------------|
| 結構 | `*動作* + 對話` | `*動作* → 對話 → *動作* ...` 多段落 |
| 重點 | 情感連結、陪伴感 | 場景描寫、心理刻畫、氛圍營造 |
| 語氣 | 親近、口語化 | 文學敘事、細膩描寫 |
| 動作描述 | 簡潔、強調親密細節 | 詳盡，含環境與多感官細節 |
| 視角 | 對話導向 | 第三人稱敘述、舞台化呈現 |

## 3. 引擎搭配
| 模式 | OpenAI (L1-L3) | Grok (L4-L5) |
|------|----------------|--------------|
| `chat` | 使用 OpenAI 專用 prompt，注重情感智慧與陪伴節奏 | 使用 Grok 專用 prompt，保留角色語氣並強調創意 |
| `novel` | 小說敘事模板，輸出長篇段落與心理描寫 | 放大文學風格，同時遵守 NSFW 導引與角色設定 |

## 4. AI 回應 JSON 合約
> 來源：`services/prompt_base.go:GetStrictJSONContract`。無論使用哪種模式或角色，AI 最終輸出都必須符合下列 JSON 結構。

| 欄位 | 必填 | 允許值 / 範例 | 說明 |
|------|------|----------------|------|
| `content` | ✅ | `"*動作*對話內容"` | 需同時包含 `*動作*` 與實際對話；請保持畫面感與細節。 |
| `emotion_delta.affection_change` | ✅ | `-10` ~ `+10` | 控制好感度變化；缺值時後端會 fallback 0 或 1。 |
| `mood` | ✅ | `neutral / happy / excited / shy / romantic / passionate / pleased / loving / friendly / polite / concerned / annoyed / upset / disappointed` | 僅能選擇列出的心情。 |
| `relationship` | ✅ | `stranger / friend / close_friend / lover / soulmate` | 對應 `relationships.relationship` 欄位。 |
| `intimacy_level` | ✅ | `distant / friendly / close / intimate / deeply_intimate` | 對應 `relationships.intimacy_level`。 |
| `reasoning` | ⭕️ | 任意字串 | 說明回應原因，可留空。 |

**限制**：
- 必須輸出「純 JSON」，不可添加 Markdown code block 或額外文字。
- 即使 Prompt 要求其他欄位，AI 仍會被此合約檢查；不符合者會被後端視為解析失敗。

## 5. 模式設定方式
| 操作 | 說明 |
|------|------|
| 建立會話 | `POST /api/v1/chats` 時 `chat_mode` 留空 → 預設 `chat` |
| 切換模式 | `PUT /api/v1/chats/{chat_id}/mode`，Body 範例：`{"chat_mode": "novel"}` |
| 查詢模式 | `GET /api/v1/chats/{chat_id}` 回傳的 `chat_mode` 欄位 |

## 6. 注意事項
| 項目 | 說明 |
|------|------|
| 模式持久性 | 模式存於 `chats.chat_mode`，同一會話持續生效，除非手動切換 |
| NSFW 分級 | 模式不會改變 NSFW 判定；分級結果仍由語意檢測決定 |
| 歷史摘要 | Prompt 會依模式調整示例，但取樣最近 5~6 筆訊息的邏輯相同 |
| 引擎相容 | 兩種模式皆支援 OpenAI / Grok；Mistral 目前保留介面（預設不啟用）|

---
若新增模式或調整 JSON 合約內容，請同步更新 `BasePromptBuilder`、前端 `chat_mode` 控制及本文件，避免文件與實作不一致。
