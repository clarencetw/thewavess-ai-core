package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// GinStyleFormatter 自定義的 Gin 風格格式化器
type GinStyleFormatter struct{}

// Format 實現 logrus.Formatter 接口
func (f *GinStyleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006/01/02 - 15:04:05")
	level := strings.ToUpper(entry.Level.String())
	
	// 根據級別添加顏色（僅在 TTY 環境下）
	var levelColor string
	switch entry.Level {
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = "\033[31m" // 紅色
	case logrus.WarnLevel:
		levelColor = "\033[33m" // 黃色
	case logrus.InfoLevel:
		levelColor = "\033[32m" // 綠色
	case logrus.DebugLevel:
		levelColor = "\033[36m" // 青色
	default:
		levelColor = "\033[37m" // 白色
	}
	resetColor := "\033[0m"
	
	// 檢查是否在 TTY 環境
	if !isTerminal() {
		levelColor = ""
		resetColor = ""
	}
	
	var output strings.Builder
	
	// 檢查是否為數據庫查詢日誌（需要特殊格式）
	if _, hasOp := entry.Data["operation"]; hasOp {
		duration, hasDuration := entry.Data["duration"]
		
		// 第一行：簡潔的概覽信息
		output.WriteString(fmt.Sprintf("[CORE] %s | %s%s%s | %s", 
			timestamp, levelColor, level, resetColor, entry.Message))
		
		if hasDuration {
			output.WriteString(fmt.Sprintf(" | %v", duration))
		}
		output.WriteString("\n")
		
		// 第二行：完整的結構化數據
		output.WriteString("  ↪ {")
		
		// 按重要性排序字段
		orderedFields := []string{"operation", "duration", "query", "error"}
		first := true
		
		for _, key := range orderedFields {
			if value, exists := entry.Data[key]; exists {
				if !first {
					output.WriteString(",")
				}
				output.WriteString(fmt.Sprintf("\"%s\":\"%v\"", key, value))
				first = false
			}
		}
		
		// 添加其他字段
		for key, value := range entry.Data {
			isOrdered := false
			for _, orderedKey := range orderedFields {
				if key == orderedKey {
					isOrdered = true
					break
				}
			}
			if !isOrdered {
				if !first {
					output.WriteString(",")
				}
				output.WriteString(fmt.Sprintf("\"%s\":\"%v\"", key, value))
				first = false
			}
		}
		
		output.WriteString("}\n")
	} else {
		// 普通日誌：使用簡潔格式
		output.WriteString(fmt.Sprintf("[CORE] %s | %s%s%s | %s", 
			timestamp, levelColor, level, resetColor, entry.Message))
		
		// 簡要添加重要字段
		importantFields := []string{"service", "version", "user_id", "session_id", "error"}
		hasImportant := false
		
		for _, key := range importantFields {
			if value, exists := entry.Data[key]; exists {
				if !hasImportant {
					output.WriteString(" |")
					hasImportant = true
				}
				output.WriteString(fmt.Sprintf(" %s=%v", key, value))
			}
		}
		
		// 如果有其他字段，在下一行顯示
		otherFields := make(map[string]interface{})
		for key, value := range entry.Data {
			isImportant := false
			for _, importantKey := range importantFields {
				if key == importantKey {
					isImportant = true
					break
				}
			}
			if !isImportant {
				otherFields[key] = value
			}
		}
		
		if len(otherFields) > 0 {
			output.WriteString("\n  ↪ {")
			first := true
			for key, value := range otherFields {
				if !first {
					output.WriteString(",")
				}
				output.WriteString(fmt.Sprintf("\"%s\":\"%v\"", key, value))
				first = false
			}
			output.WriteString("}")
		}
		
		output.WriteString("\n")
	}
	
	return []byte(output.String()), nil
}

// isTerminal 檢查是否在終端環境
func isTerminal() bool {
	return GetEnvWithDefault("TERM", "") != ""
}

// InitLogger 初始化日誌記錄器
func InitLogger() {
	Logger = logrus.New()
	
	// 設置輸出格式為 Gin 風格
	Logger.SetFormatter(&GinStyleFormatter{})
	
	// 設置輸出到標準輸出
	Logger.SetOutput(os.Stdout)
	
	// 根據環境變數設置日誌等級
	level := GetEnvWithDefault("LOG_LEVEL", "info")
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