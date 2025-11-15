package domain

import "time"

// GeneratedContract представляет сгенерированный договор
type GeneratedContract struct {
	ID             ID        `db:"id"`
	OrganizationID ID        `db:"organization_id"`
	TemplateID     ID        `db:"template_id"`
	Name           string    `db:"name"`
	FilledData     string    `db:"filled_data"` // JSON с заполненными значениями
	S3Key          string    `db:"s3_key"`
	FileType       string    `db:"file_type"`
	CreatedAt      time.Time `db:"created_at"`
}

// NewGeneratedContract создает новый сгенерированный договор
func NewGeneratedContract(organizationID, templateID ID, name, filledData, s3Key, fileType string) GeneratedContract {
	return GeneratedContract{
		ID:             NewID(),
		OrganizationID: organizationID,
		TemplateID:     templateID,
		Name:           name,
		FilledData:     filledData,
		S3Key:          s3Key,
		FileType:       fileType,
		CreatedAt:      time.Now(),
	}
}
