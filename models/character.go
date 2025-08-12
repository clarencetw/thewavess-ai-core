package models

// Character 角色模型
type Character struct {
	BaseModel
	Name        string                 `json:"name" example:"陸寒淵"`
	Type        string                 `json:"type" example:"dominant" enums:"gentle,dominant,ascetic,sunny,cunning"`
	Description string                 `json:"description" example:"霸道總裁，冷峻外表下隱藏深情"`
	AvatarURL   string                 `json:"avatar_url" example:"https://example.com/character-avatar.jpg"`
	VoiceID     string                 `json:"voice_id" example:"voice_001"`
	Popularity  int                    `json:"popularity" example:"95"`
	Tags        []string               `json:"tags" example:"霸道總裁,深情,禁慾系"`
	Appearance  CharacterAppearance    `json:"appearance"`
	Personality CharacterPersonality   `json:"personality"`
	Background  string                 `json:"background" example:"商業帝國繼承人，年少成名的商界天才"`
	IsActive    bool                   `json:"is_active" example:"true"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// CharacterAppearance 角色外貌
type CharacterAppearance struct {
	Height      string `json:"height" example:"185cm"`
	HairColor   string `json:"hair_color" example:"黑髮"`
	EyeColor    string `json:"eye_color" example:"深邃黑眸"`
	Description string `json:"description" example:"俊朗五官，總是穿著剪裁合身的西裝"`
}

// CharacterPersonality 角色性格
type CharacterPersonality struct {
	Traits   []string `json:"traits" example:"冷酷,強勢,專一"`
	Likes    []string `json:"likes" example:"工作,掌控,用戶"`
	Dislikes []string `json:"dislikes" example:"被違抗,失去控制"`
}

// CharacterListResponse 角色列表回應
type CharacterListResponse struct {
	Characters []CharacterSummary `json:"characters"`
	Pagination PaginationResponse `json:"pagination"`
}

// CharacterSummary 角色摘要
type CharacterSummary struct {
	ID          string   `json:"id" example:"char_001"`
	Name        string   `json:"name" example:"陸寒淵"`
	Type        string   `json:"type" example:"dominant"`
	Description string   `json:"description" example:"霸道總裁，冷峻外表下隱藏深情"`
	AvatarURL   string   `json:"avatar_url" example:"https://example.com/avatar.jpg"`
	VoiceID     string   `json:"voice_id" example:"voice_001"`
	Popularity  int      `json:"popularity" example:"95"`
	Tags        []string `json:"tags" example:"霸道總裁,深情"`
}

// CharacterStats 角色統計數據
type CharacterStats struct {
	TotalConversations int      `json:"total_conversations" example:"1523"`
	AverageRating      float64  `json:"average_rating" example:"4.8"`
	TotalUsers         int      `json:"total_users" example:"892"`
	PopularTags        []string `json:"popular_tags" example:"溫柔,體貼,浪漫"`
	MonthlyActiveUsers int      `json:"monthly_active_users" example:"456"`
	AverageSessionTime int      `json:"average_session_time" example:"1800"`
}

// SelectCharacterRequest 選擇角色請求
type SelectCharacterRequest struct {
	CharacterID string `json:"character_id" binding:"required" example:"char_001"`
}