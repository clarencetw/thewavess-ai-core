package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

// GetCharacterList godoc
// @Summary      獲取角色列表
// @Description  獲取可用角色列表，支援分頁和篩選
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(20)
// @Param        type query string false "角色類型篩選" Enums(gentle,dominant,ascetic,sunny,cunning)
// @Param        tags query string false "標籤篩選，多個用逗號分隔"
// @Success      200 {object} models.APIResponse{data=models.CharacterListResponse} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /character/list [get]
func GetCharacterList(c *gin.Context) {
	// 驗證認證
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 20 {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "缺少或無效的認證 Token",
			},
		})
		return
	}

	// 解析查詢參數
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := utils.ParseInt(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := utils.ParseInt(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// 模擬角色數據
	allCharacters := []models.Character{
		{
			BaseModel: models.BaseModel{
				ID:        "char_001",
				CreatedAt: time.Now().AddDate(0, -1, 0),
				UpdatedAt: time.Now(),
			},
			Name:        "陸寒淵",
			Type:        "dominant",
			Description: "霸道總裁，冷峻外表下隱藏深情",
			AvatarURL:   "https://example.com/avatars/lu_hanyuan.jpg",
			VoiceID:     "voice_001",
			Popularity:  95,
			Tags:        []string{"霸道總裁", "深情", "禁慾系"},
			Appearance: models.CharacterAppearance{
				Height:      "185cm",
				HairColor:   "黑髮",
				EyeColor:    "深邃黑眸",
				Description: "俊朗五官，總是穿著剪裁合身的西裝",
			},
			Personality: models.CharacterPersonality{
				Traits:   []string{"冷酷", "強勢", "專一", "佔有欲"},
				Likes:    []string{"工作", "掌控", "用戶"},
				Dislikes: []string{"被違抗", "失去控制"},
			},
			Background:       "跨國集團CEO，商業帝國繼承人",
		},
		{
			BaseModel: models.BaseModel{
				ID:        "char_002",
				CreatedAt: time.Now().AddDate(0, -1, 0),
				UpdatedAt: time.Now(),
			},
			Name:        "沈言墨",
			Type:        "gentle",
			Description: "溫柔醫生，治癒系學長",
			AvatarURL:   "https://example.com/avatars/shen_yanmo.jpg",
			VoiceID:     "voice_002",
			Popularity:  88,
			Tags:        []string{"溫柔", "醫生", "治癒系"},
			Appearance: models.CharacterAppearance{
				Height:      "180cm",
				HairColor:   "栗色短髮",
				EyeColor:    "溫潤琥珀色",
				Description: "溫和的笑容，常穿白大褂或休閒裝",
			},
			Personality: models.CharacterPersonality{
				Traits:   []string{"溫柔", "體貼", "細心", "略帶腹黑"},
				Likes:    []string{"醫學", "幫助他人", "用戶"},
				Dislikes: []string{"看到痛苦", "無能為力"},
			},
			Background:       "醫學研究生，醫學世家出身",
		},
	}

	// 應用篩選
	typeFilter := c.Query("type")
	var filteredCharacters []models.Character
	for _, char := range allCharacters {
		if typeFilter == "" || char.Type == typeFilter {
			filteredCharacters = append(filteredCharacters, char)
		}
	}

	// 分頁處理
	totalCount := len(filteredCharacters)
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	if endIndex > totalCount {
		endIndex = totalCount
	}

	var paginatedCharacters []models.Character
	if startIndex < totalCount {
		paginatedCharacters = filteredCharacters[startIndex:endIndex]
	}

	totalPages := (totalCount + limit - 1) / limit

	response := map[string]interface{}{
		"characters": paginatedCharacters,
		"pagination": map[string]interface{}{
			"current_page": page,
			"total_pages":  totalPages,
			"total_count":  totalCount,
			"has_next":     page < totalPages,
			"has_prev":     page > 1,
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取角色列表成功",
		Data:    response,
	})
}

// GetCharacterDetails - 已移除（未實現功能）

// GetCharacterStats - 已移除（未實現功能）

// GetCurrentCharacter - 已移除（未實現功能）

// SelectCharacter - 已移除（未實現功能）