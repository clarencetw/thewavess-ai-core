# NSFW 語意檢測指南

## 核心模組總覽
| 模組 | 位置 | 職責 |
|------|------|------|
| `NSFWClassifier` | `services/nsfw_classifier.go` | 讀取語料與向量、計算語意相似度、輸出等級與信心值 |
| `ChatService.analyzeContent` | `services/chat_service.go` | 呼叫分類器並產生 `ContentAnalysis`（含 `IsNSFW`、`ShouldUseGrok`、Categories） |
| `ChatService.selectAIEngine` | `services/chat_service.go` | 依等級、角色標籤與 sticky 狀態在 OpenAI / Grok 之間切換 |
| `markNSFWSticky` / `isNSFWSticky` | `services/chat_service.go` | 維護 5 分鐘的黏滯狀態，確保後續請求持續使用 Grok |
| Grok / OpenAI 客戶端 | `services/grok_client.go`、`services/openai_client.go` | 實際呼叫對應模型輸出回覆 |

## 資料資源
| 類型 | 檔案 | 內容 | 維護備註 |
|------|------|------|----------|
| 語料主檔 | `configs/nsfw/corpus.json` | 每筆語料包含 `id`, `level`, `tags`, `locale`, `text`, `reason`, `version` | 以人類可讀格式維護；新增語料後請記錄版本 |
| 預計算向量 | `configs/nsfw/embeddings.json` | 對應語料的 1536 維向量 (`id`, `embedding`, `version`) | 更新語料後必須重新產生；缺向量的語料會被跳過並產生警告 |
| 環境設定 | `.env` | 決定語料路徑、Top-K、門檻等 | 詳見下方環境變數表 |

## 分級流程
| 步驟 | 描述 | 相關程式碼 |
|------|------|-----------|
| 1. 正規化 | 轉小寫、移除零寬字元、修正常見變體 (`seggs` → `sex`) | `NSFWClassifier.normalize` |
| 2. 產生向量 | 呼叫 OpenAI Embedding API 取得輸入向量 | `embedClient.EmbedText` |
| 3. 語意比對 | 與語料向量計算 cosine similarity，依分數排序 | `NSFWClassifier.scoreAgainstCorpus` |
| 4. 聚合判斷 | 聚合 Top-K 分數並對照門檻 (`defaultThresholds`) 決定等級 | `NSFWClassifier.resolveLevel` |
| 5. 回傳結果 | 回傳 `ClassificationResult(Level, Confidence, Reason, ChunkID)` | `NSFWClassifier.ClassifyContent` |

## 分类輸出欄位
| 欄位 | 來源 | 說明 |
|------|------|------|
| `Level` | 加總後的等級判定 | L1~L5，L4/L5 視為高強度 NSFW |
| `Confidence` | 主命中語料的 similarity | 介於 0~0.99，用於 Log 與分析 |
| `Reason` | 語料 `reason` 或 `id` | 追蹤命中規則、區分非法情境 |
| `ChunkID` | 語料 `id` | 供黏滯與診斷使用 |

## ChatService 整合行為
| 邏輯節點 | 條件 / 輸入 | 行為 | 程式位置 |
|-----------|--------------|------|----------|
| `analyzeContent` | 使用者訊息 | 呼叫 `NSFWClassifier`，組裝 `ContentAnalysis`，標記 `illegal_content` 類別 | `services/chat_service.go:360+` |
| `generatePersonalizedResponse` | `ContentAnalysis.Categories` 含 `illegal_content` | 直接回傳拒絕訊息並保持好感度不變 | `services/chat_service.go:816+` |
| `selectAIEngine` | 角色標籤含 `nsfw`/`adult` | 立即使用 Grok | `services/chat_service.go:504+` |
| `selectAIEngine` | 會話 sticky 中 | 直接 Grok，並更新 sticky 過期時間 | 同上 |
| `selectAIEngine` | 等級 L4/L5 | Grok + `markNSFWSticky` | 同上 |
| `selectAIEngine` | 等級 L2/L3 | 預設 OpenAI（保留 Mistral 介面） | 同上 |
| `selectAIEngine` | 其他 | OpenAI | 同上 |

## 引擎與 Sticky 對照
| 條件 | 選擇引擎 | 附帶動作 |
|------|----------|----------|
| 角色設定為 NSFW | Grok | 不標記 sticky（角色本身已固定） |
| 已進入 sticky（5 分鐘內） | Grok | sticky 到期時間刷新 |
| 分級 L4/L5 | Grok | 呼叫 `markNSFWSticky`（聊天在 5 分鐘內都會走 Grok） |
| OpenAI API 拒絕（錯誤） | Grok | sticky，並重新組 prompt 再送 Grok |
| OpenAI 回傳拒絕文字 | Grok | sticky，重新送 Grok |
| 其他情況 | OpenAI | 無 |

## Fallback / 安全機制
| 觸發原因 | 偵測方式 | 後續處理 |
|----------|----------|----------|
| OpenAI 回傳內容包含拒絕語 | `isOpenAIRefusalContent` | 標記 sticky → 重新呼叫 Grok |
| OpenAI API 回傳內容政策錯誤 | `isOpenAIContentRejection` | 標記 sticky → 重新呼叫 Grok |
| 偵測到違法內容 (incest / 未成年等) | `analysis.Categories` 含 `illegal_content` | 直接回覆拒絕訊息，不進入生成流程 |

## 環境變數
| 變數 | 用途 | 預設 |
|------|------|------|
| `NSFW_CORPUS_DATA_PATH` | 語料檔路徑 | `configs/nsfw/corpus.json` |
| `NSFW_CORPUS_EMBEDDING_PATH` | 向量檔路徑 | `configs/nsfw/embeddings.json` |
| `NSFW_RAG_LOCALE` | 語料語系過濾 | `zh-Hant` |
| `NSFW_RAG_TOP_K` | 聚合排名數量 | `4` |
| `NSFW_RAG_LEVEL_THRESHOLDS` | 各等級門檻 | `5:0.55,4:0.42,3:0.30,2:0.18,1:0.10` |
| `NSFW_EMBED_TIMEOUT_MS` | Embedding API timeout | `2000` |
| `OPENAI_API_KEY` | Embedding / GPT 使用 | — |
| `GROK_MODEL` 等 | Grok 參數 | `.env` 配置 |

## 維護與診斷
| 操作 | 指令 | 說明 |
|------|------|------|
| 重新產生向量 | `make nsfw-embeddings` | 讀取 `corpus.json`，寫出 `embeddings.json` |
| 檢查語料與向量 | `make nsfw-check` | 確認語料/向量筆數與版本一致 |
| 觀察初始化 | 啟動服務 | Log 會輸出語料筆數與向量載入狀態 |
| 偵測缺向量 | — | Log level `WARN`，訊息為「缺少預計算的 embedding 向量」 |

---
若程式邏輯更新，請同步修訂本文件，並優先以 `services/nsfw_classifier.go`、`services/chat_service.go` 之實作為準。
