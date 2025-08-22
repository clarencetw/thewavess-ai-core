package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

var (
	ttsService            *services.TTSService
	characterServiceForTTS *services.CharacterService
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

func getCharacterServiceForTTS() *services.CharacterService {
	if characterServiceForTTS == nil {
		characterServiceForTTS = services.GetCharacterService()
	}
	return characterServiceForTTS
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
			"audio_data":    base64.StdEncoding.EncodeToString(response.AudioData), // Base64編碼的音頻數據
			"audio_url":     fmt.Sprintf("data:audio/%s;base64,%s", response.Format, base64.StdEncoding.EncodeToString(response.AudioData)), // 可直接播放的 Data URL
			"duration":      response.Duration,
			"file_size":     formatFileSize(response.Size),
			"format":        response.Format,
			"sample_rate":   "44100Hz",
			"voice_name":    getVoiceNameByID(c.Request.Context(), req.Voice),
		},
		"processing": gin.H{
			"started_at":   time.Now(),
			"completed_at": time.Now(),
			"queue_time":   "0.1s",
			"render_time":  "1.2s",
		},
		"usage": gin.H{
			"characters_processed": len(req.Text),
			"tokens_used":         len(req.Text) / 4, // 估算token使用
			"cost":               "$0.002",
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
// @Param        character_id query string false "角色ID過濾"
// @Param        language query string false "語言過濾"
// @Success      200 {object} models.APIResponse "獲取成功"
// @Router       /tts/voices [get]
func GetVoiceList(c *gin.Context) {
	characterID := c.Query("character_id")
	language := c.Query("language")

	// 從 TTS 服務獲取語音列表
	availableVoices := getTTSService().GetAvailableVoices()
	
	// 從資料庫獲取所有角色信息
	characters, err := getCharacterServiceForTTS().GetActiveCharacters(c.Request.Context())
	if err != nil {
		utils.Logger.WithError(err).Error("獲取角色列表失敗")
	}

	// 擴展語音信息
	voices := []gin.H{}
	for _, voice := range availableVoices {
		// 尋找對應角色
		var character *models.Character
		for _, char := range characters {
			if char.ID == voice["character_id"] {
				character = char
				break
			}
		}

		voiceInfo := gin.H{
			"voice_id":     voice["voice_id"],
			"name":         voice["name"],
			"character_id": voice["character_id"],
			"character":    voice["character"],
			"openai_voice": voice["openai_voice"],
			"description":  voice["description"],
			"language":     "zh-TW",
			"gender":       "male",
			"age_range":    "25-35",
			"quality":      "premium",
			"is_default":   voice["voice_id"] == "voice_001",
		}

		// 如果找到角色，使用資料庫中的信息
		if character != nil {
			voiceInfo["personality"] = character.Metadata.Tags // 使用標籤作為性格特點
			voiceInfo["emotions"] = getEmotionsByCharacterType(string(character.Type)) // 根據角色類型獲取情感
			voiceInfo["character"] = character.Name
		} else {
			// 如果找不到角色，使用預設值
			voiceInfo["personality"] = []string{"溫和", "成熟"}
			voiceInfo["emotions"] = []string{"neutral", "gentle"}
		}
		voices = append(voices, voiceInfo)
	}

	// 過濾邏輯
	filteredVoices := []gin.H{}
	for _, voice := range voices {
		include := true
		
		if characterID != "" && voice["character_id"] != characterID {
			include = false
		}
		
		if language != "" && voice["language"] != language {
			include = false
		}
		
		if include {
			filteredVoices = append(filteredVoices, voice)
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取語音列表成功",
		Data: gin.H{
			"voices":      filteredVoices,
			"total_count": len(filteredVoices),
			"filters": gin.H{
				"character_id": characterID,
				"language":     language,
			},
			"categories": generateVoiceCategories(filteredVoices, characters),
		},
	})
}




// 輔助函數

// getCharacterByVoice 根據語音ID獲取角色名稱
func getCharacterByVoice(ctx context.Context, voiceID, characterID string) string {
	// 如果有直接提供角色ID，先嘗試獲取
	if characterID != "" {
		character, err := getCharacterServiceForTTS().GetCharacter(ctx, characterID)
		if err == nil && character != nil {
			return character.Name
		}
	}

	// 從 TTS 服務獲取語音對應關係
	availableVoices := getTTSService().GetAvailableVoices()
	for _, voice := range availableVoices {
		if voice["voice_id"] == voiceID {
			if charID, ok := voice["character_id"].(string); ok && charID != "" {
				character, err := getCharacterServiceForTTS().GetCharacter(ctx, charID)
				if err == nil && character != nil {
					return character.Name
				}
			}
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

// getEmotionsByCharacterType 根據角色類型獲取支持的情感
func getEmotionsByCharacterType(characterType string) []string {
	switch strings.ToLower(characterType) {
	case "dominant":
		return []string{"neutral", "romantic", "serious", "gentle", "dominant", "passionate"}
	case "gentle":
		return []string{"gentle", "wise", "caring", "warm", "understanding"}
	case "playful":
		return []string{"cheerful", "playful", "excited", "warm", "energetic"}
	default:
		return []string{"neutral", "gentle"} // 默認情感
	}
}

// generateVoiceCategories 產生語音分類統計
func generateVoiceCategories(voices []gin.H, characters []*models.Character) gin.H {
	characterCount := make(map[string]int)
	qualityCount := make(map[string]int)

	for _, voice := range voices {
		// 統計角色分類
		if charName, ok := voice["character"].(string); ok {
			characterCount[charName]++
		}

		// 統計品質分類
		if quality, ok := voice["quality"].(string); ok {
			qualityCount[quality]++
		}
	}

	return gin.H{
		"by_character": characterCount,
		"by_quality":   qualityCount,
	}
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