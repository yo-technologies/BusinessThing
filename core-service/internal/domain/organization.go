package domain

import "time"

// Organization представляет организацию/компанию
type Organization struct {
	ID          ID         `db:"id"`
	Name        string     `db:"name"`
	Industry    string     `db:"industry"`
	Region      string     `db:"region"`
	Description string     `db:"description"`
	ProfileData string     `db:"profile_data"` // JSON с расширяемыми полями
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"` // Soft delete
}

// NewOrganization создает новую организацию
func NewOrganization(name, industry, region, description, profileData string) Organization {
	now := time.Now()
	return Organization{
		ID:          NewID(),
		Name:        name,
		Industry:    industry,
		Region:      region,
		Description: description,
		ProfileData: profileData,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IsDeleted проверяет, удалена ли организация
func (o *Organization) IsDeleted() bool {
	return o.DeletedAt != nil
}
