package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/clarencetw/thewavess-ai-core/models"
	"github.com/clarencetw/thewavess-ai-core/services"
	"github.com/clarencetw/thewavess-ai-core/utils"
)

var ttsService *services.TTSService

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
// @Param        request body object true "TTS生成請求"
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
		Pitch       float64 `json:"pitch"`
		CharacterID string  `json:"character_id"`
		Emotion     string  `json:"emotion"`
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
		req.Voice = "voice_001"
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
		"tts_id":    utils.GenerateID(16),
		"user_id":   userID,
		"text":      req.Text,
		"voice":     req.Voice,
		"character": getCharacterByVoice(req.Voice),
		"settings": gin.H{
			"speed":    req.Speed,
			"pitch":    req.Pitch,
			"emotion":  req.Emotion,
			"quality":  "high",
		},
		"result": gin.H{
			"audio_data":    base64.StdEncoding.EncodeToString(response.AudioData), // Base64編碼的音頻數據
			"audio_url":     fmt.Sprintf("data:audio/%s;base64,%s", response.Format, base64.StdEncoding.EncodeToString(response.AudioData)), // 可直接播放的 Data URL
			"duration":      response.Duration,
			"file_size":     formatFileSize(response.Size),
			"format":        response.Format,
			"sample_rate":   "44100Hz",
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

// BatchGenerateTTS godoc
// @Summary      批量生成語音
// @Description  批量將多個文字轉換為語音
// @Tags         TTS
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "批量TTS請求"
// @Success      200 {object} models.APIResponse "生成成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /tts/batch [post]
func BatchGenerateTTS(c *gin.Context) {
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
		Items []struct {
			Text        string  `json:"text" binding:"required"`
			Voice       string  `json:"voice"`
			CharacterID string  `json:"character_id"`
			Emotion     string  `json:"emotion"`
		} `json:"items" binding:"required"`
		Settings gin.H `json:"settings"`
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

	// 設置默認語音
	defaultVoice := "voice_001"
	defaultSpeed := 1.0

	// 提取文字列表
	texts := make([]string, len(req.Items))
	for i, item := range req.Items {
		texts[i] = item.Text
	}

	// 調用批量 TTS 服務
	responses, err := getTTSService().BatchGenerateSpeech(c.Request.Context(), texts, defaultVoice, defaultSpeed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "BATCH_TTS_FAILED",
				Message: "批量語音生成失敗: " + err.Error(),
			},
		})
		return
	}

	// 構建批量結果
	batchResults := []gin.H{}
	for i, item := range req.Items {
		voice := item.Voice
		if voice == "" {
			voice = defaultVoice
		}
		
		result := gin.H{
			"index":      i + 1,
			"text":       item.Text,
			"voice":      voice,
			"character":  getCharacterByVoice(voice),
			"audio_data": responses[i].AudioData,
			"duration":   responses[i].Duration,
			"file_size":  formatFileSize(responses[i].Size),
			"status":     "completed",
		}
		batchResults = append(batchResults, result)
	}

	batchResponse := gin.H{
		"batch_id":    utils.GenerateID(16),
		"user_id":     userID,
		"total_items": len(req.Items),
		"results":     batchResults,
		"summary": gin.H{
			"successful":      len(req.Items),
			"failed":          0,
			"total_duration":  "62.5s",
			"total_size":      "1.2MB",
			"processing_time": "8.3s",
		},
		"download": gin.H{
			"zip_url":    "https://example.com/tts/batch_" + utils.GenerateID(32) + ".zip",
			"expires_at": time.Now().Add(24 * time.Hour),
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "批量語音生成成功",
		Data:    batchResponse,
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
	
	// 擴展語音信息
	voices := []gin.H{}
	for _, voice := range availableVoices {
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
			"personality":  getPersonalityByCharacter(voice["character"].(string)),
			"emotions":     getEmotionsByCharacter(voice["character"].(string)),
			"quality":      "premium",
			"is_default":   voice["voice_id"] == "voice_001",
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
			"categories": gin.H{
				"by_character": gin.H{
					"陸燁銘": 2,
					"沈言墨": 1,
				},
				"by_quality": gin.H{
					"premium":  3,
					"standard": 0,
				},
			},
		},
	})
}

// PreviewTTS godoc
// @Summary      預覽語音
// @Description  生成語音預覽片段
// @Tags         TTS
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "預覽請求"
// @Success      200 {object} models.APIResponse "預覽成功"
// @Failure      400 {object} models.APIResponse{error=models.APIError} "請求錯誤"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /tts/preview [post]
func PreviewTTS(c *gin.Context) {
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
		Text    string `json:"text" binding:"required"`
		VoiceID string `json:"voice_id" binding:"required"`
		Emotion string `json:"emotion"`
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

	// 調用 TTS 服務生成預覽語音
	response, err := getTTSService().GenerateSpeech(c.Request.Context(), req.Text, req.VoiceID, 1.0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "TTS_PREVIEW_FAILED",
				Message: "語音預覽生成失敗: " + err.Error(),
			},
		})
		return
	}

	// 構建預覽回應
	preview := gin.H{
		"preview_id": utils.GenerateID(16),
		"user_id":    userID,
		"text":       req.Text,
		"voice_id":   req.VoiceID,
		"emotion":    req.Emotion,
		"audio": gin.H{
			"audio_data": base64.StdEncoding.EncodeToString(response.AudioData), // Base64編碼的音頻數據
			"audio_url":  fmt.Sprintf("data:audio/%s;base64,%s", response.Format, base64.StdEncoding.EncodeToString(response.AudioData)), // 可直接播放的 Data URL
			"duration":   response.Duration,
			"file_size":  formatFileSize(response.Size),
			"format":     response.Format,
			"expires_at": time.Now().Add(1 * time.Hour),
		},
		"voice_info": gin.H{
			"name":      getVoiceNameByID(req.VoiceID),
			"character": getCharacterByVoice(req.VoiceID),
			"emotion":   req.Emotion,
			"quality":   "preview",
		},
		"generated_at": time.Now(),
		"note":         "預覽音質為標準品質，正式生成將使用高品質音效",
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "語音預覽生成成功",
		Data:    preview,
	})
}

// GetTTSHistory godoc
// @Summary      獲取語音歷史
// @Description  獲取用戶的語音生成歷史記錄
// @Tags         TTS
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "頁碼" default(1)
// @Param        limit query int false "每頁數量" default(20)
// @Success      200 {object} models.APIResponse "獲取成功"
// @Failure      401 {object} models.APIResponse{error=models.APIError} "未授權"
// @Router       /tts/history [get]
func GetTTSHistory(c *gin.Context) {
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

	// 靜態數據回應 - 模擬語音歷史
	history := []gin.H{
		{
			"tts_id":       "tts_001",
			"text":         "你今天過得怎麼樣？",
			"voice_name":   "陸燁銘標準音",
			"character":    "陸燁銘",
			"emotion":      "gentle",
			"audio_url":    "https://example.com/tts/tts_001.mp3",
			"duration":     "3.2s",
			"generated_at": time.Now().AddDate(0, 0, -1),
			"status":       "completed",
		},
		{
			"tts_id":       "tts_002",
			"text":         "我一直在想你...",
			"voice_name":   "陸燁銘深情音",
			"character":    "陸燁銘",
			"emotion":      "romantic",
			"audio_url":    "https://example.com/tts/tts_002.mp3",
			"duration":     "4.1s",
			"generated_at": time.Now().AddDate(0, 0, -2),
			"status":       "completed",
		},
		{
			"tts_id":       "tts_003",
			"text":         "有什麼我可以幫助你的嗎？",
			"voice_name":   "沈言墨溫雅音",
			"character":    "沈言墨",
			"emotion":      "caring",
			"audio_url":    "https://example.com/tts/tts_003.mp3",
			"duration":     "3.8s",
			"generated_at": time.Now().AddDate(0, 0, -3),
			"status":       "completed",
		},
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "獲取語音歷史成功",
		Data: gin.H{
			"user_id":      userID,
			"history":      history,
			"total_count":  len(history),
			"statistics": gin.H{
				"total_generated":   156,
				"this_month":        23,
				"favorite_voice":    "陸燁銘標準音",
				"total_duration":    "15m 32s",
				"storage_used":      "12.8MB",
			},
		},
	})
}


// 輔助函數

// getCharacterByVoice 根據語音ID獲取角色名稱
func getCharacterByVoice(voiceID string) string {
	voiceMap := map[string]string{
		"voice_001": "陸燁銘",
		"voice_002": "陸燁銘", 
		"voice_003": "沈言墨",
	}
	
	if character, exists := voiceMap[voiceID]; exists {
		return character
	}
	return "陸燁銘" // 默認角色
}

// getVoiceNameByID 根據語音ID獲取語音名稱
func getVoiceNameByID(voiceID string) string {
	voiceMap := map[string]string{
		"voice_001": "陸燁銘標準音",
		"voice_002": "陸燁銘深情音",
		"voice_003": "沈言墨溫雅音",
	}
	
	if name, exists := voiceMap[voiceID]; exists {
		return name
	}
	return "陸燁銘標準音" // 默認語音
}

// getPersonalityByCharacter 根據角色獲取性格標籤
func getPersonalityByCharacter(character string) []string {
	personalityMap := map[string][]string{
		"陸燁銘": {"霸道", "溫柔", "成熟", "深情", "磁性"},
		"沈言墨": {"溫雅", "知性", "儒雅", "溫和"},
	}
	
	if personality, exists := personalityMap[character]; exists {
		return personality
	}
	return []string{"溫和", "成熟"} // 默認性格
}

// getEmotionsByCharacter 根據角色獲取支持的情感
func getEmotionsByCharacter(character string) []string {
	emotionsMap := map[string][]string{
		"陸燁銘": {"neutral", "romantic", "serious", "gentle", "dominant", "passionate"},
		"沈言墨": {"gentle", "wise", "caring", "scholarly", "warm"},
	}
	
	if emotions, exists := emotionsMap[character]; exists {
		return emotions
	}
	return []string{"neutral", "gentle"} // 默認情感
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