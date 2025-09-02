package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/gin-gonic/gin"
)

var (
	ttsService *services.TTSService
)

// GenerateTTSRequest TTS生成請求結構
type GenerateTTSRequest struct {
	Text        string  `json:"text" binding:"required" example:"你好，今天過得怎麼樣？"`
	Voice       string  `json:"voice" example:"voice_001"`
	Speed       float64 `json:"speed" example:"1.0"`
	CharacterID string  `json:"character_id" example:"character_01"`
}

func getTTSService() *services.TTSService {
	if ttsService == nil {
		ttsService = services.NewTTSService()
	}
	return ttsService
}

// GenerateTTS godoc
// @Summary      生成語音
// @Description  將文字轉換為語音
// @Tags         TTS
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body GenerateTTSRequest true "TTS生成請求"
// @Success      200 {object} models.APIResponse "生成成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /tts/generate [post]
func GenerateTTS(c *gin.Context) {
	// 檢查認證
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "UNAUTHORIZED",
				Message: "未授權訪問",
			},
		})
		return
	}

	var req struct {
		Text        string  `json:"text" binding:"required"`
		Voice       string  `json:"voice"`
		Speed       float64 `json:"speed"`
		CharacterID string  `json:"character_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_INPUT",
				Message: "請求參數錯誤: " + err.Error(),
			},
		})
		return
	}

	// 設置默認值
	if req.Speed == 0 {
		req.Speed = 1.0
	}
	if req.Voice == "" {
		// 如果有角色ID，嘗試獲取預設聲音
		if req.CharacterID != "" {
			req.Voice = getDefaultVoiceForCharacter(c.Request.Context(), req.CharacterID)
		} else {
			req.Voice = "voice_001"
		}
	}

	// 調用 TTS 服務生成語音
	response, err := getTTSService().GenerateSpeech(c.Request.Context(), req.Text, req.Voice, req.Speed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TTS_GENERATION_FAILED",
				Message: "語音生成失敗: " + err.Error(),
			},
		})
		return
	}

	// 構建回應數據
	ttsResult := gin.H{
		"tts_id":    utils.GenerateTTSID(),
		"user_id":   userID,
		"text":      req.Text,
		"voice":     req.Voice,
		"character": getCharacterByVoice(c.Request.Context(), req.Voice, req.CharacterID),
		"settings": gin.H{
			"speed":   req.Speed,
			"quality": "high",
		},
		"result": gin.H{
			"audio_data":  base64.StdEncoding.EncodeToString(response.AudioData),                                                          // Base64編碼的音頻數據
			"audio_url":   fmt.Sprintf("data:audio/%s;base64,%s", response.Format, base64.StdEncoding.EncodeToString(response.AudioData)), // 可直接播放的 Data URL
			"duration":    response.Duration,
			"file_size":   formatFileSize(response.Size),
			"format":      response.Format,
			"sample_rate": "44100Hz",
			"voice_name":  getVoiceNameByID(c.Request.Context(), req.Voice),
		},
		"processing": gin.H{
			"started_at":   time.Now(),
			"completed_at": time.Now(),
			"queue_time":   "0.1s",
			"render_time":  "1.2s",
		},
		"usage": gin.H{
			"characters_processed": len(req.Text),
			"tokens_used":          len(req.Text) / 4, // 估算token使用
			"cost":                 "$0.002",
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "語音生成成功",
		Data:    ttsResult,
	})
}

// GetVoiceList godoc
// @Summary      獲取語音列表
// @Description  獲取可用的語音選項列表
// @Tags         TTS
// @Accept       json
// @Produce      json
// @Success      200 {object} models.APIResponse "獲取成功"
// @Router       /tts/voices [get]
func GetVoiceList(c *gin.Context) {
	// 從 TTS 服務獲取基礎語音列表
	availableVoices := getTTSService().GetAvailableVoices()

	// 處理語音信息，保留核心欄位
	voices := []gin.H{}
	for _, voice := range availableVoices {
		voiceInfo := gin.H{
			"voice_id":    voice["voice_id"],
			"name":        voice["name"],
			"character":   voice["character"],
			"description": voice["description"],
			"is_default":  voice["voice_id"] == "voice_001",
		}
		voices = append(voices, voiceInfo)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取語音列表成功",
		Data: gin.H{
			"voices": voices,
			"total":  len(voices),
		},
	})
}

// 輔助函數

// getCharacterByVoice 根據語音ID獲取角色名稱
func getCharacterByVoice(ctx context.Context, voiceID, characterID string) string {
	// 從 TTS 服務的數據獲取角色名稱
	availableVoices := getTTSService().GetAvailableVoices()
	for _, voice := range availableVoices {
		if voice["voice_id"] == voiceID {
			if charName, ok := voice["character"].(string); ok {
				return charName
			}
		}
	}

	return "沈宸" // 默認角色
}

// getVoiceNameByID 根據語音ID獲取語音名稱
func getVoiceNameByID(ctx context.Context, voiceID string) string {
	// 從 TTS 服務獲取語音信息
	availableVoices := getTTSService().GetAvailableVoices()
	for _, voice := range availableVoices {
		if voice["voice_id"] == voiceID {
			if name, ok := voice["name"].(string); ok {
				return name
			}
		}
	}

	return "沈宸標準音" // 默認語音
}

// getDefaultVoiceForCharacter 根據角色ID獲取預設語音
func getDefaultVoiceForCharacter(ctx context.Context, characterID string) string {
	// 從 TTS 服務獲取語音列表，尋找第一個匹配的語音
	availableVoices := getTTSService().GetAvailableVoices()
	for _, voice := range availableVoices {
		if voice["character_id"] == characterID {
			if voiceID, ok := voice["voice_id"].(string); ok {
				return voiceID
			}
		}
	}

	return "voice_001" // 默認語音
}

// formatFileSize 格式化文件大小
func formatFileSize(bytes int64) string {
	if bytes < 1024 {
		return strconv.FormatInt(bytes, 10) + "B"
	} else if bytes < 1024*1024 {
		return strconv.FormatInt(bytes/1024, 10) + "KB"
	} else {
		return strconv.FormatInt(bytes/(1024*1024), 10) + "MB"
	}
}
