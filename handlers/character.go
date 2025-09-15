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
// @Description  支援分頁、篩選和排序的角色列表
// @Tags         Character
// @Accept       json
// @Produce      json
// @Param        page query int false "頁數" default(1)
// @Param        page_size query int false "每頁數量" default(20)
// @Param        type query string false "角色類型 (dominant/gentle/playful/mystery/reliable)"
// @Param        is_active query bool false "是否啟用"
// @Param        tags query []string false "標籤篩選"
// @Param        search query string false "搜尋角色名稱或描述"
// @Param        sort_by query string false "排序欄位" default(created_at)
// @Param        sort_order query string false "排序方向 (asc/desc)" default(desc)
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
// @Description  根據ID獲取角色詳細信息
// @Tags         Character
// @Accept       json
// @Produce      json
// @Param        id path string true "角色ID"
// @Success      200 {object} models.APIResponse{data=models.CharacterResponse} "獲取成功"
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

	// 返回角色基本資訊
	SuccessResponse(c, character, "獲取角色詳情成功")
}

// CreateCharacter godoc
// @Summary      創建角色
// @Description  創建新的AI角色。參數說明：name(角色名稱，1-50個字符，必填)，type(角色類型，必填：dominant=霸道型、gentle=溫柔型、playful=活潑型、mystery=神秘型、reliable=可靠型)，locale(語言區域，固定zh-TW繁體中文，必填)，user_description(用戶自由描述角色的詳細內容，選填)，metadata(角色元數據，選填：avatar_url=頭像圖片URL、tags=角色標籤陣列如[霸總,腹黑,現代]、popularity=人氣值0-100)
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        character body models.CharacterCreateRequest true "角色創建請求：name(角色名稱，必填)，type(角色類型，必填)，locale(語言區域zh-TW，必填)，user_description(角色描述，選填)，metadata(元數據：avatar_url頭像URL、tags標籤陣列、popularity人氣值0-100，均為選填)"
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

	// 獲取當前用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "需要用戶認證",
			},
		})
		return
	}

	character, err := getCharacterService().CreateCharacterWithUser(c.Request.Context(), &req, userID.(string))
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

	// 獲取當前用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "需要用戶認證",
			},
		})
		return
	}

	character, err := getCharacterService().UpdateCharacterWithUser(c.Request.Context(), characterID, &req, userID.(string))
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

	// 獲取當前用戶ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "需要用戶認證",
			},
		})
		return
	}

	err := getCharacterService().SoftDeleteCharacterWithUser(c.Request.Context(), characterID, userID.(string))
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

// Character Configuration Handlers

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
			"error":        err.Error(),
		}).Error("獲取角色失敗")

		ErrorResponse(c, http.StatusNotFound, "找不到角色", "CHARACTER_NOT_FOUND")
		return
	}

	SuccessResponse(c, character, "獲取角色成功")
}
