package domain

import "time"

// Note представляет заметку LLM об организации
type Note struct {
	ID             ID        `db:"id"`
	OrganizationID ID        `db:"organization_id"`
	Content        string    `db:"content"`
	CreatedAt      time.Time `db:"created_at"`
}

// NewNote создает новую заметку
func NewNote(organizationID ID, content string) Note {
	return Note{
		ID:             NewID(),
		OrganizationID: organizationID,
		Content:        content,
		CreatedAt:      time.Now(),
	}
}
