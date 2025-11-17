package vectordb

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"docs-processor/internal/domain"

	opensearch "github.com/opensearch-project/opensearch-go/v2"
	opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/opentracing/opentracing-go"
)

type TemplatesClient struct {
	client    *opensearch.Client
	indexName string
}

func NewTemplatesClient(addresses []string, username, password, indexName string) (*TemplatesClient, error) {
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

	tc := &TemplatesClient{
		client:    client,
		indexName: indexName,
	}

	if err := tc.ensureIndex(context.Background()); err != nil {
		return nil, err
	}

	return tc, nil
}

func (c *TemplatesClient) ensureIndex(ctx context.Context) error {
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

func (c *TemplatesClient) createIndex(ctx context.Context) error {
	indexBody := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"template_id": map[string]interface{}{
					"type": "keyword",
				},
				"organization_id": map[string]interface{}{
					"type": "keyword",
				},
				"name": map[string]interface{}{
					"type": "text",
				},
				"description": map[string]interface{}{
					"type": "text",
				},
				"template_type": map[string]interface{}{
					"type": "keyword",
				},
				"fields_count": map[string]interface{}{
					"type": "integer",
				},
				"embedding": map[string]interface{}{
					"type":      "knn_vector",
					"dimension": 1024,
				},
				"created_at": map[string]interface{}{
					"type": "date",
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
		return fmt.Errorf("failed to create templates index: %s", string(bodyBytes))
	}

	return nil
}

type IndexTemplateRequest struct {
	TemplateID     string    `json:"template_id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	TemplateType   string    `json:"template_type"`
	FieldsCount    int       `json:"fields_count"`
	Embedding      []float32 `json:"embedding"`
	CreatedAt      string    `json:"created_at"`
}

func (c *TemplatesClient) IndexTemplate(ctx context.Context, template *domain.Template, embedding []float32) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "vectordb.TemplatesClient.IndexTemplate")
	defer span.Finish()

	doc := IndexTemplateRequest{
		TemplateID:     template.ID.String(),
		Name:           template.Name,
		Description:    template.Description,
		TemplateType:   template.TemplateType,
		FieldsCount:    template.FieldsCount,
		Embedding:      embedding,
		CreatedAt:      template.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal template for indexing: %w", err)
	}

	req := opensearchapi.IndexRequest{
		Index:      c.indexName,
		DocumentID: template.ID.String(),
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to index template: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to index template (status %d): %s", res.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *TemplatesClient) SearchTemplates(ctx context.Context, organizationID domain.ID, queryEmbedding []float32, limit int) ([]*TemplateSearchResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "vectordb.TemplatesClient.SearchTemplates")
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

	var searchResp templateSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	results := make([]*TemplateSearchResult, 0, len(searchResp.Hits.Hits))
	for _, hit := range searchResp.Hits.Hits {
		templateID, _ := domain.ParseID(hit.Source.TemplateID)
		results = append(results, &TemplateSearchResult{
			TemplateID:   templateID,
			Name:         hit.Source.Name,
			Description:  hit.Source.Description,
			TemplateType: hit.Source.TemplateType,
			FieldsCount:  hit.Source.FieldsCount,
			Score:        hit.Score,
		})
	}

	return results, nil
}

func (c *TemplatesClient) DeleteTemplate(ctx context.Context, templateID domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "vectordb.TemplatesClient.DeleteTemplate")
	defer span.Finish()

	req := opensearchapi.DeleteRequest{
		Index:      c.indexName,
		DocumentID: templateID.String(),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete template failed (status %d): %s", res.StatusCode, string(bodyBytes))
	}

	return nil
}

type TemplateSearchResult struct {
	TemplateID   domain.ID
	Name         string
	Description  string
	TemplateType string
	FieldsCount  int
	Score        float32
}

type templateSearchResponse struct {
	Hits struct {
		Hits []struct {
			Score  float32 `json:"_score"`
			Source struct {
				TemplateID   string `json:"template_id"`
				Name         string `json:"name"`
				Description  string `json:"description"`
				TemplateType string `json:"template_type"`
				FieldsCount  int    `json:"fields_count"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
