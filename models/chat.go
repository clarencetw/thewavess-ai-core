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

// MessageResponse 標準消息響應結構（所有消息 API 的共用核心格式）
type MessageResponse struct {
	ID               string                 `json:"id"`
	ChatID           string                 `json:"chat_id"`
	Role             string                 `json:"role"`
	Dialogue         string                 `json:"dialogue"`
	SceneDescription string                 `json:"scene_description,omitempty"`
	Action           string                 `json:"action,omitempty"`
	EmotionalState   map[string]interface{} `json:"emotional_state,omitempty"`
	AIEngine         string                 `json:"ai_engine"`
	ResponseTimeMs   int                    `json:"response_time_ms"`
	NSFWLevel        int                    `json:"nsfw_level"`
	CreatedAt        time.Time              `json:"created_at"`
}

// DetailedMessageResponse 詳細消息響應格式（用於 GetMessageHistory、ChatResponse.LastMessage 等）
// 擴展 MessageResponse 增加消息狀態相關欄位
type DetailedMessageResponse struct {
	MessageResponse
	IsRegenerated      bool   `json:"is_regenerated"`
	RegenerationReason string `json:"regeneration_reason,omitempty"`
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

	// 添加最後一條消息（ChatResponse 只需要基本消息資訊）
	if len(c.Messages) > 0 {
		lastMessage := c.Messages[len(c.Messages)-1]
		response.LastMessage = &MessageResponse{
			ID:             lastMessage.ID,
			ChatID:         lastMessage.ChatID,
			Role:           lastMessage.Role,
			Dialogue:       lastMessage.Dialogue,
			EmotionalState: lastMessage.EmotionalState,
			AIEngine:       lastMessage.AIEngine,
			ResponseTimeMs: lastMessage.ResponseTimeMs,
			NSFWLevel:      lastMessage.NSFWLevel,
			CreatedAt:      lastMessage.CreatedAt,
		}
		// 處理指標欄位
		if lastMessage.SceneDescription != nil {
			response.LastMessage.SceneDescription = *lastMessage.SceneDescription
		}
		if lastMessage.Action != nil {
			response.LastMessage.Action = *lastMessage.Action
		}
	}

	return response
}

// ToResponse 轉換為詳細 API 響應格式
func (m *Message) ToResponse() *DetailedMessageResponse {
	response := &DetailedMessageResponse{
		MessageResponse: MessageResponse{
			ID:             m.ID,
			ChatID:         m.ChatID,
			Role:           m.Role,
			Dialogue:       m.Dialogue,
			EmotionalState: m.EmotionalState,
			AIEngine:       m.AIEngine,
			ResponseTimeMs: m.ResponseTimeMs,
			NSFWLevel:      m.NSFWLevel,
			CreatedAt:      m.CreatedAt,
		},
		IsRegenerated: m.IsRegenerated,
	}

	// 處理指標欄位
	if m.SceneDescription != nil {
		response.MessageResponse.SceneDescription = *m.SceneDescription
	}
	if m.Action != nil {
		response.MessageResponse.Action = *m.Action
	}
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

// SendMessageResponse 發送消息響應（SendMessage API 專用格式）
// 部分欄位使用前端相容命名，包含發送消息專用的額外欄位
type SendMessageResponse struct {
	ChatID         string    `json:"chat_id"`
	MessageID      string    `json:"message_id"` // 前端格式：message_id 而非 id
	Content        string    `json:"content"`    // 前端格式：content 而非 dialogue
	AIEngine       string    `json:"ai_engine"`
	NSFWLevel      int       `json:"nsfw_level"`
	ResponseTimeMs int64     `json:"response_time_ms"` // 統一使用 response_time_ms
	Timestamp      time.Time `json:"timestamp"`        // 響應時間戳
	// SendMessage 專用欄位
	Affection  int     `json:"affection"`  // 好感度值
	Confidence float64 `json:"confidence"` // NSFW 分級信心度
	ChatMode   string  `json:"chat_mode"`  // 對話模式："chat" 或 "novel"
}

// RegenerateMessageResponse 重新生成消息響應（RegenerateMessage API 專用格式）
type RegenerateMessageResponse struct {
	Message           *DetailedMessageResponse `json:"message"`             // 新生成的消息內容
	PreviousMessageID string                   `json:"previous_message_id"` // 被替換的原消息ID
	Regenerated       bool                     `json:"regenerated"`         // 重新生成標記，固定為 true
}

// ChatListResponse 聊天會話列表響應（GetChats API 專用格式）
type ChatListResponse struct {
	Chats      []*ChatResponse    `json:"chats"`      // 會話列表
	Pagination PaginationResponse `json:"pagination"` // 分頁信息
}

// MessageHistoryResponse 聊天歷史響應（GetMessageHistory API 專用格式）
type MessageHistoryResponse struct {
	ChatID     string                     `json:"chat_id"`    // 會話ID
	Messages   []*DetailedMessageResponse `json:"messages"`   // 消息歷史列表
	Pagination PaginationResponse         `json:"pagination"` // 分頁信息
}
