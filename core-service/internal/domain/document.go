package domain

import "time"

// DocumentStatus определяет статус обработки документа
type DocumentStatus string

const (
	DocumentStatusPending    DocumentStatus = "pending"    // Зарегистрирован, ожидает обработки
	DocumentStatusProcessing DocumentStatus = "processing" // В процессе обработки
	DocumentStatusIndexed    DocumentStatus = "indexed"    // Успешно проиндексирован
	DocumentStatusFailed     DocumentStatus = "failed"     // Ошибка обработки
)

// Document представляет документ организации
type Document struct {
	ID             ID             `db:"id"`
	OrganizationID ID             `db:"organization_id"`
	Name           string         `db:"name"`
	S3Key          string         `db:"s3_key"`
	FileType       string         `db:"file_type"`
	FileSize       int64          `db:"file_size"`
	Status         DocumentStatus `db:"status"`
	ErrorMessage   string         `db:"error_message"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

// NewDocument создает новый документ
func NewDocument(organizationID ID, name, s3Key, fileType string, fileSize int64) Document {
	now := time.Now()
	return Document{
		ID:             NewID(),
		OrganizationID: organizationID,
		Name:           name,
		S3Key:          s3Key,
		FileType:       fileType,
		FileSize:       fileSize,
		Status:         DocumentStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// UpdateStatus обновляет статус документа
func (d *Document) UpdateStatus(status DocumentStatus, errorMessage string) {
	d.Status = status
	d.ErrorMessage = errorMessage
	d.UpdatedAt = time.Now()
}

// IsIndexed проверяет, проиндексирован ли документ
func (d *Document) IsIndexed() bool {
	return d.Status == DocumentStatusIndexed
}

// IsFailed проверяет, провалилась ли обработка
func (d *Document) IsFailed() bool {
	return d.Status == DocumentStatusFailed
}
