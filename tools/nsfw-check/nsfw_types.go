package main

// corpusDataEntry NSFW 語料庫數據條目
type corpusDataEntry struct {
	ID      string   `json:"id"`
	Level   int      `json:"level"`
	Tags    []string `json:"tags"`
	Locale  string   `json:"locale"`
	Text    string   `json:"text"`
	Reason  string   `json:"reason"`
	Version string   `json:"version,omitempty"`
}

// embeddingEntry 向量嵌入條目
type embeddingEntry struct {
	ID        string    `json:"id"`
	Embedding []float64 `json:"embedding"`
	Version   string    `json:"version"`
}
