package domain

import "time"

// ContractTemplate представляет шаблон договора
type ContractTemplate struct {
	ID              ID        `db:"id"`
	OrganizationID  ID        `db:"organization_id"`
	Name            string    `db:"name"`
	Description     string    `db:"description"`
	TemplateType    string    `db:"template_type"`
	FieldsSchema    string    `db:"fields_schema"`    // JSON схема полей
	ContentTemplate string    `db:"content_template"` // Шаблон содержимого
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// NewContractTemplate создает новый шаблон договора
func NewContractTemplate(organizationID ID, name, description, templateType, fieldsSchema, contentTemplate string) ContractTemplate {
	now := time.Now()
	return ContractTemplate{
		ID:              NewID(),
		OrganizationID:  organizationID,
		Name:            name,
		Description:     description,
		TemplateType:    templateType,
		FieldsSchema:    fieldsSchema,
		ContentTemplate: contentTemplate,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Update обновляет шаблон
func (t *ContractTemplate) Update(name, description, fieldsSchema, contentTemplate *string) {
	if name != nil {
		t.Name = *name
	}
	if description != nil {
		t.Description = *description
	}
	if fieldsSchema != nil {
		t.FieldsSchema = *fieldsSchema
	}
	if contentTemplate != nil {
		t.ContentTemplate = *contentTemplate
	}
	t.UpdatedAt = time.Now()
}
