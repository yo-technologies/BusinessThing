package domain

import "time"

type JobType string

const (
	JobTypeDocument       JobType = "document"
	JobTypeTemplateIndex  JobType = "template_index"
	JobTypeTemplateDelete JobType = "template_delete"
)

type ProcessingJob struct {
	JobType JobType `json:"job_type"`
	S3Key   string  `json:"s3_key,omitempty"`

	// Для документов
	DocumentID     ID           `json:"document_id,omitempty"`
	DocumentType   DocumentType `json:"document_type,omitempty"`
	DocumentName   string       `json:"document_name,omitempty"`
	OrganizationID ID           `json:"organization_id,omitempty"`

	// Для шаблонов
	TemplateID   *ID     `json:"template_id,omitempty"`
	TemplateName *string `json:"template_name,omitempty"`
	Description  *string `json:"description,omitempty"`
	TemplateType *string `json:"template_type,omitempty"`
	FieldsCount  *int    `json:"fields_count,omitempty"`

	RetryCount int       `json:"retry_count"`
	MaxRetries int       `json:"max_retries"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewProcessingJob(documentID, organizationID ID, s3Key string, docType DocumentType, docName string) *ProcessingJob {
	return &ProcessingJob{
		JobType:        JobTypeDocument,
		DocumentID:     documentID,
		OrganizationID: organizationID,
		S3Key:          s3Key,
		DocumentType:   docType,
		DocumentName:   docName,
		RetryCount:     0,
		MaxRetries:     3,
		CreatedAt:      time.Now(),
	}
}

func NewTemplateIndexJob(templateID ID, name, description, templateType string, fieldsCount int) *ProcessingJob {
	return &ProcessingJob{
		JobType:      JobTypeTemplateIndex,
		TemplateID:   &templateID,
		TemplateName: &name,
		Description:  &description,
		TemplateType: &templateType,
		FieldsCount:  &fieldsCount,
		RetryCount:   0,
		MaxRetries:   3,
		CreatedAt:    time.Now(),
	}
}

func NewTemplateDeleteJob(templateID ID) *ProcessingJob {
	return &ProcessingJob{
		JobType:    JobTypeTemplateDelete,
		TemplateID: &templateID,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}
}

func (j *ProcessingJob) CanRetry() bool {
	return j.RetryCount < j.MaxRetries
}

func (j *ProcessingJob) IncrementRetry() {
	j.RetryCount++
}
