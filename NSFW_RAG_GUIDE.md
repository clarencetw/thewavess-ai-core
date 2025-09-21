# NSFW RAG 快速參考

> 詳細說明請參閱 [NSFW_GUIDE.md](./NSFW_GUIDE.md)。此文件著重於決策表與快速查詢。

## 一覽表
| 主題 | 重點 |
|------|------|
| 分級來源 | `services/nsfw_classifier.go` (`ClassifyContent`) |
| 整合入口 | `ChatService.analyzeContent` / `selectAIEngine` |
| 黏滯保護 | `markNSFWSticky` 讓 Grok 連線維持 5 分鐘 |
| 語料資源 | `configs/nsfw/corpus.json` + `configs/nsfw/embeddings.json` |

## 引擎路由決策表
| 條件 | 使用引擎 | Sticky 狀態 | 備註 |
|------|----------|-------------|------|
| 角色 Tag 含 `nsfw` / `adult` | Grok | 不變 | 角色層級固定視為 NSFW |
| 會話 Sticky 尚未過期 | Grok | 刷新到期時間 | Sticky TTL 預設 5 分鐘 |
| 分級 L4 或 L5 | Grok | 標記 Sticky | 同時紀錄命中片段與 reason |
| OpenAI API 拒絕 / 回傳拒絕語 | Grok | 標記 Sticky | 由 `isOpenAIContentRejection` / `isOpenAIRefusalContent` 偵測 |
| 分級 L2-L3 | OpenAI | 不變 | 目前預設仍走 OpenAI（保留 Mistral 擴充空間） |
| 分級 L1 | OpenAI | 不變 | 一般對話 |

## `ContentAnalysis` 欄位對照
| 欄位 | 來源 | 說明 |
|------|------|------|
| `IsNSFW` | `level >= 4` | 高強度 NSFW 判斷 |
| `Intensity` | `ClassificationResult.Level` | 1~5 等級 |
| `ShouldUseGrok` | `level >= 4` | 提供給呼叫端的快速判斷 |
| `Categories` | 固定標籤 + `reason` | 例如 `semantic_rag_analysis`, `rag_chunk:xxx`, `illegal_content` |
| `Confidence` | `ClassificationResult.Confidence` | 0~0.99，相似度分數 |

## 違法內容阻擋表
| `reason` / `category` | 行為 |
|-----------------------|------|
| `illegal_underage`, `illegal_underage_en` | `generatePersonalizedResponse` 直接回覆拒絕訊息 |
| `bestiality` | 同上 |
| `sexual_violence_or_incest`, `incest_family_roles`, `incest_step_roles_en` | 同上 |
| `rape` | 同上 |

## 指令備忘
| 動作 | 指令 | 說明 |
|------|------|------|
| 產生/更新向量 | `make nsfw-embeddings` | 修改 `corpus.json` 後執行 |
| 檢查語料/向量 | `make nsfw-check` | 確認兩檔案筆數與版本一致 |
| 查看當前門檻 | `NSFW_RAG_LEVEL_THRESHOLDS` | `.env` 覆寫，預設 `5:0.55,4:0.42,3:0.30,2:0.18,1:0.10` |
| Sticky TTL | 程式常數 `nsfwStickyTTL` | 目前固定 5 分鐘，需改程式碼調整 |

---
同步維護本文件與原始程式可避免文件老化；若有行為調整，請優先更新 `NSFW_GUIDE.md` 並回收此處表格。
