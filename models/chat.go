package models

import (
	"time"
)

// ChatSession 聊天會話領域模型
type ChatSession struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	CharacterID     string     `json:"character_id"`
	Title           string     `json:"title,omitempty"`
	Status          string     `json:"status"`
	MessageCount    int        `json:"message_count"`
	TotalCharacters int        `json:"total_characters"`
	LastMessageAt   *time.Time `json:"last_message_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// 關聯
	User      *User       `json:"user,omitempty"`
	Character *Character  `json:"character,omitempty"`
	Messages  []*Message  `json:"messages,omitempty"`
}

// Message 消息領域模型
type Message struct {
	ID                  string                 `json:"id"`
	SessionID           string                 `json:"session_id"`
	Role                string                 `json:"role"`
	Content             string                 `json:"content"`
	SceneDescription    *string                `json:"scene_description,omitempty"`
	CharacterAction     *string                `json:"character_action,omitempty"`
	EmotionalState      map[string]interface{} `json:"emotional_state,omitempty"`
	AIEngine            *string                `json:"ai_engine,omitempty"`
	ResponseTimeMs      *int                   `json:"response_time_ms,omitempty"`
	NSFWLevel           int                    `json:"nsfw_level"`
	IsRegenerated       bool                   `json:"is_regenerated"`
	RegenerationReason  *string                `json:"regeneration_reason,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`

	// 關聯
	Session *ChatSession `json:"session,omitempty"`
}



// ChatSessionResponse 會話響應格式
type ChatSessionResponse struct {
	ID              string             `json:"id"`
	UserID          string             `json:"user_id"`
	CharacterID     string             `json:"character_id"`
	Title           string             `json:"title,omitempty"`
	Status          string             `json:"status"`
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
	ID                  string                 `json:"id"`
	SessionID           string                 `json:"session_id"`
	Role                string                 `json:"role"`
	Content             string                 `json:"content"`
	SceneDescription    string                 `json:"scene_description,omitempty"`
	CharacterAction     string                 `json:"character_action,omitempty"`
	EmotionalState      map[string]interface{} `json:"emotional_state,omitempty"`
	AIEngine            string                 `json:"ai_engine,omitempty"`
	ResponseTimeMs      int                    `json:"response_time_ms,omitempty"`
	NSFWLevel           int                    `json:"nsfw_level"`
	IsRegenerated       bool                   `json:"is_regenerated"`
	RegenerationReason  string                 `json:"regeneration_reason,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
}

// ToResponse 轉換為響應格式
func (cs *ChatSession) ToResponse() *ChatSessionResponse {
	response := &ChatSessionResponse{
		ID:              cs.ID,
		UserID:          cs.UserID,
		CharacterID:     cs.CharacterID,
		Title:           cs.Title,
		Status:          cs.Status,
		MessageCount:    cs.MessageCount,
		TotalCharacters: cs.TotalCharacters,
		LastMessageAt:   cs.LastMessageAt,
		CreatedAt:       cs.CreatedAt,
		UpdatedAt:       cs.UpdatedAt,
	}

	// 添加角色信息
	if cs.Character != nil {
		response.Character = cs.Character.ToResponse()
	}

	// 添加最後一條消息
	if len(cs.Messages) > 0 {
		lastMessage := cs.Messages[len(cs.Messages)-1]
		response.LastMessage = lastMessage.ToResponse()
	}

	return response
}

func (m *Message) ToResponse() *MessageResponse {
	response := &MessageResponse{
		ID:               m.ID,
		SessionID:        m.SessionID,
		Role:             m.Role,
		Content:          m.Content,
		EmotionalState:   m.EmotionalState,
		NSFWLevel:        m.NSFWLevel,
		IsRegenerated:    m.IsRegenerated,
		CreatedAt:        m.CreatedAt,
	}
	
	// 處理指標欄位
	if m.SceneDescription != nil {
		response.SceneDescription = *m.SceneDescription
	}
	if m.CharacterAction != nil {
		response.CharacterAction = *m.CharacterAction
	}
	if m.AIEngine != nil {
		response.AIEngine = *m.AIEngine
	}
	if m.ResponseTimeMs != nil {
		response.ResponseTimeMs = *m.ResponseTimeMs
	}
	if m.RegenerationReason != nil {
		response.RegenerationReason = *m.RegenerationReason
	}
	
	return response
}

// CreateSessionRequest 創建會話請求
type CreateSessionRequest struct {
	CharacterID string `json:"character_id" binding:"required"`
	Title       string `json:"title,omitempty"`
}

// SendMessageRequest 發送消息請求
type SendMessageRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Message   string `json:"message" binding:"required,max=2000"`
}

// 聊天相關響應結構
type ChatSessionListResponse struct {
	Sessions   []*ChatSessionResponse `json:"sessions"`
	Pagination PaginationResponse     `json:"pagination"`
}

type MessageHistoryResponse struct {
	SessionID  string               `json:"session_id"`
	Messages   []*MessageResponse   `json:"messages"`
	Pagination PaginationResponse   `json:"pagination"`
}