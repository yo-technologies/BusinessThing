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

// User представляет пользователя системы (независимо от организаций)
type User struct {
	ID                    ID        `db:"id"`
	TelegramID            string    `db:"telegram_id"`
	FirstName             string    `db:"first_name"`
	LastName              string    `db:"last_name"`
	IsActive              bool      `db:"is_active"`
	RegistrationCompleted bool      `db:"registration_completed"`
	CreatedAt             time.Time `db:"created_at"`
	UpdatedAt             time.Time `db:"updated_at"`
}

// OrganizationMember представляет членство пользователя в организации
type OrganizationMember struct {
	ID             ID         `db:"id"`
	OrganizationID ID         `db:"organization_id"`
	UserID         ID         `db:"user_id"`
	Email          string     `db:"email"`
	Role           UserRole   `db:"role"`
	Status         UserStatus `db:"status"`
	JoinedAt       time.Time  `db:"joined_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

// UserWithMembership объединяет пользователя с его членством в организации (для API)
type UserWithMembership struct {
	User
	OrganizationMember
}

// NewUser создает нового пользователя с Telegram ID (при первой авторизации)
func NewUser(telegramID string) User {
	now := time.Now()
	return User{
		ID:                    NewID(),
		TelegramID:            telegramID,
		IsActive:              true,
		RegistrationCompleted: false,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

// NewOrganizationMember создает новое членство в организации
func NewOrganizationMember(organizationID, userID ID, email string, role UserRole) OrganizationMember {
	now := time.Now()
	return OrganizationMember{
		ID:             NewID(),
		OrganizationID: organizationID,
		UserID:         userID,
		Email:          email,
		Role:           role,
		Status:         UserStatusActive,
		JoinedAt:       now,
		UpdatedAt:      now,
	}
}

// CompleteProfile завершает регистрацию пользователя (добавление ФИО)
func (u *User) CompleteProfile(firstName, lastName string) {
	u.FirstName = firstName
	u.LastName = lastName
	u.RegistrationCompleted = true
	u.UpdatedAt = time.Now()
}

// Deactivate деактивирует пользователя
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// UpdateRole обновляет роль пользователя в организации
func (m *OrganizationMember) UpdateRole(role UserRole) {
	m.Role = role
	m.UpdatedAt = time.Now()
}

// Deactivate деактивирует членство в организации
func (m *OrganizationMember) Deactivate() {
	m.Status = UserStatusInactive
	m.UpdatedAt = time.Now()
}

// IsActive проверяет, активно ли членство
func (m *OrganizationMember) IsActive() bool {
	return m.Status == UserStatusActive
}

// Invitation представляет приглашение пользователя
type Invitation struct {
	ID             ID         `db:"id"`
	OrganizationID ID         `db:"organization_id"`
	Token          string     `db:"token"`
	Role           UserRole   `db:"role"`
	ExpiresAt      time.Time  `db:"expires_at"`
	UsedAt         *time.Time `db:"used_at"`
	CreatedAt      time.Time  `db:"created_at"`
}

// NewInvitation создает новое приглашение
func NewInvitation(organizationID ID, role UserRole, token string, expiresAt time.Time) Invitation {
	return Invitation{
		ID:             NewID(),
		OrganizationID: organizationID,
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
