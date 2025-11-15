package domain

import "time"

type JobType string

const (
	JobTypeDocument       JobType = "document"
	JobTypeTemplateIndex  JobType = "template_index"
	JobTypeTemplateDelete JobType = "template_delete"
)

type ProcessingJob struct {
	JobType        JobType
	DocumentID     ID
	OrganizationID ID
	S3Key          string
	DocumentType   DocumentType
	DocumentName   string

	// Для шаблонов
	TemplateID   *ID
	TemplateName *string
	Description  *string
	TemplateType *string
	FieldsCount  *int

	RetryCount int
	MaxRetries int
	CreatedAt  time.Time
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

func NewTemplateIndexJob(templateID, organizationID ID, name, description, templateType string, fieldsCount int) *ProcessingJob {
	return &ProcessingJob{
		JobType:        JobTypeTemplateIndex,
		OrganizationID: organizationID,
		TemplateID:     &templateID,
		TemplateName:   &name,
		Description:    &description,
		TemplateType:   &templateType,
		FieldsCount:    &fieldsCount,
		RetryCount:     0,
		MaxRetries:     3,
		CreatedAt:      time.Now(),
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
