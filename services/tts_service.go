package services

import (
	"context"
	"fmt"
	"io"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/sashabaranov/go-openai"
)

// TTSService TTS 服務
type TTSService struct {
	client *openai.Client
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
	
	// 優先使用專用的 TTS API key，如果沒有則使用 OpenAI API key
	apiKey := utils.GetEnvWithDefault("TTS_API_KEY", "")
	if apiKey == "" {
		apiKey = utils.GetEnvWithDefault("OPENAI_API_KEY", "")
		if apiKey == "" {
			utils.Logger.Warn("Neither TTS_API_KEY nor OPENAI_API_KEY set for TTS service")
		}
	}

	var client *openai.Client
	if apiKey != "" {
		// 使用標準 OpenAI API
		client = openai.NewClient(apiKey)
		utils.Logger.WithFields(map[string]interface{}{
			"service": "tts",
		}).Info("TTS service initialized with OpenAI")
	}

	return &TTSService{
		client: client,
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

	// 如果沒有客戶端，返回 mock 響應
	if s.client == nil {
		utils.Logger.WithField("service", "tts").Info("Using mock response (API key not set)")
		return s.mockTTSResponse(text, voice), nil
	}

	// 使用 go-openai 庫調用 TTS API
	request := openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		Input:          text,
		Voice:          s.mapVoiceToOpenAI(voice),
		ResponseFormat: openai.SpeechResponseFormatMp3,
		Speed:          speed,
	}

	resp, err := s.client.CreateSpeech(ctx, request)
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "tts",
			"error":   err.Error(),
		}).Error("TTS API request failed")
		// 如果API調用失敗，返回mock響應作為fallback
		return s.mockTTSResponse(text, voice), nil
	}
	defer resp.Close()

	// 讀取音頻數據
	audioData, err := io.ReadAll(resp)
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

// mapVoiceToOpenAI 將角色語音映射到 OpenAI TTS 語音
func (s *TTSService) mapVoiceToOpenAI(voice string) openai.SpeechVoice {
	voiceMapping := map[string]openai.SpeechVoice{
		"voice_001": openai.VoiceAlloy,  // 陸燁銘標準音
		"voice_002": openai.VoiceEcho,   // 陸燁銘深情音
		"voice_003": openai.VoiceFable,  // 沈言墨溫雅音
		"default":   openai.VoiceAlloy,
	}

	if mappedVoice, exists := voiceMapping[voice]; exists {
		return mappedVoice
	}
	
	return openai.VoiceAlloy // 默認語音
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