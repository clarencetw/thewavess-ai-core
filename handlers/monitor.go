package handlers

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// HealthResponse 健康檢查回應結構
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Services  map[string]string `json:"services"`
	Message   string            `json:"message,omitempty"`
}

// SystemStatsResponse 系統狀態回應結構
type SystemStatsResponse struct {
	Status    string         `json:"status"`
	Timestamp time.Time      `json:"timestamp"`
	System    SystemInfo     `json:"system"`
	Database  DatabaseInfo   `json:"database"`
	Runtime   RuntimeInfo    `json:"runtime"`
	Services  ServicesStatus `json:"services"`
}

// SystemInfo 系統資訊
type SystemInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	NumCPU       int    `json:"num_cpu"`
	GoVersion    string `json:"go_version"`
}

// DatabaseInfo 資料庫資訊
type DatabaseInfo struct {
	Status      string `json:"status"`
	Type        string `json:"type"`
	Connected   bool   `json:"connected"`
	PingLatency string `json:"ping_latency,omitempty"`
}

// RuntimeInfo 運行時資訊
type RuntimeInfo struct {
	Goroutines  int    `json:"goroutines"`
	MemoryUsage string `json:"memory_usage"`
	GCCount     uint32 `json:"gc_count"`
	NextGC      string `json:"next_gc"`
	LastGC      string `json:"last_gc"`
}

// ServicesStatus 服務狀態
type ServicesStatus struct {
	OpenAI string `json:"openai"`
	Grok   string `json:"grok"`
	TTS    string `json:"tts"`
}

var MonitorStartTime = time.Now() // 導出供其他包使用

// HealthCheck godoc
// @Summary      系統健康檢查
// @Description  檢查系統整體健康狀態，包括服務運行時間、資料庫連接狀態等基本信息
// @Tags         Monitor
// @Accept       json
// @Produce      json
// @Success      200 {object} models.APIResponse{data=HealthResponse} "系統健康"
// @Success      503 {object} models.APIResponse{data=HealthResponse} "系統服務降級"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "內部服務器錯誤"
// @Router       /monitor/health [get]
func HealthCheck(c *gin.Context) {
	utils.Logger.Info("執行健康檢查")

	// 檢查資料庫連接
	dbStatus := "healthy"
	services := map[string]string{
		"database": "healthy",
		"api":      "healthy",
	}

	// 測試資料庫連接
	if GetDB() != nil {
		ctx := c.Request.Context()
		if err := GetDB().PingContext(ctx); err != nil {
			utils.Logger.WithError(err).Error("資料庫連接檢查失敗")
			dbStatus = "unhealthy"
			services["database"] = "unhealthy"
		}
	} else {
		dbStatus = "unhealthy"
		services["database"] = "disconnected"
	}

	// 決定整體狀態
	status := "healthy"
	if dbStatus == "unhealthy" {
		status = "degraded"
	}

	uptime := time.Since(MonitorStartTime).Round(time.Second).String()

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   utils.GetEnvWithDefault("APP_VERSION", "1.0.0"),
		Uptime:    uptime,
		Services:  services,
	}

	// 根據狀態返回適當的HTTP狀態碼
	httpStatus := http.StatusOK
	if status == "degraded" {
		httpStatus = http.StatusServiceUnavailable
		response.Message = "某些服務不可用"
	}

	utils.Logger.WithFields(logrus.Fields{
		"status":    status,
		"uptime":    uptime,
		"db_status": dbStatus,
	}).Info("健康檢查完成")

	c.JSON(httpStatus, gin.H{
		"success": status == "healthy",
		"data":    response,
	})
}

// GetSystemStats godoc
// @Summary      獲取系統詳細狀態
// @Description  獲取系統詳細運行狀態，包括硬體信息、記憶體使用、資料庫狀態、外部服務配置等完整監控信息
// @Tags         Monitor
// @Accept       json
// @Produce      json
// @Success      200 {object} models.APIResponse{data=SystemStatsResponse} "獲取成功"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "獲取系統狀態失敗"
// @Router       /monitor/stats [get]
func GetSystemStats(c *gin.Context) {
	utils.Logger.Info("獲取系統狀態統計")

	// 獲取運行時統計
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 檢查資料庫狀態
	dbInfo := DatabaseInfo{
		Type:      "postgresql",
		Connected: false,
		Status:    "disconnected",
	}

	if GetDB() != nil {
		ctx := c.Request.Context()
		start := time.Now()
		if err := GetDB().PingContext(ctx); err == nil {
			dbInfo.Connected = true
			dbInfo.Status = "connected"
			dbInfo.PingLatency = time.Since(start).Round(time.Microsecond).String()
		} else {
			dbInfo.Status = "error"
			utils.Logger.WithError(err).Error("資料庫ping失敗")
		}
	}

	// 檢查外部服務狀態
	services := ServicesStatus{
		OpenAI: getServiceStatus("OPENAI_API_KEY"),
		Grok:   getServiceStatus("GROK_API_KEY"),
		TTS:    "configured", // TTS通常是內建或基於其他服務
	}

	response := SystemStatsResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		System: SystemInfo{
			OS:           runtime.GOOS,
			Architecture: runtime.GOARCH,
			NumCPU:       runtime.NumCPU(),
			GoVersion:    runtime.Version(),
		},
		Database: dbInfo,
		Runtime: RuntimeInfo{
			Goroutines:  runtime.NumGoroutine(),
			MemoryUsage: formatBytes(memStats.Alloc),
			GCCount:     memStats.NumGC,
			NextGC:      formatBytes(memStats.NextGC),
			LastGC:      time.Unix(0, int64(memStats.LastGC)).Format(time.RFC3339),
		},
		Services: services,
	}

	utils.Logger.WithFields(logrus.Fields{
		"goroutines":   response.Runtime.Goroutines,
		"memory_usage": response.Runtime.MemoryUsage,
		"db_connected": dbInfo.Connected,
	}).Info("系統狀態統計獲取完成")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetMetrics godoc
// @Summary      獲取Prometheus監控指標
// @Description  以Prometheus格式輸出系統監控指標，包括運行時間、記憶體使用量、Goroutine數量、GC統計等
// @Tags         Monitor
// @Accept       json
// @Produce      text/plain
// @Success      200 {string} string "Prometheus格式的監控指標"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "獲取指標失敗"
// @Router       /monitor/metrics [get]
func GetMetrics(c *gin.Context) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	uptime := time.Since(MonitorStartTime).Seconds()

	// 簡單的Prometheus格式指標
	metrics := `# HELP thewavess_uptime_seconds 系統運行時間（秒）
# TYPE thewavess_uptime_seconds counter
thewavess_uptime_seconds ` + formatFloat(uptime) + `

# HELP thewavess_memory_usage_bytes 記憶體使用量（字節）
# TYPE thewavess_memory_usage_bytes gauge
thewavess_memory_usage_bytes ` + formatUint64(memStats.Alloc) + `

# HELP thewavess_goroutines_total 當前Goroutine數量
# TYPE thewavess_goroutines_total gauge
thewavess_goroutines_total ` + formatInt(runtime.NumGoroutine()) + `

# HELP thewavess_gc_total GC執行總次數
# TYPE thewavess_gc_total counter
thewavess_gc_total ` + formatUint32(memStats.NumGC) + `
`

	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.String(http.StatusOK, metrics)
}

// Ready godoc
// @Summary      服務就緒檢查
// @Description  檢查服務是否已就緒可以接受流量，驗證關鍵依賴項（資料庫、AI引擎配置）是否正常，適用於Kubernetes readiness probe
// @Tags         Monitor
// @Accept       json
// @Produce      json
// @Success      200 {object} object{ready=bool,services=object,timestamp=string} "服務已就緒"
// @Success      503 {object} object{ready=bool,services=object,timestamp=string} "服務未就緒"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "就緒檢查失敗"
// @Router       /monitor/ready [get]
func Ready(c *gin.Context) {
	// 檢查關鍵服務是否就緒
	ready := true
	services := make(map[string]bool)

	// 檢查資料庫
	if GetDB() != nil {
		ctx := c.Request.Context()
		if err := GetDB().PingContext(ctx); err != nil {
			ready = false
			services["database"] = false
		} else {
			services["database"] = true
		}
	} else {
		ready = false
		services["database"] = false
	}

	// 檢查必要的環境變數
	if utils.GetEnvWithDefault("OPENAI_API_KEY", "") == "" &&
		utils.GetEnvWithDefault("GROK_API_KEY", "") == "" {
		ready = false
		services["ai_engines"] = false
	} else {
		services["ai_engines"] = true
	}

	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"ready":     ready,
		"services":  services,
		"timestamp": time.Now(),
	})
}

// Live godoc
// @Summary      服務存活檢查
// @Description  檢查服務是否仍在運行，用於確認應用程式進程存活狀態，適用於Kubernetes liveness probe
// @Tags         Monitor
// @Accept       json
// @Produce      json
// @Success      200 {object} object{alive=bool,timestamp=string,uptime=string} "服務存活"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "存活檢查失敗"
// @Router       /monitor/live [get]
func Live(c *gin.Context) {
	// 簡單的存活檢查
	c.JSON(http.StatusOK, gin.H{
		"alive":     true,
		"timestamp": time.Now(),
		"uptime":    time.Since(MonitorStartTime).Round(time.Second).String(),
	})
}

// 輔助函數

func getServiceStatus(envKey string) string {
	if utils.GetEnvWithDefault(envKey, "") != "" {
		return "configured"
	}
	return "not_configured"
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return formatUint64(bytes) + " B"
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return formatFloat(float64(bytes)/float64(div)) + " " + "KMGTPE"[exp:exp+1] + "B"
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

func formatUint64(n uint64) string {
	return fmt.Sprintf("%d", n)
}

func formatUint32(n uint32) string {
	return fmt.Sprintf("%d", n)
}

func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
}
