package domain

import "time"

// TemplateIndexJob represents a job to index a contract template
type TemplateIndexJob struct {
	JobType      string    `json:"JobType"`
	TemplateID   ID        `json:"TemplateID"`
	TemplateName string    `json:"TemplateName"`
	Description  string    `json:"Description"`
	TemplateType string    `json:"TemplateType"`
	FieldsCount  int       `json:"FieldsCount"`
	RetryCount   int       `json:"RetryCount"`
	MaxRetries   int       `json:"MaxRetries"`
	CreatedAt    time.Time `json:"CreatedAt"`
}

func NewTemplateIndexJob(templateID ID, name, description, templateType string, fieldsCount int) *TemplateIndexJob {
	return &TemplateIndexJob{
		JobType:      "template_index",
		TemplateID:   templateID,
		TemplateName: name,
		Description:  description,
		TemplateType: templateType,
		FieldsCount:  fieldsCount,
		RetryCount:   0,
		MaxRetries:   3,
		CreatedAt:    time.Now(),
	}
}

// TemplateDeleteJob represents a job to delete a contract template from index
type TemplateDeleteJob struct {
	JobType    string    `json:"JobType"`
	TemplateID ID        `json:"TemplateID"`
	RetryCount int       `json:"RetryCount"`
	MaxRetries int       `json:"MaxRetries"`
	CreatedAt  time.Time `json:"CreatedAt"`
}

func NewTemplateDeleteJob(templateID ID) *TemplateDeleteJob {
	return &TemplateDeleteJob{
		JobType:    "template_delete",
		TemplateID: templateID,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}
}
