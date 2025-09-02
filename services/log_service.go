package services

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/sirupsen/logrus"
)

// LogService 日誌服務
type LogService struct {
	mutex      sync.RWMutex
	logs       []models.SystemLog
	maxLogSize int
	nextID     int64
}

var logService *LogService
var logServiceOnce sync.Once

// GetLogService 獲取日誌服務實例（單例模式）
func GetLogService() *LogService {
	logServiceOnce.Do(func() {
		logService = &LogService{
			logs:       make([]models.SystemLog, 0),
			maxLogSize: 10000, // 最多保存 10000 條日誌
			nextID:     1,
		}
	})
	return logService
}

// WriteLog 寫入日誌到內存存儲
func (ls *LogService) WriteLog(level, message string, data map[string]interface{}) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	// 創建新的日誌記錄
	logEntry := models.SystemLog{
		ID:        ls.nextID,
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	}

	// 添加到日誌列表
	ls.logs = append([]models.SystemLog{logEntry}, ls.logs...) // 新日誌添加到開頭

	// 如果超過最大長度，移除最舊的日誌
	if len(ls.logs) > ls.maxLogSize {
		ls.logs = ls.logs[:ls.maxLogSize]
	}

	ls.nextID++
}

// GetLogs 獲取日誌列表
func (ls *LogService) GetLogs(page, limit int, level string) ([]models.SystemLog, int) {
	ls.mutex.RLock()
	defer ls.mutex.RUnlock()

	// 根據級別篩選
	filteredLogs := ls.logs
	if level != "" && level != "all" {
		filteredLogs = make([]models.SystemLog, 0)
		for _, log := range ls.logs {
			if log.Level == level {
				filteredLogs = append(filteredLogs, log)
			}
		}
	}

	total := len(filteredLogs)

	// 分頁
	start := (page - 1) * limit
	if start >= total {
		return []models.SystemLog{}, total
	}

	end := start + limit
	if end > total {
		end = total
	}

	return filteredLogs[start:end], total
}

// GetLogStats 獲取日誌統計
func (ls *LogService) GetLogStats() map[string]int {
	ls.mutex.RLock()
	defer ls.mutex.RUnlock()

	stats := map[string]int{
		"total":   len(ls.logs),
		"debug":   0,
		"info":    0,
		"warning": 0,
		"error":   0,
	}

	for _, log := range ls.logs {
		if count, exists := stats[log.Level]; exists {
			stats[log.Level] = count + 1
		}
	}

	return stats
}

// ClearLogs 清空日誌
func (ls *LogService) ClearLogs() {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	ls.logs = make([]models.SystemLog, 0)
	ls.nextID = 1
}

// LogHook 實現 logrus.Hook 接口，用於攔截 logrus 日誌並寫入到服務
type LogHook struct {
	logService *LogService
}

// NewLogHook 創建新的日誌鉤子
func NewLogHook() *LogHook {
	return &LogHook{
		logService: GetLogService(),
	}
}

// Levels 返回需要攔截的日誌級別
func (hook *LogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 當日誌事件觸發時調用
func (hook *LogHook) Fire(entry *logrus.Entry) error {
	// 轉換 logrus.Entry 為我們的格式
	data := make(map[string]interface{})

	// 複製字段
	for key, value := range entry.Data {
		data[key] = value
	}

	// 轉換級別名稱
	levelName := entry.Level.String()
	if levelName == "warn" {
		levelName = "warning"
	}

	// 寫入到服務
	hook.logService.WriteLog(levelName, entry.Message, data)

	return nil
}

// LogStructuredData 記錄結構化數據
func LogStructuredData(level string, message string, data interface{}) {
	logService := GetLogService()

	// 將數據轉換為 map
	var dataMap map[string]interface{}
	if data != nil {
		// 嘗試直接轉換
		if m, ok := data.(map[string]interface{}); ok {
			dataMap = m
		} else {
			// 通過 JSON 序列化和反序列化來轉換
			jsonData, err := json.Marshal(data)
			if err == nil {
				json.Unmarshal(jsonData, &dataMap)
			}
		}
	}

	logService.WriteLog(level, message, dataMap)
}

// LogSystemEvent 記錄系統事件
func LogSystemEvent(event string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["event_type"] = "system_event"
	data["event_name"] = event

	LogStructuredData("info", "System event: "+event, data)
}

// LogAPIEvent 記錄 API 事件
func LogAPIEvent(method, path string, statusCode int, duration int64, userID string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}

	data["event_type"] = "api_request"
	data["method"] = method
	data["path"] = path
	data["status_code"] = statusCode
	data["duration_ms"] = duration

	if userID != "" {
		data["user_id"] = userID
	}

	level := "info"
	if statusCode >= 400 {
		level = "warning"
	}
	if statusCode >= 500 {
		level = "error"
	}

	message := "API request: " + method + " " + path
	LogStructuredData(level, message, data)
}

// LogChatEvent 記錄聊天事件
func LogChatEvent(chatID, userID, characterID string, success bool, duration int64, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}

	data["event_type"] = "chat_message"
	data["chat_id"] = chatID
	data["user_id"] = userID
	data["character_id"] = characterID
	data["success"] = success
	data["duration_ms"] = duration

	level := "info"
	message := "Chat message processed"

	if !success {
		level = "error"
		message = "Chat message failed"
	}

	LogStructuredData(level, message, data)
}

// LogError 記錄錯誤
func LogError(err error, context string, data map[string]interface{}) {
	if data == nil {
		data = make(map[string]interface{})
	}

	data["event_type"] = "error"
	data["error_message"] = err.Error()
	data["context"] = context

	message := "Error occurred: " + context
	LogStructuredData("error", message, data)
}
