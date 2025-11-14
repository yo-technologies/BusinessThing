package websearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"llm-service/internal/domain"
	"llm-service/internal/logger"

	"github.com/opentracing/opentracing-go"
)

const (
	tavilyAPIURL = "https://api.tavily.com/search"
)

// TavilyClient - клиент для работы с Tavily API
type TavilyClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewTavilyClient создает новый клиент для Tavily API
func NewTavilyClient(apiKey string) *TavilyClient {
	return &TavilyClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchRequest - запрос на поиск в Tavily API
type SearchRequest struct {
	Query             string   `json:"query"`
	SearchDepth       string   `json:"search_depth,omitempty"`        // basic or advanced
	MaxResults        int      `json:"max_results,omitempty"`         // default 5
	IncludeAnswer     bool     `json:"include_answer,omitempty"`      // default false
	IncludeRawContent bool     `json:"include_raw_content,omitempty"` // default false
	IncludeDomains    []string `json:"include_domains,omitempty"`     // список доменов для включения
	ExcludeDomains    []string `json:"exclude_domains,omitempty"`     // список доменов для исключения
}

// SearchResult - один результат поиска
type SearchResult struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	RawContent string  `json:"raw_content,omitempty"`
}

// SearchResponse - ответ от Tavily API
type SearchResponse struct {
	Query        string         `json:"query"`
	Answer       string         `json:"answer,omitempty"`
	Results      []SearchResult `json:"results"`
	ResponseTime float64        `json:"response_time"`
}

// Search выполняет поиск через Tavily API
func (c *TavilyClient) Search(ctx context.Context, query string, maxResults int) (interface{}, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "websearch.TavilyClient.Search")
	defer span.Finish()

	if query == "" {
		return nil, domain.NewInvalidArgumentError("search query cannot be empty")
	}

	if maxResults <= 0 {
		maxResults = 5
	}

	// Подготавливаем запрос
	reqBody := SearchRequest{
		Query:         query,
		SearchDepth:   "basic",
		MaxResults:    maxResults,
		IncludeAnswer: true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, domain.NewInternalError("failed to marshal request", err)
	}

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "POST", tavilyAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, domain.NewInternalError("failed to create request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Выполняем запрос
	logger.Info(ctx, fmt.Sprintf("Executing Tavily search: %s", query))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, domain.NewInternalError("failed to execute request", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, domain.NewInternalError("failed to read response", err)
	}

	// Проверяем статус код
	if resp.StatusCode != http.StatusOK {
		logger.Error(ctx, fmt.Sprintf("Tavily API error: %d - %s", resp.StatusCode, string(body)))
		return nil, domain.NewInternalError(fmt.Sprintf("tavily api returned status %d", resp.StatusCode), nil)
	}

	// Парсим ответ
	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, domain.NewInternalError("failed to unmarshal response", err)
	}

	logger.Info(ctx, fmt.Sprintf("Tavily search completed: found %d results in %.2fs", len(searchResp.Results), searchResp.ResponseTime))

	return &searchResp, nil
}
