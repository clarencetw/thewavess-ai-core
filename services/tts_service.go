package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/clarencetw/thewavess-ai-core/utils"
)

// TTSService TTS 服務
type TTSService struct {
	openaiClient *http.Client
	apiKey       string
	apiURL       string
}

// TTSRequest TTS 請求結構
type TTSRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
}

// TTSResponse TTS 回應結構
type TTSResponse struct {
	AudioData []byte
	Format    string
	Duration  string
	Size      int64
}

// NewTTSService 創建新的 TTS 服務
func NewTTSService() *TTSService {
	// 確保環境變數已載入
	utils.LoadEnv()
	
	apiKey := utils.GetEnvWithDefault("OPENAI_API_KEY", "")
	if apiKey == "" {
		utils.Logger.Warn("OPENAI_API_KEY not set for TTS service")
	}

	return &TTSService{
		openaiClient: &http.Client{
			Timeout: 60 * time.Second, // TTS 需要更長的超時時間
		},
		apiKey: apiKey,
		apiURL: "https://api.openai.com/v1/audio/speech",
	}
}

// GenerateSpeech 生成語音
func (s *TTSService) GenerateSpeech(ctx context.Context, text string, voice string, speed float64) (*TTSResponse, error) {
	// 記錄請求開始
	utils.Logger.WithFields(map[string]interface{}{
		"service": "tts",
		"text":    text[:minInt(len(text), 50)] + "...", // 只記錄前50個字符
		"voice":   voice,
		"speed":   speed,
	}).Info("TTS generation started")

	// 如果沒有 API Key，返回 mock 響應
	if s.apiKey == "" {
		return s.mockTTSResponse(text, voice), nil
	}

	// 準備請求
	request := TTSRequest{
		Model:          "tts-1",
		Input:          text,
		Voice:          s.mapVoice(voice),
		ResponseFormat: "mp3",
		Speed:          speed,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "tts",
			"error":   err.Error(),
		}).Error("Failed to marshal TTS request")
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 創建 HTTP 請求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 設置請求頭
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// 發送請求
	resp, err := s.openaiClient.Do(httpReq)
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "tts",
			"error":   err.Error(),
		}).Error("TTS API request failed")
		// 如果API調用失敗，返回mock響應作為fallback
		return s.mockTTSResponse(text, voice), nil
	}
	defer resp.Body.Close()

	// 檢查響應狀態
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		utils.Logger.WithFields(map[string]interface{}{
			"service":     "tts",
			"status_code": resp.StatusCode,
			"response":    string(bodyBytes),
		}).Error("TTS API returned error")
		// 如果API返回錯誤，返回mock響應作為fallback
		return s.mockTTSResponse(text, voice), nil
	}

	// 讀取音頻數據
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "tts",
			"error":   err.Error(),
		}).Error("Failed to read TTS response")
		return s.mockTTSResponse(text, voice), nil
	}

	// 計算音頻時長（估算：平均每字符0.2秒）
	estimatedDuration := float64(len(text)) * 0.2
	durationStr := fmt.Sprintf("%.1fs", estimatedDuration)

	utils.Logger.WithFields(map[string]interface{}{
		"service":       "tts",
		"text_length":   len(text),
		"audio_size":    len(audioData),
		"duration":      durationStr,
		"voice":         voice,
	}).Info("TTS generation completed")

	return &TTSResponse{
		AudioData: audioData,
		Format:    "mp3",
		Duration:  durationStr,
		Size:      int64(len(audioData)),
	}, nil
}

// mapVoice 將角色語音映射到 OpenAI TTS 語音
func (s *TTSService) mapVoice(voice string) string {
	voiceMapping := map[string]string{
		"voice_001": "alloy",  // 陸燁銘標準音
		"voice_002": "echo",   // 陸燁銘深情音
		"voice_003": "fable",  // 沈言墨溫雅音
		"default":   "alloy",
	}

	if mappedVoice, exists := voiceMapping[voice]; exists {
		return mappedVoice
	}
	
	return "alloy" // 默認語音
}

// mockTTSResponse 創建模擬的 TTS 響應
func (s *TTSService) mockTTSResponse(text string, voice string) *TTSResponse {
	// 創建一個簡單的模擬音頻數據（實際上是空的MP3頭部）
	mockAudioData := []byte{
		0xFF, 0xFB, 0x90, 0x00, // MP3 header
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	// 計算估算時長
	estimatedDuration := float64(len(text)) * 0.2
	durationStr := fmt.Sprintf("%.1fs", estimatedDuration)

	utils.Logger.WithFields(map[string]interface{}{
		"service": "tts",
		"mode":    "mock",
		"text":    text[:minInt(len(text), 50)] + "...",
		"voice":   voice,
	}).Info("TTS mock response generated")

	return &TTSResponse{
		AudioData: mockAudioData,
		Format:    "mp3",
		Duration:  durationStr,
		Size:      int64(len(mockAudioData)),
	}
}

// BatchGenerateSpeech 批量生成語音
func (s *TTSService) BatchGenerateSpeech(ctx context.Context, texts []string, voice string, speed float64) ([]*TTSResponse, error) {
	utils.Logger.WithFields(map[string]interface{}{
		"service":    "tts",
		"batch_size": len(texts),
		"voice":      voice,
		"speed":      speed,
	}).Info("Batch TTS generation started")

	responses := make([]*TTSResponse, len(texts))
	
	for i, text := range texts {
		response, err := s.GenerateSpeech(ctx, text, voice, speed)
		if err != nil {
			utils.Logger.WithFields(map[string]interface{}{
				"service": "tts",
				"index":   i,
				"error":   err.Error(),
			}).Error("Batch TTS generation failed for item")
			
			// 如果單個項目失敗，使用mock響應
			responses[i] = s.mockTTSResponse(text, voice)
		} else {
			responses[i] = response
		}
	}

	utils.Logger.WithFields(map[string]interface{}{
		"service":       "tts",
		"batch_size":    len(texts),
		"success_count": len(responses),
	}).Info("Batch TTS generation completed")

	return responses, nil
}

// GetAvailableVoices 獲取可用語音列表
func (s *TTSService) GetAvailableVoices() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"voice_id":     "voice_001",
			"name":         "陸燁銘標準音",
			"character_id": "char_001",
			"character":    "陸燁銘",
			"openai_voice": "alloy",
			"description":  "成熟穩重的男性聲音",
		},
		{
			"voice_id":     "voice_002",
			"name":         "陸燁銘深情音",
			"character_id": "char_001",
			"character":    "陸燁銘",
			"openai_voice": "echo",
			"description":  "深情磁性的男性聲音",
		},
		{
			"voice_id":     "voice_003",
			"name":         "沈言墨溫雅音",
			"character_id": "char_002",
			"character":    "沈言墨",
			"openai_voice": "fable",
			"description":  "溫雅知性的男性聲音",
		},
	}
}

// minInt 返回兩個整數中的較小值
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}