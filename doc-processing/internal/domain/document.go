package domain

import "time"

type DocumentType string

const (
	DocumentTypePDF  DocumentType = "pdf"
	DocumentTypeDOCX DocumentType = "docx"
	DocumentTypeTXT  DocumentType = "txt"
)

type Document struct {
	ID             ID
	OrganizationID ID
	Name           string
	Type           DocumentType
	S3Key          string
	Size           int64
	Status         ProcessingStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
)

func (d *Document) IsProcessable() bool {
	switch d.Type {
	case DocumentTypePDF, DocumentTypeDOCX, DocumentTypeTXT:
		return true
	default:
		return false
	}
}
