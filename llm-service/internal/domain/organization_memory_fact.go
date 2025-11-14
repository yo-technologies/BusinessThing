package domain

// OrganizationMemoryFact — короткий факт об организации, сохраняемый агентом.
// Используется для сохранения важной информации об организации для использования в будущих диалогах.
type OrganizationMemoryFact struct {
	Model
	OrganizationID ID     `db:"organization_id"`
	Content        string `db:"content"`
}

func NewOrganizationMemoryFact(organizationID ID, content string) OrganizationMemoryFact {
	return OrganizationMemoryFact{
		Model:          NewModel(),
		OrganizationID: organizationID,
		Content:        content,
	}
}
