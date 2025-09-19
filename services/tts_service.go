package services

import (
	"context"
	"fmt"
	"io"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// TTSService TTS 服務
type TTSService struct {
	client openai.Client
	hasAPIKey bool
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

	var client openai.Client
	if apiKey != "" {
		// 使用標準 OpenAI API
		client = openai.NewClient(
			option.WithAPIKey(apiKey),
		)
		utils.Logger.WithFields(map[string]interface{}{
			"service": "tts",
		}).Info("TTS service initialized with OpenAI")
	}

	return &TTSService{
		client: client,
		hasAPIKey: apiKey != "",
	}
}

// GenerateSpeech 生成語音
func (s *TTSService) GenerateSpeech(ctx context.Context, text string, voice string, speed float64) (*TTSResponse, error) {
	// 記錄請求開始
	utils.Logger.WithFields(map[string]interface{}{
		"service": "tts",
		"text":    text[:utils.Min(len(text), 50)] + "...", // 只記錄前50個字符
		"voice":   voice,
		"speed":   speed,
	}).Info("TTS generation started")

	// 如果沒有 API key，返回 mock 響應
	if !s.hasAPIKey {
		utils.Logger.WithField("service", "tts").Info("Using mock response (API key not set)")
		return s.mockTTSResponse(text, voice), nil
	}

	// 使用官方 OpenAI Go SDK 調用 TTS API
	resp, err := s.client.Audio.Speech.New(ctx, openai.AudioSpeechNewParams{
		Model: openai.SpeechModelTTS1,
		Input: text,
		Voice: s.mapVoiceToOpenAI(voice),
		ResponseFormat: openai.AudioSpeechNewParamsResponseFormatMP3,
		Speed: openai.Float(speed),
	})
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"service": "tts",
			"error":   err.Error(),
		}).Error("TTS API request failed")
		// 如果API調用失敗，返回mock響應作為fallback
		return s.mockTTSResponse(text, voice), nil
	}
	// Note: Audio.Speech.New returns *http.Response which should be closed
	defer resp.Body.Close()

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
		"service":     "tts",
		"text_length": len(text),
		"audio_size":  len(audioData),
		"duration":    durationStr,
		"voice":       voice,
	}).Info("TTS generation completed")

	return &TTSResponse{
		AudioData: audioData,
		Format:    "mp3",
		Duration:  durationStr,
		Size:      int64(len(audioData)),
	}, nil
}

// mapVoiceToOpenAI 將角色語音映射到 OpenAI TTS 語音
func (s *TTSService) mapVoiceToOpenAI(voice string) openai.AudioSpeechNewParamsVoice {
	voiceMapping := map[string]openai.AudioSpeechNewParamsVoice{
		"voice_001": openai.AudioSpeechNewParamsVoiceAlloy,
		"voice_002": openai.AudioSpeechNewParamsVoiceEcho,
		"voice_003": openai.AudioSpeechNewParamsVoiceBallad,
		"voice_004": openai.AudioSpeechNewParamsVoiceCoral,
		"voice_005": openai.AudioSpeechNewParamsVoiceShimmer,
		"voice_006": openai.AudioSpeechNewParamsVoiceAsh,
		"voice_007": openai.AudioSpeechNewParamsVoiceSage,
		"voice_008": openai.AudioSpeechNewParamsVoiceVerse,
		"voice_009": openai.AudioSpeechNewParamsVoiceMarin,
		"voice_010": openai.AudioSpeechNewParamsVoiceCedar,
		"default":   openai.AudioSpeechNewParamsVoiceAlloy,
	}

	if mappedVoice, exists := voiceMapping[voice]; exists {
		return mappedVoice
	}

	return openai.AudioSpeechNewParamsVoiceAlloy // 默認語音
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
		"text":    text[:utils.Min(len(text), 50)] + "...",
		"voice":   voice,
	}).Info("TTS mock response generated")

	return &TTSResponse{
		AudioData: mockAudioData,
		Format:    "mp3",
		Duration:  durationStr,
		Size:      int64(len(mockAudioData)),
	}
}

// GetAvailableVoices 獲取可用語音列表
func (s *TTSService) GetAvailableVoices() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"voice_id":     "voice_001",
			"name":         "標準男聲",
			"openai_voice": "alloy",
			"description":  "成熟穩重的男性聲音",
			"is_default":   true,
		},
		{
			"voice_id":     "voice_002",
			"name":         "深情男聲",
			"openai_voice": "echo",
			"description":  "深情磁性的男性聲音",
			"is_default":   false,
		},
		{
			"voice_id":     "voice_003",
			"name":         "詩意男聲",
			"openai_voice": "ballad",
			"description":  "充滿詩意的男性聲音",
			"is_default":   false,
		},
		{
			"voice_id":     "voice_004",
			"name":         "溫暖女聲",
			"openai_voice": "coral",
			"description":  "溫暖親切的女性聲音",
			"is_default":   false,
		},
		{
			"voice_id":     "voice_005",
			"name":         "清新女聲",
			"openai_voice": "shimmer",
			"description":  "清新甜美的女性聲音",
			"is_default":   false,
		},
		{
			"voice_id":     "voice_006",
			"name":         "低沉男聲",
			"openai_voice": "ash",
			"description":  "低沉有力的男性聲音",
			"is_default":   false,
		},
		{
			"voice_id":     "voice_007",
			"name":         "智慧男聲",
			"openai_voice": "sage",
			"description":  "智慧沉穩的男性聲音",
			"is_default":   false,
		},
		{
			"voice_id":     "voice_008",
			"name":         "優雅女聲",
			"openai_voice": "verse",
			"description":  "優雅細膩的女性聲音",
			"is_default":   false,
		},
		{
			"voice_id":     "voice_009",
			"name":         "海洋女聲",
			"openai_voice": "marin",
			"description":  "如海洋般深邃的女性聲音",
			"is_default":   false,
		},
		{
			"voice_id":     "voice_010",
			"name":         "雪松男聲",
			"openai_voice": "cedar",
			"description":  "如雪松般穩重的男性聲音",
			"is_default":   false,
		},
	}
}
