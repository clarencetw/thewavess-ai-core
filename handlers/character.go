package handlers

import (
	"net/http"
	"strconv"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 全域服務實例 (lazy initialization)
var (
	characterService *services.CharacterService
)

// getCharacterService 獲取角色服務實例 (lazy initialization)
func getCharacterService() *services.CharacterService {
	if characterService == nil {
		characterService = services.GetCharacterService()
	}
	return characterService
}

// GetCharacterList godoc
// @Summary      獲取角色列表
// @Description  獲取角色列表，支援分頁、篩選和排序
// @Tags         Character
// @Accept       json
// @Produce      json
// @Param        page query int false "頁數" default(1)
// @Param        limit query int false "每頁數量" default(20)
// @Param        category query string false "角色類別"
// @Param        gender query string false "性別篩選"
// @Param        tag query string false "標籤篩選"
// @Param        sort_by query string false "排序欄位" default(created_at)
// @Param        sort_order query string false "排序方向" default(desc)
// @Success      200 {object} models.APIResponse{data=models.CharacterListResponse} "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character/list [get]
func GetCharacterList(c *gin.Context) {
	var query models.CharacterListQuery
	
	// 綁定查詢參數
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_QUERY_PARAMS",
				Message: "查詢參數錯誤: " + err.Error(),
			},
		})
		return
	}

	characters, pagination, err := getCharacterService().ListCharacters(c.Request.Context(), &query)
	if err != nil {
		utils.Logger.WithError(err).Error("獲取角色列表失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SERVICE_ERROR",
				Message: "獲取角色列表失敗",
			},
		})
		return
	}

	// 轉換為響應格式
	characterResponses := make([]*models.CharacterResponse, len(characters))
	for i, char := range characters {
		characterResponses[i] = char.ToResponse()
	}

	response := &models.CharacterListResponse{
		Characters: characterResponses,
		Pagination: *pagination,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取角色列表成功",
		Data:    response,
	})
}

// GetCharacterByID godoc
// @Summary      獲取角色詳情
// @Description  根據角色ID獲取詳細角色信息
// @Tags         Character
// @Accept       json
// @Produce      json
// @Param        id path string true "角色ID"
// @Success      200 {object} models.APIResponse{data=models.CharacterDetailResponse} "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character/{id} [get]
func GetCharacterByID(c *gin.Context) {
	characterID := c.Param("id")

	character, err := getCharacterService().GetCharacter(c.Request.Context(), characterID)
	if err != nil {
		if characterError, ok := err.(models.CharacterError); ok && characterError.Type == "CHARACTER_NOT_FOUND" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "CHARACTER_NOT_FOUND",
					Message: "角色不存在",
				},
			})
			return
		}

		utils.Logger.WithError(err).WithField("character_id", characterID).Error("獲取角色失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SERVICE_ERROR",
				Message: "獲取角色失敗",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取角色詳情成功",
		Data:    character.ToDetailResponse(),
	})
}

// CreateCharacter godoc
// @Summary      創建角色
// @Description  創建新的AI角色
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        character body models.CharacterCreateRequest true "角色信息"
// @Success      201 {object} models.APIResponse{data=models.CharacterResponse} "創建成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      403 {object} models.APIResponse{error=models.APIError} "禁止訪問"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character [post]
func CreateCharacter(c *gin.Context) {
	var req models.CharacterCreateRequest
	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	character, err := getCharacterService().CreateCharacter(c.Request.Context(), &req)
	if err != nil {
		utils.Logger.WithError(err).Error("創建角色失敗")
		
		if validationError, ok := err.(models.CharacterError); ok {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    validationError.Type,
					Message: validationError.Message,
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SERVICE_ERROR",
				Message: "創建角色失敗",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "角色創建成功",
		Data:    character.ToResponse(),
	})
}

// UpdateCharacter godoc
// @Summary      更新角色
// @Description  更新指定角色的信息
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "角色ID"
// @Param        character body models.CharacterUpdateRequest true "角色更新信息"
// @Success      200 {object} models.APIResponse{data=models.CharacterResponse} "更新成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      403 {object} models.APIResponse{error=models.APIError} "禁止訪問"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character/{id} [put]
func UpdateCharacter(c *gin.Context) {
	characterID := c.Param("id")

	var req models.CharacterUpdateRequest
	if !utils.ValidationHelperInstance.BindAndValidate(c, &req) {
		return
	}

	character, err := getCharacterService().UpdateCharacter(c.Request.Context(), characterID, &req)
	if err != nil {
		if characterError, ok := err.(models.CharacterError); ok {
			var statusCode int
			switch characterError.Type {
			case "CHARACTER_NOT_FOUND":
				statusCode = http.StatusNotFound
			case "CHARACTER_VALIDATION_ERROR":
				statusCode = http.StatusBadRequest
			default:
				statusCode = http.StatusInternalServerError
			}
			
			c.JSON(statusCode, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    characterError.Type,
					Message: characterError.Message,
				},
			})
			return
		}

		utils.Logger.WithError(err).WithField("character_id", characterID).Error("更新角色失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SERVICE_ERROR",
				Message: "更新角色失敗",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "角色更新成功",
		Data:    character.ToResponse(),
	})
}

// DeleteCharacter godoc
// @Summary      刪除角色
// @Description  刪除指定角色（系統預設角色無法刪除）
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "角色ID"
// @Success      200 {object} models.APIResponse{data=object} "刪除成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      403 {object} models.APIResponse{error=models.APIError} "禁止訪問或系統角色保護"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character/{id} [delete]
func DeleteCharacter(c *gin.Context) {
	characterID := c.Param("id")

	// 檢查是否為系統預設角色
	systemCharacters := []string{"char_001", "char_002"}
	for _, sysChar := range systemCharacters {
		if characterID == sysChar {
			c.JSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "SYSTEM_CHARACTER_PROTECTED",
					Message: "系統預設角色無法刪除",
				},
			})
			return
		}
	}

	err := getCharacterService().DeleteCharacter(c.Request.Context(), characterID)
	if err != nil {
		if characterError, ok := err.(models.CharacterError); ok && characterError.Type == "CHARACTER_NOT_FOUND" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "CHARACTER_NOT_FOUND",
					Message: "角色不存在",
				},
			})
			return
		}

		utils.Logger.WithError(err).WithField("character_id", characterID).Error("刪除角色失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SERVICE_ERROR",
				Message: "刪除角色失敗",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "角色刪除成功",
		Data: map[string]interface{}{
			"character_id": characterID,
			"deleted_at":   utils.Now(),
		},
	})
}

// GetCharacterStats godoc
// @Summary      獲取角色統計
// @Description  獲取指定角色的使用統計信息
// @Tags         Character
// @Accept       json
// @Produce      json
// @Param        id path string true "角色ID"
// @Success      200 {object} models.APIResponse{data=object} "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character/{id}/stats [get]
func GetCharacterStats(c *gin.Context) {
	characterID := c.Param("id")

	stats, err := getCharacterService().GetCharacterStats(c.Request.Context(), characterID)
	if err != nil {
		if characterError, ok := err.(models.CharacterError); ok && characterError.Type == "CHARACTER_NOT_FOUND" {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Success: false,
				Error: &models.APIError{
					Code:    "CHARACTER_NOT_FOUND",
					Message: "角色不存在",
				},
			})
			return
		}

		utils.Logger.WithError(err).WithField("character_id", characterID).Error("獲取角色統計失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SERVICE_ERROR",
				Message: "獲取角色統計失敗",
			},
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取角色統計成功",
		Data:    stats,
	})
}

// SearchCharacters godoc
// @Summary      搜尋角色
// @Description  根據關鍵詞搜尋角色
// @Tags         Character
// @Accept       json
// @Produce      json
// @Param        q query string true "搜尋關鍵詞"
// @Param        limit query int false "結果數量限制" default(10) maximum(50)
// @Success      200 {object} models.APIResponse{data=object} "搜尋成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character/search [get]
func SearchCharacters(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_KEYWORD",
				Message: "搜尋關鍵詞不能為空",
			},
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 50 {
		limit = 10
	}

	characters, err := getCharacterService().SearchCharacters(c.Request.Context(), keyword, limit)
	if err != nil {
		utils.Logger.WithError(err).WithField("keyword", keyword).Error("搜尋角色失敗")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "SERVICE_ERROR",
				Message: "搜尋角色失敗",
			},
		})
		return
	}

	// 轉換為響應格式
	characterResponses := make([]*models.CharacterResponse, len(characters))
	for i, char := range characters {
		characterResponses[i] = char.ToResponse()
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "搜尋角色成功",
		Data: gin.H{
			"keyword":    keyword,
			"characters": characterResponses,
			"count":      len(characterResponses),
		},
	})
}

// =============================================================================
// CHARACTER CONFIGURATION HANDLERS
// =============================================================================

// ErrorResponse 統一錯誤回應
func ErrorResponse(c *gin.Context, statusCode int, message, code string) {
	c.JSON(statusCode, models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
		},
	})
}

// SuccessResponse 統一成功回應
func SuccessResponse(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// GetCharacterProfile godoc
// @Summary      獲取角色檔案
// @Description  獲取指定角色的詳細檔案信息
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "角色ID"
// @Success      200 {object} models.APIResponse{data=object} "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character/{id}/profile [get]
func GetCharacterProfile(c *gin.Context) {
	characterID := c.Param("id")
	if characterID == "" {
		ErrorResponse(c, http.StatusBadRequest, "角色ID不能為空", "INVALID_CHARACTER_ID")
		return
	}

	character, err := getCharacterService().GetCharacter(c.Request.Context(), characterID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"error":       err.Error(),
		}).Error("獲取角色失敗")
		
		ErrorResponse(c, http.StatusNotFound, "找不到角色", "CHARACTER_NOT_FOUND")
		return
	}

	SuccessResponse(c, character, "獲取角色成功")
}

// GetCharacterSpeechStyles godoc
// @Summary      獲取角色對話風格
// @Description  獲取指定角色的所有對話風格，可根據類型篩選
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "角色ID"
// @Param        style_type query string false "風格類型篩選"
// @Success      200 {object} models.APIResponse{data=object} "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character/{id}/speech-styles [get]
func GetCharacterSpeechStyles(c *gin.Context) {
	characterID := c.Param("id")
	if characterID == "" {
		ErrorResponse(c, http.StatusBadRequest, "角色ID不能為空", "INVALID_CHARACTER_ID")
		return
	}

	character, err := getCharacterService().GetCharacter(c.Request.Context(), characterID)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"error":       err.Error(),
		}).Error("獲取角色失敗")
		
		ErrorResponse(c, http.StatusNotFound, "找不到角色", "CHARACTER_NOT_FOUND")
		return
	}

	styleType := c.Query("style_type") // 可選篩選
	
	styles := character.Behavior.SpeechStyles
	if styleType != "" {
		// 篩選指定類型的風格
		filteredStyles := make([]models.CharacterSpeechStyle, 0)
		for _, style := range styles {
			if string(style.StyleType) == styleType {
				filteredStyles = append(filteredStyles, style)
			}
		}
		styles = filteredStyles
	}

	SuccessResponse(c, gin.H{
		"character_id": characterID,
		"style_type":   styleType,
		"styles":       styles,
		"count":        len(styles),
	}, "獲取對話風格成功")
}

// GetNSFWGuideline godoc
// @Summary      獲取NSFW指引
// @Description  獲取指定等級的NSFW內容指引
// @Tags         Character
// @Accept       json
// @Produce      json
// @Param        level path int true "NSFW等級 (1-5)" minimum(1) maximum(5)
// @Param        locale query string false "語言地區設定" default(zh-Hant)
// @Param        engine query string false "推薦引擎 (openai/grok)"
// @Success      200 {object} models.APIResponse{data=object} "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "指引不存在"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character/nsfw-guideline/{level} [get]
func GetNSFWGuideline(c *gin.Context) {
	levelStr := c.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 || level > 5 {
		ErrorResponse(c, http.StatusBadRequest, "無效的NSFW等級", "INVALID_NSFW_LEVEL")
		return
	}

	locale := c.DefaultQuery("locale", "zh-Hant")
	engine := c.DefaultQuery("engine", "")
	
	// 根據等級決定引擎
	if engine == "" {
		if level >= 4 {
			engine = "grok"
		} else {
			engine = "openai"
		}
	}

	guideline, err := getCharacterService().GetNSFWGuideline(c.Request.Context(), level, locale, engine)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"level":  level,
			"locale": locale,
			"engine": engine,
			"error":  err.Error(),
		}).Error("獲取NSFW指引失敗")
		
		ErrorResponse(c, http.StatusNotFound, "找不到NSFW指引", "NSFW_GUIDELINE_NOT_FOUND")
		return
	}

	SuccessResponse(c, guideline, "獲取NSFW指引成功")
}

// GetCharacterScenes godoc
// @Summary      獲取角色場景
// @Description  獲取指定角色的場景設定，可根據類型、時間、好感度等篩選
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "角色ID"
// @Param        scene_type query string false "場景類型 (default/romantic/intimate)"
// @Param        time_of_day query string false "時間段 (morning/afternoon/evening/night)"
// @Param        affection query int false "好感度 (0-100)" default(50)
// @Param        nsfw_level query int false "NSFW等級 (1-5)" default(1)
// @Success      200 {object} models.APIResponse{data=object} "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色場景不存在"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character/{id}/scenes [get]
func GetCharacterScenes(c *gin.Context) {
	characterID := c.Param("id")
	if characterID == "" {
		ErrorResponse(c, http.StatusBadRequest, "角色ID不能為空", "INVALID_CHARACTER_ID")
		return
	}

	sceneType := c.Query("scene_type")   // default, romantic, intimate
	timeOfDay := c.Query("time_of_day")  // morning, afternoon, evening, night
	
	affectionStr := c.DefaultQuery("affection", "50")
	affection, err := strconv.Atoi(affectionStr)
	if err != nil {
		affection = 50
	}

	nsfwLevelStr := c.DefaultQuery("nsfw_level", "1")
	nsfwLevel, err := strconv.Atoi(nsfwLevelStr)
	if err != nil {
		nsfwLevel = 1
	}

	scenes, err := getCharacterService().GetCharacterScenes(c.Request.Context(), characterID, sceneType, timeOfDay, affection, nsfwLevel)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"scene_type":  sceneType,
			"time_of_day": timeOfDay,
			"affection":   affection,
			"nsfw_level":  nsfwLevel,
			"error":       err.Error(),
		}).Error("獲取角色場景失敗")
		
		ErrorResponse(c, http.StatusNotFound, "找不到角色場景", "CHARACTER_SCENES_NOT_FOUND")
		return
	}

	SuccessResponse(c, gin.H{
		"character_id": characterID,
		"scene_type":   sceneType,
		"time_of_day":  timeOfDay,
		"affection":    affection,
		"nsfw_level":   nsfwLevel,
		"scenes":       scenes,
		"count":        len(scenes),
	}, "獲取角色場景成功")
}

// GetBestSpeechStyle godoc
// @Summary      獲取最適合的對話風格
// @Description  根據當前對話情境獲取最適合的角色對話風格
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "角色ID"
// @Param        nsfw_level query int false "NSFW等級 (1-5)" default(1)
// @Param        affection query int false "好感度 (0-100)" default(50)
// @Success      200 {object} models.APIResponse{data=object} "獲取成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "找不到適合的對話風格"
// @Failure      500 {object} models.APIResponse{error=models.APIError} "伺服器錯誤"
// @Router       /character/{id}/speech-styles/best [get]
func GetBestSpeechStyle(c *gin.Context) {
	characterID := c.Param("id")
	if characterID == "" {
		ErrorResponse(c, http.StatusBadRequest, "角色ID不能為空", "INVALID_CHARACTER_ID")
		return
	}

	nsfwLevelStr := c.DefaultQuery("nsfw_level", "1")
	nsfwLevel, err := strconv.Atoi(nsfwLevelStr)
	if err != nil {
		nsfwLevel = 1
	}

	affectionStr := c.DefaultQuery("affection", "50")
	affection, err := strconv.Atoi(affectionStr)
	if err != nil {
		affection = 50
	}

	// 構建簡單的對話上下文
	context := &services.ConversationContext{
		EmotionState: &services.EmotionState{
			Affection: affection,
		},
	}

	style, err := getCharacterService().GetBestSpeechStyle(c.Request.Context(), characterID, nsfwLevel, context)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"character_id": characterID,
			"nsfw_level":  nsfwLevel,
			"affection":   affection,
			"error":       err.Error(),
		}).Error("獲取最適合對話風格失敗")
		
		ErrorResponse(c, http.StatusNotFound, "找不到適合的對話風格", "BEST_SPEECH_STYLE_NOT_FOUND")
		return
	}

	SuccessResponse(c, gin.H{
		"character_id": characterID,
		"nsfw_level":  nsfwLevel,
		"affection":   affection,
		"style":       style,
	}, "獲取最適合對話風格成功")
}

// CharacterHealthCheck 角色服務健康檢查
func CharacterHealthCheck(c *gin.Context) {
	isHealthy := getCharacterService().HealthCheck(c.Request.Context())
	
	status := "healthy"
	if !isHealthy {
		status = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"timestamp": utils.Now(),
	})
}