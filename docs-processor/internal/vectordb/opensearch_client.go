package vectordb

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"docs-processor/internal/domain"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
	opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/opentracing/opentracing-go"
)

type OpenSearchClient struct {
	client    *opensearch.Client
	indexName string
}

func NewOpenSearchClient(addresses []string, username, password, indexName string) (*OpenSearchClient, error) {
	cfg := opensearch.Config{
		Addresses: addresses,
		Username:  username,
		Password:  password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	osc := &OpenSearchClient{
		client:    client,
		indexName: indexName,
	}

	if err := osc.ensureIndex(context.Background()); err != nil {
		return nil, err
	}

	return osc, nil
}

func (c *OpenSearchClient) ensureIndex(ctx context.Context) error {
	req := opensearchapi.IndicesExistsRequest{
		Index: []string{c.indexName},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return c.createIndex(ctx)
	}

	return nil
}

func (c *OpenSearchClient) createIndex(ctx context.Context) error {
	indexBody := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"chunk_id": map[string]interface{}{
					"type": "keyword",
				},
				"document_id": map[string]interface{}{
					"type": "keyword",
				},
				"organization_id": map[string]interface{}{
					"type": "keyword",
				},
				"document_name": map[string]interface{}{
					"type": "text",
				},
				"content": map[string]interface{}{
					"type": "text",
				},
				"position": map[string]interface{}{
					"type": "integer",
				},
				"embedding": map[string]interface{}{
					"type":      "knn_vector",
					"dimension": 1024,
				},
				"metadata": map[string]interface{}{
					"type": "object",
				},
			},
		},
		"settings": map[string]interface{}{
			"index": map[string]interface{}{
				"knn": true,
			},
		},
	}

	body, err := json.Marshal(indexBody)
	if err != nil {
		return err
	}

	req := opensearchapi.IndicesCreateRequest{
		Index: c.indexName,
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to create index: %s", string(bodyBytes))
	}

	return nil
}

type IndexChunkRequest struct {
	ChunkID        string            `json:"chunk_id"`
	DocumentID     string            `json:"document_id"`
	OrganizationID string            `json:"organization_id"`
	DocumentName   string            `json:"document_name"`
	Content        string            `json:"content"`
	Position       int               `json:"position"`
	Embedding      []float32         `json:"embedding"`
	Metadata       map[string]string `json:"metadata"`
}

func (c *OpenSearchClient) IndexChunk(ctx context.Context, chunk *domain.Chunk, organizationID domain.ID, documentName string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "vectordb.OpenSearchClient.IndexChunk")
	defer span.Finish()

	doc := IndexChunkRequest{
		ChunkID:        chunk.ID.String(),
		DocumentID:     chunk.DocumentID.String(),
		OrganizationID: organizationID.String(),
		DocumentName:   documentName,
		Content:        chunk.Content,
		Position:       chunk.Position,
		Embedding:      chunk.Embedding,
		Metadata:       chunk.Metadata,
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal chunk for indexing: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      c.indexName,
		DocumentID: chunk.ID.String(),
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to index chunk: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to index chunk (status %d): %s", res.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *OpenSearchClient) SearchChunks(ctx context.Context, organizationID domain.ID, queryEmbedding []float32, limit int, minScore float32) ([]*domain.SearchResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "vectordb.OpenSearchClient.SearchChunks")
	defer span.Finish()

	query := map[string]interface{}{
		"size": limit,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"organization_id": organizationID.String(),
						},
					},
					map[string]interface{}{
						"knn": map[string]interface{}{
							"embedding": map[string]interface{}{
								"vector": queryEmbedding,
								"k":      limit,
							},
						},
					},
				},
			},
		},
		"min_score": minScore,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	req := opensearchapi.SearchRequest{
		Index: []string{c.indexName},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search failed (status %d): %s", res.StatusCode, string(bodyBytes))
	}

	var searchResp searchResponse
	if err := json.NewDecoder(res.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	results := make([]*domain.SearchResult, 0, len(searchResp.Hits.Hits))
	for _, hit := range searchResp.Hits.Hits {
		results = append(results, &domain.SearchResult{
			ChunkID:      domain.ID(hit.Source.ChunkID),
			DocumentID:   domain.ID(hit.Source.DocumentID),
			DocumentName: hit.Source.DocumentName,
			Content:      hit.Source.Content,
			Position:     hit.Source.Position,
			Score:        hit.Score,
			Metadata:     hit.Source.Metadata,
		})
	}

	return results, nil
}

func (c *OpenSearchClient) DeleteDocumentChunks(ctx context.Context, documentID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "vectordb.OpenSearchClient.DeleteDocumentChunks")
	defer span.Finish()

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"document_id": documentID.String(),
			},
		},
	}

	body, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("failed to marshal delete query: %w", err)
	}

	req := opensearchapi.DeleteByQueryRequest{
		Index: []string{c.indexName},
		Body:  strings.NewReader(string(body)),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete chunks failed (status %d): %s", res.StatusCode, string(bodyBytes))
	}

	return nil
}

type searchResponse struct {
	Hits struct {
		Hits []struct {
			Score  float32 `json:"_score"`
			Source struct {
				ChunkID      string            `json:"chunk_id"`
				DocumentID   string            `json:"document_id"`
				DocumentName string            `json:"document_name"`
				Content      string            `json:"content"`
				Position     int               `json:"position"`
				Metadata     map[string]string `json:"metadata"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
