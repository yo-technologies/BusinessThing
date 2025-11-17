package domain

import "time"

type Template struct {
	ID           ID
	Name         string
	Description  string
	TemplateType string
	FieldsCount  int
	CreatedAt    time.Time
}
