package models

import (
	"github.com/clarencetw/thewavess-ai-core/models/db"
)

// ChatMapper 聊天模型轉換器
type ChatMapper struct{}

// NewChatMapper 創建新的聊天映射器
func NewChatMapper() *ChatMapper {
	return &ChatMapper{}
}

// ChatSessionFromDB 從資料庫模型轉換為聊天會話領域模型
func (m *ChatMapper) ChatSessionFromDB(sessionDB *db.ChatSessionDB) *ChatSession {
	if sessionDB == nil {
		return nil
	}

	return &ChatSession{
		ID:              sessionDB.ID,
		UserID:          sessionDB.UserID,
		CharacterID:     sessionDB.CharacterID,
		Title:           sessionDB.Title,
		Status:          sessionDB.Status,
		MessageCount:    sessionDB.MessageCount,
		TotalCharacters: sessionDB.TotalCharacters,
		LastMessageAt:   sessionDB.LastMessageAt,
		CreatedAt:       sessionDB.CreatedAt,
		UpdatedAt:       sessionDB.UpdatedAt,
	}
}

// ChatSessionToDB 從聊天會話領域模型轉換為資料庫模型
func (m *ChatMapper) ChatSessionToDB(session *ChatSession) *db.ChatSessionDB {
	if session == nil {
		return nil
	}

	return &db.ChatSessionDB{
		ID:              session.ID,
		UserID:          session.UserID,
		CharacterID:     session.CharacterID,
		Title:           session.Title,
		Status:          session.Status,
		MessageCount:    session.MessageCount,
		TotalCharacters: session.TotalCharacters,
		LastMessageAt:   session.LastMessageAt,
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
	}
}

// MessageFromDB 從資料庫模型轉換為消息領域模型
func (m *ChatMapper) MessageFromDB(messageDB *db.MessageDB) *Message {
	if messageDB == nil {
		return nil
	}

	return &Message{
		ID:                  messageDB.ID,
		SessionID:           messageDB.SessionID,
		Role:                messageDB.Role,
		Content:             messageDB.Content,
		SceneDescription:    messageDB.SceneDescription,
		CharacterAction:     messageDB.CharacterAction,
		EmotionalState:      messageDB.EmotionalState,
		AIEngine:            messageDB.AIEngine,
		ResponseTimeMs:      messageDB.ResponseTimeMs,
		NSFWLevel:           messageDB.NSFWLevel,
		IsRegenerated:       messageDB.IsRegenerated,
		RegenerationReason:  messageDB.RegenerationReason,
		CreatedAt:           messageDB.CreatedAt,
	}
}

// MessageToDB 從消息領域模型轉換為資料庫模型
func (m *ChatMapper) MessageToDB(message *Message) *db.MessageDB {
	if message == nil {
		return nil
	}

	return &db.MessageDB{
		ID:                  message.ID,
		SessionID:           message.SessionID,
		Role:                message.Role,
		Content:             message.Content,
		SceneDescription:    message.SceneDescription,
		CharacterAction:     message.CharacterAction,
		EmotionalState:      message.EmotionalState,
		AIEngine:            message.AIEngine,
		ResponseTimeMs:      message.ResponseTimeMs,
		NSFWLevel:           message.NSFWLevel,
		IsRegenerated:       message.IsRegenerated,
		RegenerationReason:  message.RegenerationReason,
		CreatedAt:           message.CreatedAt,
	}
}

// ChatSessionFromDBList 從資料庫模型列表轉換為聊天會話領域模型列表
func (m *ChatMapper) ChatSessionFromDBList(sessionDBs []*db.ChatSessionDB) []*ChatSession {
	if sessionDBs == nil {
		return nil
	}

	sessions := make([]*ChatSession, len(sessionDBs))
	for i, sessionDB := range sessionDBs {
		sessions[i] = m.ChatSessionFromDB(sessionDB)
	}
	return sessions
}

// ChatSessionToDBList 從聊天會話領域模型列表轉換為資料庫模型列表
func (m *ChatMapper) ChatSessionToDBList(sessions []*ChatSession) []*db.ChatSessionDB {
	if sessions == nil {
		return nil
	}

	sessionDBs := make([]*db.ChatSessionDB, len(sessions))
	for i, session := range sessions {
		sessionDBs[i] = m.ChatSessionToDB(session)
	}
	return sessionDBs
}

// MessageFromDBList 從資料庫模型列表轉換為消息領域模型列表
func (m *ChatMapper) MessageFromDBList(messageDBs []*db.MessageDB) []*Message {
	if messageDBs == nil {
		return nil
	}

	messages := make([]*Message, len(messageDBs))
	for i, messageDB := range messageDBs {
		messages[i] = m.MessageFromDB(messageDB)
	}
	return messages
}

// MessageToDBList 從消息領域模型列表轉換為資料庫模型列表
func (m *ChatMapper) MessageToDBList(messages []*Message) []*db.MessageDB {
	if messages == nil {
		return nil
	}

	messageDBs := make([]*db.MessageDB, len(messages))
	for i, message := range messages {
		messageDBs[i] = m.MessageToDB(message)
	}
	return messageDBs
}