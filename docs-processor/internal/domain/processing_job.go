package domain

import "time"

type ProcessingJob struct {
	DocumentID     ID
	OrganizationID ID
	S3Key          string
	DocumentType   DocumentType
	DocumentName   string
	RetryCount     int
	MaxRetries     int
	CreatedAt      time.Time
}

func NewProcessingJob(documentID, organizationID ID, s3Key string, docType DocumentType, docName string) *ProcessingJob {
	return &ProcessingJob{
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

func (j *ProcessingJob) CanRetry() bool {
	return j.RetryCount < j.MaxRetries
}

func (j *ProcessingJob) IncrementRetry() {
	j.RetryCount++
}
