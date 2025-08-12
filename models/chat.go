package models

import "time"

// ChatSession 對話會話模型
type ChatSession struct {
	BaseModel
	UserID      string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CharacterID string `json:"character_id" example:"char_001"`
	Title       string `json:"title" example:"與陸寒淵的對話"`
	Mode        string `json:"mode" example:"normal" enums:"normal,novel,nsfw"`
	Status      string `json:"status" example:"active" enums:"active,ended,paused"`
	Tags        []string `json:"tags" example:"浪漫,日常"`
	MessageCount int `json:"message_count" example:"25"`
	LastMessageAt time.Time `json:"last_message_at" example:"2023-12-01T12:00:00Z"`
}

// CreateSessionRequest 創建會話請求
type CreateSessionRequest struct {
	CharacterID string   `json:"character_id" binding:"required" example:"char_001"`
	Title       string   `json:"title,omitempty" example:"新的對話"`
	Mode        string   `json:"mode,omitempty" example:"normal" enums:"normal,novel,nsfw"`
	Tags        []string `json:"tags,omitempty" example:"浪漫,日常"`
}

// ChatMessage 對話訊息模型
type ChatMessage struct {
	BaseModel
	SessionID   string                 `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role        string                 `json:"role" example:"user" enums:"user,assistant"`
	Content     string                 `json:"content" example:"你好"`
	MessageType string                 `json:"message_type" example:"text" enums:"text,image,audio"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SendMessageRequest 發送訊息請求
type SendMessageRequest struct {
	SessionID string   `json:"session_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Message   string   `json:"message" binding:"required" example:"你好"`
	Tags      []string `json:"tags,omitempty" example:"問候"`
	Context   MessageContext `json:"context,omitempty"`
}

// MessageContext 訊息上下文
type MessageContext struct {
	Affection    int    `json:"affection,omitempty" example:"75"`
	Relationship string `json:"relationship,omitempty" example:"friend" enums:"stranger,friend,ambiguous,lover"`
	Scene        string `json:"scene,omitempty" example:"辦公室"`
}

// ChatResponse 對話回應
type ChatResponse struct {
	SessionID       string        `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MessageID       string        `json:"message_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	CharacterID     string        `json:"character_id" example:"char_001"`
	Response        string        `json:"response" example:"你好，很高興見到你"`
	Emotion         string        `json:"emotion" example:"happy" enums:"happy,sad,angry,shy,excited"`
	AffectionChange int           `json:"affection_change" example:"5"`
	EngineUsed      string        `json:"engine_used" example:"openai" enums:"openai,grok"`
	NovelChoices    []NovelChoice `json:"novel_choices,omitempty"`
	SpecialEvent    *SpecialEvent `json:"special_event,omitempty"`
}

// NovelChoice 小說選項
type NovelChoice struct {
	ID          string `json:"id" example:"choice_001"`
	Text        string `json:"text" example:"選擇A：主動握住他的手"`
	Consequence string `json:"consequence" example:"好感度+10"`
}

// SpecialEvent 特殊事件
type SpecialEvent struct {
	Triggered   bool   `json:"triggered" example:"true"`
	Type        string `json:"type" example:"confession"`
	Description string `json:"description" example:"陸寒淵向你表白了"`
}

// SessionListResponse 會話列表回應
type SessionListResponse struct {
	Sessions   []ChatSession      `json:"sessions"`
	Pagination PaginationResponse `json:"pagination"`
}

// MessageHistoryResponse 訊息歷史回應
type MessageHistoryResponse struct {
	Messages   []ChatMessage      `json:"messages"`
	Pagination PaginationResponse `json:"pagination"`
}

// UpdateModeRequest 切換模式請求
type UpdateModeRequest struct {
	Mode              string `json:"mode" binding:"required" example:"novel" enums:"normal,novel,nsfw"`
	TransitionMessage string `json:"transition_message,omitempty" example:"我們來玩個遊戲吧"`
}

// AddTagsRequest 添加標籤請求
type AddTagsRequest struct {
	Tags []string `json:"tags" binding:"required" example:"浪漫,甜蜜,日常"`
}

// RegenerateRequest 重新生成請求
type RegenerateRequest struct {
	MessageID string `json:"message_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
}