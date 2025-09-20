package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/clarencetw/thewavess-ai-core/utils"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// EmbeddingClient 定義 NSFW RAG 分類器所需的嵌入介面。
type EmbeddingClient interface {
	EmbedText(ctx context.Context, input string) ([]float32, error)
}

// OpenAIEmbeddingClient 透過 OpenAI 嵌入 API 實作 EmbeddingClient。
type OpenAIEmbeddingClient struct {
	client openai.Client
	model  openai.EmbeddingModel
}

// NewOpenAIEmbeddingClient 依環境變數建立嵌入客戶端。
func NewOpenAIEmbeddingClient() (*OpenAIEmbeddingClient, error) {
	utils.LoadEnv()

	apiKey := utils.GetEnvWithDefault("OPENAI_API_KEY", "")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required for embedding support")
	}

	baseURL := utils.GetEnvWithDefault("OPENAI_API_URL", "https://api.openai.com/v1")
	modelName := utils.GetEnvWithDefault("NSFW_EMBED_MODEL", string(openai.EmbeddingModelTextEmbedding3Small))

	var model openai.EmbeddingModel
	switch modelName {
	case string(openai.EmbeddingModelTextEmbeddingAda002):
		model = openai.EmbeddingModelTextEmbeddingAda002
	case string(openai.EmbeddingModelTextEmbedding3Large):
		model = openai.EmbeddingModelTextEmbedding3Large
	case string(openai.EmbeddingModelTextEmbedding3Small):
		model = openai.EmbeddingModelTextEmbedding3Small
	default:
		utils.Logger.WithField("model", modelName).Warn("Unknown embedding model, defaulting to text-embedding-3-small")
		model = openai.EmbeddingModelTextEmbedding3Small
	}

	var client openai.Client
	if baseURL != "https://api.openai.com/v1" {
		client = openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseURL))
	} else {
		client = openai.NewClient(option.WithAPIKey(apiKey))
	}

	utils.Logger.WithFields(map[string]interface{}{
		"service": "openai_embeddings",
		"model":   model,
		"baseURL": baseURL,
	}).Info("OpenAI embedding client initialized")

	return &OpenAIEmbeddingClient{client: client, model: model}, nil
}

// EmbedText 為指定文字產生嵌入向量。
func (c *OpenAIEmbeddingClient) EmbedText(ctx context.Context, input string) ([]float32, error) {
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("cannot embed empty input")
	}

	params := openai.EmbeddingNewParams{
		Model: c.model,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(input),
		},
	}

	resp, err := c.client.Embeddings.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("embedding response had no data")
	}

	raw := resp.Data[0].Embedding
	vector := make([]float32, len(raw))
	for i, v := range raw {
		vector[i] = float32(v)
	}

	return vector, nil
}
