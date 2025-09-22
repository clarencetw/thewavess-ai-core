# ⚙️ 環境設定指南

建議先複製 `.env.example` → `.env`，再依下列步驟與對照表填寫參數；完整清單與註解請參考 `.env.example`。

> 📋 **相關文檔**: 完整文檔索引請參考 [DOCS_INDEX.md](./DOCS_INDEX.md)

## 1. 建議設定流程
1. 複製 `.env.example` 為 `.env`
2. 填寫必填變數：OpenAI、資料庫、JWT
3. 依需求啟用 Grok（高強度 NSFW），確保 NSFW 語料路徑正確
4. 若需 TTS 或 Mistral，補上對應金鑰
5. 調整伺服器埠、Log、CORS 等安全設定
6. （若更新 NSFW 語料）執行 `make nsfw-embeddings`
7. 執行 `make db-setup` 或 `make fresh-start`，再跑 `./tests/test-all.sh`

## 2. 必填變數
| 類別 | 變數 | 範例 |
|------|------|------|
| OpenAI | `OPENAI_API_KEY` | `sk-xxxx` |
| 資料庫 | `DB_HOST=localhost`, `DB_PORT=5432`, `DB_USER=postgres`, `DB_PASSWORD=***`, `DB_NAME=thewavess_ai_core` |
| JWT | `JWT_SECRET=your-super-secret` | 管理端可另設 `ADMIN_JWT_SECRET` |

## 3. AI 服務設定
| 類別 | 變數 / 預設 | 說明 |
|------|--------------|------|
| OpenAI | `OPENAI_API_URL=https://api.openai.com/v1` | 可自訂代理 |
|  | `OPENAI_MODEL=gpt-4o` | L1-L3 對話模型 |
|  | `OPENAI_MAX_TOKENS=1200`、`OPENAI_TEMPERATURE=0.8` | Token 限制與創意度 |
| Grok | `GROK_API_KEY` | 未設定時無法處理 L4/L5 |
|  | `GROK_API_URL=https://api.x.ai/v1`、`GROK_MODEL=grok-4-fast`、`GROK_MAX_TOKENS=2000`、`GROK_TEMPERATURE=0.9` |
| Mistral | `MISTRAL_API_KEY`、`MISTRAL_MODEL=mistral-medium-latest` | 目前保留介面，預設不啟用 |
| TTS | `TTS_API_KEY` | 未設定時沿用 `OPENAI_API_KEY` |

## 4. NSFW RAG 設定
| 變數 | 預設 | 說明 |
|------|------|------|
| `NSFW_CORPUS_DATA_PATH` | `configs/nsfw/corpus.json` | 語料檔 |
| `NSFW_CORPUS_EMBEDDING_PATH` | `configs/nsfw/embeddings.json` | 預計算向量 |
| `NSFW_RAG_LOCALE` | `zh-Hant` | 語系過濾 |
| `NSFW_RAG_TOP_K` | `4` | 聚合筆數 |
| `NSFW_RAG_LEVEL_THRESHOLDS` | `5:0.55,4:0.42,3:0.30,2:0.18,1:0.10` | 門檻表（可覆寫）|
| `NSFW_EMBED_MODEL` | `text-embedding-3-small` | Embedding 模型 |
| `NSFW_EMBED_TIMEOUT_MS` | `2000` | Embedding API timeout (ms) |

## 5. 伺服器與日誌
| 變數 | 預設 | 功能 |
|------|------|------|
| `PORT` | `8080` | HTTP 監聽埠 |
| `GIN_MODE` | `debug` | `debug` / `release` |
| `ENVIRONMENT` | `development` | 自訂環境標記 |
| `LOG_LEVEL` | `debug` | 建議生產改 `info` |
| `API_HOST` | `localhost:8080` | Swagger Host 欄位 |

## 6. CORS 與安全
| 變數 | 預設 | 說明 |
|------|------|------|
| `CORS_ALLOWED_ORIGINS` | `*` | 生產環境請改為白名單 |
| `CORS_ALLOWED_METHODS` | `GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS` | 允許方法 |
| `CORS_ALLOWED_HEADERS` | 常見標頭 | 允許請求標頭 |
| `CORS_EXPOSED_HEADERS` | （空） | 暴露的回應標頭 |
| `ADMIN_JWT_SECRET` | （空） | 設定後管理員 JWT 使用独立簽名 |

## 7. 其他常用設定
| 功能 | 相關變數 | 備註 |
|------|----------|------|
| TTS | `TTS_API_KEY`, `OPENAI_API_KEY` | 未設定時沿用 OpenAI 金鑰 |
| 聊天創意度 | `OPENAI_TEMPERATURE`, `GROK_TEMPERATURE` | 視模式需求調整 |
| Docker 部署 | `.env` 與 `docker-compose.yml` | Compose 會載入 `.env` |

## 8. 檢查清單
- [ ] `.env` 已複製並填妥必填變數
- [ ] Grok/TTS 等選用服務已視需求填寫
- [ ] 若更新 NSFW 語料，已執行 `make nsfw-embeddings`
- [ ] `make db-setup` 或 `make fresh-start` 成功
- [ ] `./tests/test-all.sh` 全數通過
- [ ] 生產環境將 `GIN_MODE=release`、限制 `CORS_ALLOWED_ORIGINS`

---
新增或移除環境變數時，請同步更新 `.env.example`、本文件以及相關部署手冊，確保團隊設定一致。
