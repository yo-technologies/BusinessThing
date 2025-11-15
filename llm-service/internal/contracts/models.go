package contracts

import "time"

type TemplateSearchResult struct {
	TemplateID   string          `json:"template_id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	TemplateType string          `json:"template_type"`
	Fields       []TemplateField `json:"fields"`
	FieldsCount  int             `json:"fields_count"`
	Score        float32         `json:"relevance_score"`
}

type TemplateField struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
	Min         *int   `json:"min,omitempty"`
	Max         *int   `json:"max,omitempty"`
}

type GeneratedContract struct {
	ContractID   string    `json:"contract_id"`
	Name         string    `json:"name"`
	DownloadURL  string    `json:"download_url"`
	S3Key        string    `json:"s3_key"`
	CreatedAt    time.Time `json:"created_at"`
	TemplateName string    `json:"template_name"`
}

type ContractListItem struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	TemplateName string    `json:"template_name"`
	TemplateID   string    `json:"template_id"`
	CreatedAt    time.Time `json:"created_at"`
	DownloadURL  string    `json:"download_url"`
}
