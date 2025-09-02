package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models/db"
)

// CharacterType 角色類型枚舉
type CharacterType string

const (
	CharacterTypeDominant CharacterType = "dominant" // 霸道型（如霸道總裁）
	CharacterTypeGentle   CharacterType = "gentle"   // 溫柔型（如溫柔醫生）
	CharacterTypePlayful  CharacterType = "playful"  // 活潑型
	CharacterTypeMystery  CharacterType = "mystery"  // 神秘型
	CharacterTypeReliable CharacterType = "reliable" // 可靠型
)

// Locale 語言區域枚舉
type Locale string

const (
	LocaleChinese Locale = "zh-TW" // 繁體中文
)


// EngineType AI 引擎類型枚舉
type EngineType string

const (
	EngineOpenAI EngineType = "openai" // OpenAI GPT
	EngineGrok   EngineType = "grok"   // Grok
)

// NSFWLevel NSFW 等級枚舉
type NSFWLevel int

const (
	NSFWLevelSafe     NSFWLevel = 1 // 安全內容
	NSFWLevelRomantic NSFWLevel = 2 // 浪漫內容
	NSFWLevelIntimate NSFWLevel = 3 // 親密內容
	NSFWLevelAdult    NSFWLevel = 4 // 成人內容
	NSFWLevelExplicit NSFWLevel = 5 // 明確成人內容
)

// Character 核心角色模型
type Character struct {
	// 基本身份信息
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Type     CharacterType `json:"type"`
	Locale   Locale        `json:"locale"`
	IsActive bool          `json:"is_active"`

	// 用戶自由描述
	UserDescription *string `json:"user_description,omitempty"` // 用戶自由描述角色的詳細內容

	// 元數據信息
	Metadata CharacterMetadata `json:"metadata"`

	// 用戶追蹤字段
	CreatedBy *string `json:"created_by,omitempty"` // 創建者用戶ID
	UpdatedBy *string `json:"updated_by,omitempty"` // 最後更新者用戶ID
	
	// 角色狀態
	IsPublic  bool `json:"is_public"`  // 是否公開（所有人可使用）
	IsSystem  bool `json:"is_system"`  // 是否為系統角色

	// 時間戳記（包含軟刪除）
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// CharacterMetadata 角色基本元數據
type CharacterMetadata struct {
	AvatarURL  *string  `json:"avatar_url,omitempty"` // 頭像圖片URL
	Tags       []string `json:"tags"`                 // 角色標籤（如：霸總、腹黑、現代、成熟）
	Popularity int      `json:"popularity"`           // 人氣值（0-100）
}



// CharacterStats 角色統計數據（分離的統計模型，主要用於分析報表）
type CharacterStats struct {
	CharacterID       string                     `json:"character_id"`       // 角色ID
	BasicInfo         CharacterBasicInfo         `json:"basic_info"`         // 基本信息
	InteractionStats  CharacterInteractionStats  `json:"interaction_stats"`  // 互動統計（對話次數、用戶數等）
	RelationshipStats CharacterRelationshipStats `json:"relationship_stats"` // 關係統計（好感度、親密度等，較少使用）
	ContentStats      CharacterContentStats      `json:"content_stats"`      // 內容統計（NSFW分佈、場景使用等）
	PerformanceStats  CharacterPerformanceStats  `json:"performance_stats"`  // 性能統計（回應時間、成功率等）
	GeneratedAt       time.Time                  `json:"generated_at"`       // 統計生成時間
}

// CharacterBasicInfo 基本信息統計
type CharacterBasicInfo struct {
	Name        string    `json:"name"`        // 角色名稱
	Type        string    `json:"type"`        // 角色類型（如霸道總裁、溫柔醫生）
	Description string    `json:"description"` // 角色描述
	Tags        []string  `json:"tags"`        // 角色標籤
	Popularity  int       `json:"popularity"`  // 人氣值 (0-100)
	IsActive    bool      `json:"is_active"`   // 是否啟用
	CreatedAt   time.Time `json:"created_at"`  // 創建時間
	Version     string    `json:"version"`     // 版本號
}

// CharacterInteractionStats 互動統計
type CharacterInteractionStats struct {
	TotalConversations int            `json:"total_conversations"`        // 總對話次數
	TotalMessages      int            `json:"total_messages"`             // 總消息數
	TotalUsers         int            `json:"total_users"`                // 總用戶數
	AvgSessionLength   int64          `json:"avg_session_length_seconds"` // 平均對話時長 (秒)
	LastInteraction    *time.Time     `json:"last_interaction"`           // 最後互動時間
	ActiveDays         int            `json:"active_days"`                // 活躍天數
	EngineUsage        map[string]int `json:"engine_usage"`               // AI 引擎使用統計 (OpenAI/Grok)
}

// CharacterRelationshipStats 關係統計（專注核心指標）
type CharacterRelationshipStats struct {
	AvgAffectionLevel  float64        `json:"avg_affection_level"`  // 平均好感度（核心指標）
	TotalRelationships int            `json:"total_relationships"`  // 總關係數
	RelationshipTypes  map[string]int `json:"relationship_types"`   // 關係類型分佈（重要業務指標）
	IntimacyLevels     map[string]int `json:"intimacy_levels"`      // 親密度分佈（重要業務指標）
}

// CharacterContentStats 內容統計（簡化為實際需要的統計）
type CharacterContentStats struct {
	NSFWLevelDistribution map[string]int `json:"nsfw_level_distribution"` // NSFW 等級分佈 (1-5 級)
}


// CharacterPerformanceStats 性能統計（簡化為基本指標）
type CharacterPerformanceStats struct {
	AvgResponseTime int64   `json:"avg_response_time_ms"` // 平均回應時間 (毫秒)
	SuccessRate     float64 `json:"success_rate"`         // 成功率 (0-1)
}

// 請求和響應結構

// CharacterCreateRequest 創建角色請求
type CharacterCreateRequest struct {
	Name   string        `json:"name" binding:"required,min=1,max=50"`
	Type   CharacterType `json:"type" binding:"required"`
	Locale Locale        `json:"locale" binding:"required"`

	// 角色描述
	UserDescription *string `json:"user_description,omitempty"` // 用戶自由描述角色

	// 可選的基本配置
	Metadata *CharacterMetadataRequest `json:"metadata,omitempty"`
}

// CharacterUpdateRequest 更新角色請求
type CharacterUpdateRequest struct {
	Name            *string                   `json:"name,omitempty" binding:"omitempty,min=1,max=50"`
	Type            *CharacterType            `json:"type,omitempty"`
	IsActive        *bool                     `json:"is_active,omitempty"`
	UserDescription *string                   `json:"user_description,omitempty"`
	Metadata        *CharacterMetadataRequest `json:"metadata,omitempty"`
}

// CharacterMetadataRequest 元數據請求
type CharacterMetadataRequest struct {
	AvatarURL  *string  `json:"avatar_url,omitempty"`
	Tags       []string `json:"tags"`
	Popularity *int     `json:"popularity,omitempty" binding:"omitempty,min=0,max=100"`
}

// CharacterResponse 角色響應
type CharacterResponse struct {
	ID              string                    `json:"id"`
	Name            string                    `json:"name"`
	Type            CharacterType             `json:"type"`
	IsActive        bool                      `json:"is_active"`
	UserDescription *string                   `json:"user_description,omitempty"`
	Metadata        CharacterMetadataResponse `json:"metadata"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

// CharacterMetadataResponse 元數據響應
type CharacterMetadataResponse struct {
	AvatarURL  *string  `json:"avatar_url,omitempty"`
	Tags       []string `json:"tags"`
	Popularity int      `json:"popularity"`
}

// CharacterListResponse 角色列表響應
type CharacterListResponse struct {
	Characters []*CharacterResponse `json:"characters"`
	Pagination PaginationResponse   `json:"pagination"`
}

// CharacterListQuery 角色列表查詢參數
type CharacterListQuery struct {
	Page      int           `form:"page,default=1" binding:"min=1"`
	PageSize  int           `form:"page_size,default=20" binding:"min=1,max=100"`
	Type      CharacterType `form:"type"`
	IsActive  *bool         `form:"is_active"`
	Tags      []string      `form:"tags"`
	Search    string        `form:"search"`
	SortBy    string        `form:"sort_by,default=created_at"`
	SortOrder string        `form:"sort_order,default=desc" binding:"oneof=asc desc"`
}

// 驗證方法

// IsValid 驗證角色類型
func (ct CharacterType) IsValid() bool {
	switch ct {
	case CharacterTypeDominant, CharacterTypeGentle, CharacterTypePlayful, CharacterTypeMystery, CharacterTypeReliable:
		return true
	}
	return false
}

// IsValid 驗證語言區域
func (l Locale) IsValid() bool {
	switch l {
	case LocaleChinese:
		return true
	}
	return false
}


// IsValid 驗證引擎類型
func (et EngineType) IsValid() bool {
	switch et {
	case EngineOpenAI, EngineGrok:
		return true
	}
	return false
}

// IsValid 驗證NSFW等級
func (nl NSFWLevel) IsValid() bool {
	return nl >= NSFWLevelSafe && nl <= NSFWLevelExplicit
}


// 域專用錯誤處理類型

// CharacterError 角色相關錯誤
type CharacterError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e CharacterError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Type, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// 常見錯誤建構器
func NewCharacterNotFoundError(id string) CharacterError {
	return CharacterError{
		Type:    "CHARACTER_NOT_FOUND",
		Message: fmt.Sprintf("角色 %s 不存在", id),
		Field:   "id",
	}
}

func NewCharacterValidationError(field, message string) CharacterError {
	return CharacterError{
		Type:    "CHARACTER_VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

func NewCharacterConfigurationError(message string) CharacterError {
	return CharacterError{
		Type:    "CHARACTER_CONFIGURATION_ERROR",
		Message: message,
	}
}

// 模型轉換方法

// ToResponse 轉換為基本響應格式
func (c *Character) ToResponse() *CharacterResponse {
	return &CharacterResponse{
		ID:              c.ID,
		Name:            c.Name,
		Type:            c.Type,
		IsActive:        c.IsActive,
		UserDescription: c.UserDescription,
		Metadata: CharacterMetadataResponse{
			AvatarURL:  c.Metadata.AvatarURL,
			Tags:       c.Metadata.Tags,
			Popularity: c.Metadata.Popularity,
		},
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// FromCreateRequest 從創建請求構建角色
func (c *Character) FromCreateRequest(req *CharacterCreateRequest) {
	c.Name = req.Name
	c.Type = req.Type
	c.Locale = req.Locale
	c.IsActive = true

	// 設置用戶描述
	if req.UserDescription != nil {
		c.UserDescription = req.UserDescription
	}

	// 設置基本元數據
	if req.Metadata != nil {
		if req.Metadata.AvatarURL != nil {
			c.Metadata.AvatarURL = req.Metadata.AvatarURL
		}
		if req.Metadata.Tags != nil {
			c.Metadata.Tags = req.Metadata.Tags
		}
		if req.Metadata.Popularity != nil {
			c.Metadata.Popularity = *req.Metadata.Popularity
		}
	}

	// 設置時間戳記
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
}

// ApplyUpdateRequest 應用更新請求
func (c *Character) ApplyUpdateRequest(req *CharacterUpdateRequest) {
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.Type != nil {
		c.Type = *req.Type
	}
	// Locale 固定為中文
	c.Locale = LocaleChinese
	if req.IsActive != nil {
		c.IsActive = *req.IsActive
	}

	// 更新元數據
	if req.Metadata != nil {
		if req.Metadata.AvatarURL != nil {
			c.Metadata.AvatarURL = req.Metadata.AvatarURL
		}
		if req.Metadata.Tags != nil {
			c.Metadata.Tags = req.Metadata.Tags
		}
		if req.Metadata.Popularity != nil {
			c.Metadata.Popularity = *req.Metadata.Popularity
		}
	}

	// 更新時間戳記
	c.UpdatedAt = time.Now()
}

// Utility methods

// GetName 獲取角色名稱
func (c *Character) GetName() string {
	return c.Name
}

// MatchesQuery 檢查角色是否符合查詢條件
func (c *Character) MatchesQuery(query *CharacterListQuery) bool {
	// 類型過濾
	if query.Type != "" && c.Type != query.Type {
		return false
	}

	// 狀態過濾
	if query.IsActive != nil && c.IsActive != *query.IsActive {
		return false
	}

	// 標籤過濾
	if len(query.Tags) > 0 {
		hasTag := false
		for _, queryTag := range query.Tags {
			for _, charTag := range c.Metadata.Tags {
				if strings.EqualFold(charTag, queryTag) {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	// 搜索過濾
	if query.Search != "" {
		search := strings.ToLower(query.Search)
		userDesc := ""
		if c.UserDescription != nil {
			userDesc = strings.ToLower(*c.UserDescription)
		}

		if !strings.Contains(strings.ToLower(c.Name), search) &&
			!strings.Contains(userDesc, search) {
			return false
		}
	}

	return true
}

// Validate 驗證角色數據
func (c *Character) Validate() error {
	if c.ID == "" {
		return NewCharacterValidationError("id", "角色ID不能為空")
	}
	if c.Name == "" {
		return NewCharacterValidationError("name", "角色名稱不能為空")
	}
	if !c.Type.IsValid() {
		return NewCharacterValidationError("type", "無效的角色類型")
	}
	// 語言區域固定為中文
	c.Locale = LocaleChinese

	return nil
}

// CharacterFromDB 從資料庫模型轉換為領域模型
func CharacterFromDB(charDB *db.CharacterDB) *Character {
	if charDB == nil {
		return nil
	}

	character := &Character{
		ID:              charDB.ID,
		Name:            charDB.Name,
		Type:            CharacterType(charDB.Type),
		Locale:          Locale(charDB.Locale),
		IsActive:        charDB.IsActive,
		UserDescription: charDB.UserDescription,
		Metadata: CharacterMetadata{
			AvatarURL:  charDB.AvatarURL,
			Tags:       charDB.Tags,
			Popularity: charDB.Popularity,
		},
		CreatedAt: charDB.CreatedAt,
		UpdatedAt: charDB.UpdatedAt,
	}

	return character
}
