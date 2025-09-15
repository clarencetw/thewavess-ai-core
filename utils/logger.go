package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

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
		importantFields := []string{"service", "version", "user_id", "chat_id", "error"}
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
func LogAPIRequest(method, path, userID string, statusCode int, duration time.Duration, err error) {
	entry := Logger.WithFields(logrus.Fields{
		"event_type":  "api_request",
		"http_method": method,
		"http_path":   path,
		"user_id":     userID,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
		"success":     err == nil,
	})

	if err != nil {
		entry.WithError(err).Error("API request failed")
	} else {
		entry.Info("API request completed")
	}
}

// LogChatMessage 記錄聊天消息
func LogChatMessage(chatID, userID, characterID, aiEngine string, responseTime time.Duration, success bool, err error) {
	entry := Logger.WithFields(logrus.Fields{
		"event_type":    "chat_message",
		"chat_id":       chatID,
		"user_id":       userID,
		"character_id":  characterID,
		"ai_engine":     aiEngine,
		"response_time": responseTime.Milliseconds(),
		"success":       success,
	})

	if err != nil {
		entry.WithError(err).Error("Chat message processing failed")
	} else {
		entry.Info("Chat message processed successfully")
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

// LogSecurityEvent 記錄安全相關事件
func LogSecurityEvent(eventType, userID, description string, severity string, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"event_type":     "security",
		"security_event": eventType,
		"user_id":        userID,
		"severity":       severity,
		"timestamp":      time.Now().UTC(),
	}

	// 合併額外元數據
	for k, v := range metadata {
		fields[k] = v
	}

	entry := Logger.WithFields(fields)

	switch severity {
	case "critical":
		entry.Error("Critical security event: " + description)
	case "high":
		entry.Warn("High severity security event: " + description)
	case "medium":
		entry.Warn("Medium severity security event: " + description)
	default:
		entry.Info("Security event: " + description)
	}
}

// LogPerformanceMetric 記錄性能指標
func LogPerformanceMetric(operation string, duration time.Duration, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"event_type":  "performance",
		"operation":   operation,
		"duration_ms": duration.Milliseconds(),
		"duration_ns": duration.Nanoseconds(),
	}

	for k, v := range metadata {
		fields[k] = v
	}

	entry := Logger.WithFields(fields)

	// 根據執行時間判斷日誌級別
	switch {
	case duration > 10*time.Second:
		entry.Error("Very slow operation detected")
	case duration > 5*time.Second:
		entry.Warn("Slow operation detected")
	case duration > 1*time.Second:
		entry.Info("Performance metric recorded")
	default:
		entry.Debug("Performance metric recorded")
	}
}

// LogBusinessEvent 記錄業務事件
func LogBusinessEvent(eventType, description string, metadata map[string]interface{}) {
	fields := logrus.Fields{
		"event_type":     "business",
		"business_event": eventType,
		"timestamp":      time.Now().UTC(),
	}

	for k, v := range metadata {
		fields[k] = v
	}

	Logger.WithFields(fields).Info(description)
}

// LogAdminAction 記錄管理員操作
func LogAdminAction(adminID, adminEmail, action, target, clientIP, userAgent string, success bool, details string) {
	fields := logrus.Fields{
		"event_type":  "admin_action",
		"admin_id":    adminID,
		"admin_email": adminEmail,
		"action":      action,
		"target":      target,
		"client_ip":   clientIP,
		"user_agent":  userAgent,
		"success":     success,
		"details":     details,
		"audit_event": true,
		"timestamp":   Now(),
	}

	if success {
		Logger.WithFields(fields).Info("管理員操作成功")
	} else {
		Logger.WithFields(fields).Warn("管理員操作失敗")
	}
}

// LogUserAuthEvent 記錄用戶認證事件
func LogUserAuthEvent(eventType, username, email, clientIP, userAgent string, success bool, reason string) {
	fields := logrus.Fields{
		"event_type": eventType,
		"username":   username,
		"email":      email,
		"client_ip":  clientIP,
		"user_agent": userAgent,
		"success":    success,
		"auth_event": true,
		"timestamp":  Now(),
	}

	if reason != "" {
		fields["reason"] = reason
	}

	if success {
		Logger.WithFields(fields).Info("用戶認證成功")
	} else {
		Logger.WithFields(fields).Warn("用戶認證失敗")
	}
}

// ===== 向後兼容的包裝函數 =====

// WithContext 創建帶上下文的日誌條目
func WithContext(ctx map[string]interface{}) *logrus.Entry {
	return Logger.WithFields(logrus.Fields(ctx))
}

// WithError 創建帶錯誤的日誌條目
func WithError(err error) *logrus.Entry {
	return Logger.WithError(err)
}

// WithUser 創建帶用戶信息的日誌條目
func WithUser(userID, username string) *logrus.Entry {
	return Logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"username": username,
	})
}

// WithRequest 創建帶請求信息的日誌條目
func WithRequest(method, path, userAgent, clientIP string) *logrus.Entry {
	return Logger.WithFields(logrus.Fields{
		"http_method": method,
		"http_path":   path,
		"user_agent":  userAgent,
		"client_ip":   clientIP,
	})
}

// WithChat 創建帶聊天信息的日誌條目
func WithChat(chatID, characterID string) *logrus.Entry {
	return Logger.WithFields(logrus.Fields{
		"chat_id":      chatID,
		"character_id": characterID,
	})
}
