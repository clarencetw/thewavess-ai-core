# 🚀 部署指南

## 1. 環境需求與建議
| 項目 | 最低需求 | 建議 |
|------|-----------|------|
| 作業系統 | Linux / macOS | 支援 systemd 或容器化更佳 |
| Go | 1.23+ | 若使用 Docker，可不安裝本機 Go |
| PostgreSQL | 15+ | 建議與應用同區網，降低延遲 |
| 其他 | Docker / Docker Compose（可選） | 方便快速啟動整套服務 |

## 2. 快速部署流程
| 步驟 | 指令 | 說明 |
|------|------|------|
| 取得程式碼 | `git clone ... && cd thewavess-ai-core` | 下載專案 |
| 複製設定 | `cp .env.example .env` | 填寫至少 `OPENAI_API_KEY`、資料庫連線、`JWT_SECRET` |
| 安裝依賴 | `make install` | 下載 Go module 與 Swagger 工具 |
| 初始化資料庫 | `make db-setup` 或 `make fresh-start` | 建立資料表並載入 fixtures |
| 啟動服務 | `make dev` | 生成 Swagger、啟動 API（預設 8080）|
| 驗證 | 參考下方健康檢查與 Swagger 連結 | 確認服務就緒 |

## 3. Docker Compose
```bash
docker-compose up -d
# 查看狀態
docker-compose ps
# 停止
docker-compose down
```
*確保 `.env` 已填寫必填變數，`docker-compose.yml` 會引用。*

## 4. 重要端點
| 類別 | URL | 說明 |
|------|-----|------|
| 健康檢查 | `GET /health` | 簡易活性檢查（無認證）|
| Ready/Live | `GET /api/v1/monitor/ready` / `live` | Kubernetes / 負載平衡器使用 |
| 監控統計 | `GET /api/v1/monitor/stats` | JSON 格式系統指標 |
| Prometheus | `GET /api/v1/monitor/metrics` | Prometheus exposition |
| Swagger | `GET /swagger/index.html` | 自動生成 API 文件 |

## 5. 生產環境配置重點
| 類別 | 建議 |
|------|------|
| 運行模式 | `GIN_MODE=release`、`LOG_LEVEL=info`|
| 安全性 | 使用 HTTPS 反向代理（Nginx / Caddy），設定 `CORS_ALLOWED_ORIGINS` 為白名單 |
| 秘鑰管理 | 使用 Vault/Secret Manager 管理 `OPENAI_API_KEY`、`GROK_API_KEY`、`JWT_SECRET` |
| 資料庫 | 啟用自動備份、設定連線池（例如 pgbouncer）|
| 監控 | 將 `/metrics` 接入 Prometheus + Grafana，搭配 Alertmanager |

## 6. 常用 Make 指令對照
| 類別 | 指令 | 功能 |
|------|------|------|
| 開發 | `make dev` | 生成 Swagger 並啟動 API |
| 重建 | `make fresh-start` | 清理 → 安裝 → 遷移 → Fixtures |
| 測試 | `./tests/test-all.sh` | 執行 24 項整合測試 |
| 建置 | `make build` | 產出 `bin/thewavess-ai-core` |
| Docker | `make docker-build` | 建置專案映像檔 |

## 7. 監控與除錯
| 操作 | 指令 | 備註 |
|------|------|------|
| 健康檢查 | `curl http://host:8080/health` | 正常回傳 `OK` |
| 檢查 Port 占用 | `lsof -i :8080` | 若衝突請改 `PORT` 或釋放埠口 |
| 查看日誌 | `docker-compose logs -f` 或應用標準輸出 | 建議納入 ELK 或 Loki |
| 測試 API Key | `curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models` | 驗證 OpenAI 金鑰是否有效 |
