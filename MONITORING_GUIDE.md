# 系統監控實作指南

本文件說明系統的監控機制、健康檢查策略、性能指標追蹤與故障排除方法，專注於提供實用的運維指導。

## 目標

- 提供簡單有效的系統健康狀態檢查
- 建立完整的性能監控和指標收集機制
- 支援容器化部署的健康檢查需求
- 實現快速故障定位和問題排除

## 監控架構

### 監控層級
- **基礎健康檢查**：服務是否正常運行
- **詳細系統狀態**：性能指標、資源使用情況
- **服務依賴檢查**：資料庫、AI API、外部服務狀態
- **容器健康探針**：Kubernetes/Docker 環境支援

### 端點分類
- **簡單檢查**：`/health` - 快速健康確認
- **詳細監控**：`/monitor/*` - 專業監控工具
- **容器探針**：`/monitor/ready`, `/monitor/live`
- **指標收集**：`/monitor/metrics` - Prometheus 格式

## 實作策略

### 健康檢查邏輯
```
1. 資料庫連接檢查 → 核心功能可用性
2. AI 服務配置檢查 → 功能完整性
3. 系統資源檢查 → 性能指標
4. 整體狀態評估 → 服務等級
```

### 狀態分級
- **healthy**: 所有服務正常
- **degraded**: 部分服務不可用但核心功能正常
- **unhealthy**: 核心服務故障

### 關鍵指標追蹤
```
系統指標:
- 記憶體使用量 (目標: <100MB)
- Goroutine 數量 (目標: <50)
- GC 頻率 (目標: 適中)
- 運行時間 (uptime)

服務指標:
- 資料庫延遲 (目標: <5ms)
- API 配置狀態
- 錯誤率 (目標: <1%)
```

## 使用指南

### 日常監控
```bash
# 快速健康檢查
curl http://localhost:8080/health

# 詳細系統狀態
curl http://localhost:8080/api/v1/monitor/stats | jq .

# 查看資源使用
curl http://localhost:8080/api/v1/monitor/metrics | grep memory
```

### Docker 環境
```yaml
# docker-compose.yml 健康檢查配置
healthcheck:
  test: ["CMD", "wget", "--spider", "http://localhost:8080/api/v1/monitor/health"]
  interval: 30s
  timeout: 10s
  retries: 3
```

### 容器部署檢查
```bash
# 檢查容器健康狀態
docker ps --format "table {{.Names}}\t{{.Status}}"

# 查看健康檢查日誌
docker inspect --format='{{.State.Health}}' container_name
```

## 故障排除流程

### 常見問題診斷

#### 1. 資料庫連接失敗
```
症狀: status="degraded", database="unhealthy"
檢查:
- 資料庫容器是否運行
- 連接字串是否正確
- 網路連通性
- 資料庫用戶權限
```

#### 2. 記憶體使用過高
```
症狀: memory_usage > 200MB
檢查:
- Goroutine 洩漏
- 未釋放的資源
- GC 配置
- 請求處理邏輯
```

#### 3. AI 服務配置問題
```
症狀: openai="not_configured", grok="not_configured"
檢查:
- API Key 環境變數
- API 端點配置
- 網路連接
- 配額限制
```

### 監控最佳實踐

#### 1. 定期檢查
- 每 30 秒執行健康檢查
- 每 5 分鐘收集性能指標
- 異常狀態立即告警

#### 2. 閾值設定
```
記憶體使用: 正常 <50MB, 警告 50-100MB, 危險 >100MB
資料庫延遲: 正常 <2ms, 警告 2-10ms, 危險 >10ms
Goroutines: 正常 <20, 警告 20-50, 危險 >50
```

#### 3. 日誌分析
```bash
# 查看健康檢查日誌
docker logs api_container | grep "健康檢查"

# 監控錯誤日誌
docker logs api_container | grep "ERROR"
```

## 開發指導

### 新增監控指標
1. 在 `handlers/monitor.go` 添加新指標
2. 更新 Prometheus 格式輸出
3. 調整健康檢查邏輯
4. 更新文檔說明

### 自定義健康檢查
```go
// 添加新的服務檢查
func checkCustomService() string {
    // 實作檢查邏輯
    return "healthy" // or "unhealthy"
}
```

### 性能優化
- 監控端點應保持輕量級
- 避免昂貴的計算操作
- 使用緩存減少重複檢查
- 異步處理非關鍵指標

## 生產環境建議

### 監控工具整合
- Prometheus + Grafana 監控面板
- ELK Stack 日誌分析
- 告警系統配置
- 自動故障恢復

### 維護策略
- 定期清理日誌檔案
- 監控磁碟空間使用
- 資料庫性能調優
- 備份監控數據