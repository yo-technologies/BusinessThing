package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"

	"docs-processor/internal/logger"
)

type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type embeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

func (c *Client) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "embeddings.Client.GenerateEmbeddings")
	defer span.Finish()

	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	logger.Info(ctx, "generating embeddings", "texts_count", len(texts))

	reqBody := embeddingRequest{
		Input: texts,
		Model: c.model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error(ctx, "failed to marshal embeddings request", "error", err)
		return nil, fmt.Errorf("failed to marshal embeddings request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error(ctx, "failed to create embeddings request", "error", err)
		return nil, fmt.Errorf("failed to create embeddings request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		logger.Error(ctx, "failed to send embeddings request",
			"error", err,
			"duration", duration,
			"texts_count", len(texts))
		return nil, fmt.Errorf("failed to send embeddings request: %w", err)
	}
	defer resp.Body.Close()

	logger.Info(ctx, "embeddings request completed",
		"status_code", resp.StatusCode,
		"duration", duration,
		"texts_count", len(texts))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error(ctx, "embeddings API error",
			"status_code", resp.StatusCode,
			"response_body", string(body))
		return nil, fmt.Errorf("embeddings API error (status %d): %s", resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		logger.Error(ctx, "failed to decode embeddings response", "error", err)
		return nil, fmt.Errorf("failed to decode embeddings response: %w", err)
	}

	embeddings := make([][]float32, len(texts))
	for _, data := range embResp.Data {
		if data.Index < len(embeddings) {
			embeddings[data.Index] = data.Embedding
		}
	}

	logger.Info(ctx, "embeddings generated successfully",
		"embeddings_count", len(embeddings),
		"total_duration", duration)

	return embeddings, nil
}

func (c *Client) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "embeddings.Client.GenerateEmbedding")
	defer span.Finish()

	embeddings, err := c.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned from API")
	}

	return embeddings[0], nil
}
