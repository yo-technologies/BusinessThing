package domain

import "time"

// UserRole определяет роль пользователя в организации
type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleEmployee UserRole = "employee"
)

// UserStatus определяет статус пользователя
type UserStatus string

const (
	UserStatusPending  UserStatus = "pending"  // Приглашен, но не принял
	UserStatusActive   UserStatus = "active"   // Активный пользователь
	UserStatusInactive UserStatus = "inactive" // Деактивирован
)

// User представляет пользователя системы
type User struct {
	ID             ID         `db:"id"`
	OrganizationID ID         `db:"organization_id"`
	Email          string     `db:"email"`
	TelegramID     string     `db:"telegram_id"`
	FirstName      string     `db:"first_name"`
	LastName       string     `db:"last_name"`
	Role           UserRole   `db:"role"`
	Status         UserStatus `db:"status"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

// NewUser создает нового пользователя
func NewUser(organizationID ID, email string, role UserRole) User {
	now := time.Now()
	return User{
		ID:             NewID(),
		OrganizationID: organizationID,
		Email:          email,
		Role:           role,
		Status:         UserStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Activate активирует пользователя после принятия приглашения
func (u *User) Activate(telegramID, firstName, lastName string) {
	u.TelegramID = telegramID
	u.FirstName = firstName
	u.LastName = lastName
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
}

// Deactivate деактивирует пользователя
func (u *User) Deactivate() {
	u.Status = UserStatusInactive
	u.UpdatedAt = time.Now()
}

// UpdateRole обновляет роль пользователя
func (u *User) UpdateRole(role UserRole) {
	u.Role = role
	u.UpdatedAt = time.Now()
}

// IsActive проверяет, активен ли пользователь
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// Invitation представляет приглашение пользователя
type Invitation struct {
	ID             ID         `db:"id"`
	OrganizationID ID         `db:"organization_id"`
	Email          string     `db:"email"`
	Token          string     `db:"token"`
	Role           UserRole   `db:"role"`
	ExpiresAt      time.Time  `db:"expires_at"`
	UsedAt         *time.Time `db:"used_at"`
	CreatedAt      time.Time  `db:"created_at"`
}

// NewInvitation создает новое приглашение
func NewInvitation(organizationID ID, email string, role UserRole, token string, expiresAt time.Time) Invitation {
	return Invitation{
		ID:             NewID(),
		OrganizationID: organizationID,
		Email:          email,
		Token:          token,
		Role:           role,
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
	}
}

// IsExpired проверяет, истекло ли приглашение
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsUsed проверяет, использовано ли приглашение
func (i *Invitation) IsUsed() bool {
	return i.UsedAt != nil
}

// MarkAsUsed помечает приглашение как использованное
func (i *Invitation) MarkAsUsed() {
	now := time.Now()
	i.UsedAt = &now
}
