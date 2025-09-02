package models

import (
	"time"

	"github.com/clarencetw/thewavess-ai-core/models/db"
)

// Chat 聊天會話領域模型
type Chat struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	CharacterID     string     `json:"character_id"`
	Title           string     `json:"title,omitempty"`
	Status          string     `json:"status"`
	ChatMode        string     `json:"chat_mode"`
	MessageCount    int        `json:"message_count"`
	TotalCharacters int        `json:"total_characters"`
	LastMessageAt   *time.Time `json:"last_message_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// 關聯
	User      *User      `json:"user,omitempty"`
	Character *Character `json:"character,omitempty"`
	Messages  []*Message `json:"messages,omitempty"`
}

// Message 消息領域模型
type Message struct {
	ID                 string                 `json:"id"`
	ChatID             string                 `json:"chat_id"`
	Role               string                 `json:"role"`
	Dialogue           string                 `json:"dialogue"`
	SceneDescription   *string                `json:"scene_description,omitempty"`
	Action             *string                `json:"action,omitempty"`
	EmotionalState     map[string]interface{} `json:"emotional_state,omitempty"`
	AIEngine           string                 `json:"ai_engine"`
	ResponseTimeMs     int                    `json:"response_time_ms"`
	NSFWLevel          int                    `json:"nsfw_level"`
	IsRegenerated      bool                   `json:"is_regenerated"`
	RegenerationReason *string                `json:"regeneration_reason,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`

	// 關聯
	Chat *Chat `json:"chat,omitempty"`
}

// ChatFromDB 從資料庫模型轉換為聊天會話領域模型
func ChatFromDB(chatDB *db.ChatDB) *Chat {
	if chatDB == nil {
		return nil
	}

	return &Chat{
		ID:              chatDB.ID,
		UserID:          chatDB.UserID,
		CharacterID:     chatDB.CharacterID,
		Title:           chatDB.Title,
		Status:          chatDB.Status,
		ChatMode:        chatDB.ChatMode,
		MessageCount:    chatDB.MessageCount,
		TotalCharacters: chatDB.TotalCharacters,
		LastMessageAt:   chatDB.LastMessageAt,
		CreatedAt:       chatDB.CreatedAt,
		UpdatedAt:       chatDB.UpdatedAt,
	}
}

// MessageFromDB 從資料庫模型轉換為消息領域模型
func MessageFromDB(messageDB *db.MessageDB) *Message {
	if messageDB == nil {
		return nil
	}

	return &Message{
		ID:                 messageDB.ID,
		ChatID:             messageDB.ChatID,
		Role:               messageDB.Role,
		Dialogue:           messageDB.Dialogue,
		SceneDescription:   messageDB.SceneDescription,
		Action:             messageDB.Action,
		EmotionalState:     messageDB.EmotionalState,
		AIEngine:           messageDB.AIEngine,
		ResponseTimeMs:     messageDB.ResponseTimeMs,
		NSFWLevel:          messageDB.NSFWLevel,
		IsRegenerated:      messageDB.IsRegenerated,
		RegenerationReason: messageDB.RegenerationReason,
		CreatedAt:          messageDB.CreatedAt,
	}
}

// ChatResponse 會話響應格式
type ChatResponse struct {
	ID              string             `json:"id"`
	UserID          string             `json:"user_id"`
	CharacterID     string             `json:"character_id"`
	Title           string             `json:"title,omitempty"`
	Status          string             `json:"status"`
	ChatMode        string             `json:"chat_mode,omitempty"`
	MessageCount    int                `json:"message_count"`
	TotalCharacters int                `json:"total_characters"`
	LastMessageAt   *time.Time         `json:"last_message_at,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	Character       *CharacterResponse `json:"character,omitempty"`
	LastMessage     *MessageResponse   `json:"last_message,omitempty"`
}

// MessageResponse 消息響應格式
type MessageResponse struct {
	ID                 string                 `json:"id"`
	ChatID             string                 `json:"chat_id"`
	Role               string                 `json:"role"`
	Dialogue           string                 `json:"dialogue"`
	SceneDescription   string                 `json:"scene_description,omitempty"`
	Action             string                 `json:"action,omitempty"`
	EmotionalState     map[string]interface{} `json:"emotional_state,omitempty"`
	AIEngine           string                 `json:"ai_engine"`
	ResponseTimeMs     int                    `json:"response_time_ms"`
	NSFWLevel          int                    `json:"nsfw_level"`
	IsRegenerated      bool                   `json:"is_regenerated"`
	RegenerationReason string                 `json:"regeneration_reason,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
}

// ToResponse 轉換為響應格式
func (c *Chat) ToResponse() *ChatResponse {
	response := &ChatResponse{
		ID:              c.ID,
		UserID:          c.UserID,
		CharacterID:     c.CharacterID,
		Title:           c.Title,
		Status:          c.Status,
		ChatMode:        c.ChatMode,
		MessageCount:    c.MessageCount,
		TotalCharacters: c.TotalCharacters,
		LastMessageAt:   c.LastMessageAt,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}

	// 添加角色信息
	if c.Character != nil {
		response.Character = c.Character.ToResponse()
	}

	// 添加最後一條消息
	if len(c.Messages) > 0 {
		lastMessage := c.Messages[len(c.Messages)-1]
		response.LastMessage = lastMessage.ToResponse()
	}

	return response
}

func (m *Message) ToResponse() *MessageResponse {
	response := &MessageResponse{
		ID:             m.ID,
		ChatID:         m.ChatID,
		Role:           m.Role,
		Dialogue:       m.Dialogue,
		EmotionalState: m.EmotionalState,
		NSFWLevel:      m.NSFWLevel,
		IsRegenerated:  m.IsRegenerated,
		CreatedAt:      m.CreatedAt,
	}

	// 處理指標欄位
	if m.SceneDescription != nil {
		response.SceneDescription = *m.SceneDescription
	}
	if m.Action != nil {
		response.Action = *m.Action
	}
	response.AIEngine = m.AIEngine
	response.ResponseTimeMs = m.ResponseTimeMs
	if m.RegenerationReason != nil {
		response.RegenerationReason = *m.RegenerationReason
	}

	return response
}

// CreateChatRequest 創建會話請求
type CreateChatRequest struct {
	CharacterID string `json:"character_id" binding:"required"`
	Title       string `json:"title,omitempty"`
}

// SendMessageRequest 發送消息請求
type SendMessageRequest struct {
	Message string `json:"message" binding:"required,max=2000"`
}

// 聊天相關響應結構
type ChatListResponse struct {
	Chats      []*ChatResponse    `json:"chats"`
	Pagination PaginationResponse `json:"pagination"`
}

type MessageHistoryResponse struct {
	ChatID     string             `json:"chat_id"`
	Messages   []*MessageResponse `json:"messages"`
	Pagination PaginationResponse `json:"pagination"`
}
