package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// ChatSession 聊天會話模型
type ChatSession struct {
	bun.BaseModel `bun:"table:chat_sessions,alias:cs"`

	ID              string     `bun:"id,pk" json:"id"`
	UserID          string     `bun:"user_id,notnull" json:"user_id"`
	CharacterID     string     `bun:"character_id,notnull" json:"character_id"`
	Title           string     `bun:"title" json:"title,omitempty"`
	Status          string     `bun:"status,default:'active'" json:"status"`
	MessageCount    int        `bun:"message_count,default:0" json:"message_count"`
	TotalCharacters int        `bun:"total_characters,default:0" json:"total_characters"`
	LastMessageAt   *time.Time `bun:"last_message_at" json:"last_message_at,omitempty"`
	CreatedAt       time.Time  `bun:"created_at,nullzero,default:now()" json:"created_at"`
	UpdatedAt       time.Time  `bun:"updated_at,nullzero,default:now()" json:"updated_at"`

	// 關聯
	User      *User       `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Character *Character  `bun:"rel:belongs-to,join:character_id=id" json:"character,omitempty"`
	Messages  []*Message  `bun:"rel:has-many,join:id=session_id" json:"messages,omitempty"`
}

// Message 消息模型
type Message struct {
	bun.BaseModel `bun:"table:messages,alias:m"`

	ID                  string                 `bun:"id,pk" json:"id"`
	SessionID           string                 `bun:"session_id,notnull" json:"session_id"`
	Role                string                 `bun:"role,notnull" json:"role"`
	Content             string                 `bun:"content,notnull" json:"content"`
	SceneDescription    string                 `bun:"scene_description" json:"scene_description,omitempty"`
	CharacterAction     string                 `bun:"character_action" json:"character_action,omitempty"`
	EmotionalState      map[string]interface{} `bun:"emotional_state,type:jsonb" json:"emotional_state,omitempty"`
	AIEngine            string                 `bun:"ai_engine" json:"ai_engine,omitempty"`
	ResponseTimeMs      int                    `bun:"response_time_ms" json:"response_time_ms,omitempty"`
	NSFWLevel           int                    `bun:"nsfw_level,default:0" json:"nsfw_level"`
	IsRegenerated       bool                   `bun:"is_regenerated,default:false" json:"is_regenerated"`
	RegenerationReason  string                 `bun:"regeneration_reason" json:"regeneration_reason,omitempty"`
	CreatedAt           time.Time              `bun:"created_at,nullzero,default:now()" json:"created_at"`

	// 關聯
	Session *ChatSession `bun:"rel:belongs-to,join:session_id=id" json:"session,omitempty"`
}


// TableName 返回數據庫表名
func (cs *ChatSession) TableName() string {
	return "chat_sessions"
}

func (m *Message) TableName() string {
	return "messages"
}


// BeforeAppendModel 在模型操作前執行
func (cs *ChatSession) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.UpdateQuery:
		cs.UpdatedAt = time.Now()
	}
	return nil
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
	return &MessageResponse{
		ID:                  m.ID,
		SessionID:           m.SessionID,
		Role:                m.Role,
		Content:             m.Content,
		SceneDescription:    m.SceneDescription,
		CharacterAction:     m.CharacterAction,
		EmotionalState:      m.EmotionalState,
		AIEngine:            m.AIEngine,
		ResponseTimeMs:      m.ResponseTimeMs,
		NSFWLevel:           m.NSFWLevel,
		IsRegenerated:       m.IsRegenerated,
		RegenerationReason:  m.RegenerationReason,
		CreatedAt:           m.CreatedAt,
	}
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