package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
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
	// TODO: 實作角色列表邏輯
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success: false,
		Message: "功能尚未實作",
	})
}

// GetCharacterDetails godoc
// @Summary      獲取角色詳細資訊
// @Description  獲取指定角色的詳細資訊
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        character_id path string true "角色 ID"
// @Success      200 {object} models.APIResponse{data=models.Character} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Router       /character/{character_id} [get]
func GetCharacterDetails(c *gin.Context) {
	// TODO: 實作角色詳細資訊邏輯
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success: false,
		Message: "功能尚未實作",
	})
}

// GetCharacterStats godoc
// @Summary      獲取角色統計數據
// @Description  獲取角色的使用統計和評分數據
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        character_id path string true "角色 ID"
// @Success      200 {object} models.APIResponse{data=models.CharacterStats} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Router       /character/{character_id}/stats [get]
func GetCharacterStats(c *gin.Context) {
	// TODO: 實作角色統計數據邏輯
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success: false,
		Message: "功能尚未實作",
	})
}

// GetCurrentCharacter godoc
// @Summary      獲取當前選擇角色
// @Description  獲取用戶當前選擇的角色
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} models.APIResponse{data=models.Character} "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "未選擇角色"
// @Router       /user/character [get]
func GetCurrentCharacter(c *gin.Context) {
	// TODO: 實作獲取當前角色邏輯
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success: false,
		Message: "功能尚未實作",
	})
}

// SelectCharacter godoc
// @Summary      選擇當前角色
// @Description  設定用戶當前使用的角色
// @Tags         Character
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.SelectCharacterRequest true "角色選擇"
// @Success      200 {object} models.APIResponse{data=models.Character} "選擇成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求參數錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Failure      404 {object} models.APIResponse{error=models.APIError} "角色不存在"
// @Router       /user/character [put]
func SelectCharacter(c *gin.Context) {
	// TODO: 實作選擇角色邏輯
	c.JSON(http.StatusNotImplemented, models.APIResponse{
		Success: false,
		Message: "功能尚未實作",
	})
}