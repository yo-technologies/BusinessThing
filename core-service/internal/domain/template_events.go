package domain

import "time"

// TemplateIndexJob represents a job to index a contract template
type TemplateIndexJob struct {
	JobType      string  `json:"job_type"`
	TemplateID   *ID     `json:"template_id,omitempty"`
	TemplateName *string `json:"template_name,omitempty"`
	Description  *string `json:"description,omitempty"`
	TemplateType *string `json:"template_type,omitempty"`
	FieldsCount  *int    `json:"fields_count,omitempty"`
	RetryCount   int     `json:"retry_count"`
	MaxRetries   int     `json:"max_retries"`
	CreatedAt    string  `json:"created_at"`
}

func NewTemplateIndexJob(templateID ID, name, description, templateType string, fieldsCount int) *TemplateIndexJob {
	return &TemplateIndexJob{
		JobType:      "template_index",
		TemplateID:   &templateID,
		TemplateName: &name,
		Description:  &description,
		TemplateType: &templateType,
		FieldsCount:  &fieldsCount,
		RetryCount:   0,
		MaxRetries:   3,
		CreatedAt:    time.Now().Format(time.RFC3339),
	}
}

// TemplateDeleteJob represents a job to delete a contract template from index
type TemplateDeleteJob struct {
	JobType    string `json:"job_type"`
	TemplateID *ID    `json:"template_id,omitempty"`
	RetryCount int    `json:"retry_count"`
	MaxRetries int    `json:"max_retries"`
	CreatedAt  string `json:"created_at"`
}

func NewTemplateDeleteJob(templateID ID) *TemplateDeleteJob {
	return &TemplateDeleteJob{
		JobType:    "template_delete",
		TemplateID: &templateID,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}
}
