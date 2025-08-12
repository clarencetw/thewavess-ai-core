package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// InitLogger 初始化日誌記錄器
func InitLogger() {
	Logger = logrus.New()
	
	// 設置輸出格式
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	
	// 設置輸出到標準輸出
	Logger.SetOutput(os.Stdout)
	
	// 根據環境變數設置日誌等級
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		Logger.SetLevel(logrus.WarnLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}
	
	Logger.WithFields(logrus.Fields{
		"service": "thewavess-ai-core",
		"version": "1.0.0",
	}).Info("Logger initialized successfully")
}

// GetLoggerWithFields 獲取帶有預設字段的 logger
func GetLoggerWithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

// LogAPIRequest 記錄 API 請求
func LogAPIRequest(method, path, userID, sessionID string, statusCode int, duration int64) {
	Logger.WithFields(logrus.Fields{
		"type":        "api_request",
		"method":      method,
		"path":        path,
		"user_id":     userID,
		"session_id":  sessionID,
		"status_code": statusCode,
		"duration_ms": duration,
	}).Info("API request processed")
}

// LogChatMessage 記錄對話消息
func LogChatMessage(sessionID, userID, characterID, engine string, responseTime int64, success bool) {
	entry := Logger.WithFields(logrus.Fields{
		"type":          "chat_message",
		"session_id":    sessionID,
		"user_id":       userID,
		"character_id":  characterID,
		"ai_engine":     engine,
		"response_time": responseTime,
		"success":       success,
	})
	
	if success {
		entry.Info("Chat message processed successfully")
	} else {
		entry.Error("Chat message processing failed")
	}
}

// LogError 記錄錯誤
func LogError(err error, context string, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["error"] = err.Error()
	fields["context"] = context
	
	Logger.WithFields(fields).Error("Error occurred")
}

// LogServiceEvent 記錄服務事件
func LogServiceEvent(event string, fields logrus.Fields) {
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["event"] = event
	
	Logger.WithFields(fields).Info("Service event")
}