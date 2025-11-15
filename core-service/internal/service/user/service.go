package user

import (
	"context"
	"core-service/internal/domain"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
)

type repository interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, id domain.ID) (domain.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID string) (domain.User, error)
	ListUsers(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.User, int, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	UpdateUserRole(ctx context.Context, id domain.ID, role domain.UserRole) error
	DeactivateUser(ctx context.Context, id domain.ID) error
	CreateInvitation(ctx context.Context, invitation domain.Invitation) (domain.Invitation, error)
	GetInvitationByToken(ctx context.Context, token string) (domain.Invitation, error)
	MarkInvitationAsUsed(ctx context.Context, id domain.ID) error
	CreateOrganizationMember(ctx context.Context, member domain.OrganizationMember) (domain.OrganizationMember, error)
	GetOrganizationMember(ctx context.Context, userID, organizationID domain.ID) (*domain.OrganizationMember, error)
	UpdateOrganizationMember(ctx context.Context, member domain.OrganizationMember) (domain.OrganizationMember, error)
}

type Service struct {
	repo repository
}

func New(repo repository) *Service {
	return &Service{repo: repo}
}

// InviteUser creates a new invitation for a user
func (s *Service) InviteUser(ctx context.Context, organizationID domain.ID, email string, role domain.UserRole, invitedBy domain.ID) (*domain.Invitation, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.user.InviteUser")
	defer span.Finish()

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, domain.NewInvalidArgumentError("email is required")
	}

	// Generate secure token
	token, err := generateToken()
	if err != nil {
		return nil, domain.NewInternalError("failed to generate invitation token", err)
	}

	// Invitation expires in 7 days
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	invitation := domain.NewInvitation(organizationID, email, role, token, expiresAt)

	created, err := s.repo.CreateInvitation(ctx, invitation)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

// AcceptInvitation связывает существующего пользователя с организацией через приглашение
func (s *Service) AcceptInvitation(ctx context.Context, userID domain.ID, token string) (*domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.user.AcceptInvitation")
	defer span.Finish()

	// Get invitation
	invitation, err := s.repo.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, domain.NewNotFoundError("invitation not found or invalid")
	}

	// Validate invitation
	if invitation.IsExpired() {
		return nil, domain.NewInvalidArgumentError("invitation has expired")
	}
	if invitation.IsUsed() {
		return nil, domain.NewInvalidArgumentError("invitation has already been used")
	}

	// Get user
	user, err := s.repo.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if user already member of this organization
	existingMember, err := s.repo.GetOrganizationMember(ctx, userID, invitation.OrganizationID)
	if err == nil && existingMember != nil && existingMember.Status == domain.UserStatusActive {
		return nil, domain.NewInvalidArgumentError("user is already a member of this organization")
	}

	// Create organization membership
	member := domain.NewOrganizationMember(invitation.OrganizationID, userID, invitation.Email, invitation.Role)
	_, err = s.repo.CreateOrganizationMember(ctx, member)
	if err != nil {
		return nil, err
	}

	// Mark invitation as used
	err = s.repo.MarkInvitationAsUsed(ctx, invitation.ID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ListUsers retrieves users for an organization
func (s *Service) ListUsers(ctx context.Context, organizationID domain.ID, page, pageSize int) ([]domain.User, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.user.ListUsers")
	defer span.Finish()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	return s.repo.ListUsers(ctx, organizationID, pageSize, offset)
}

// ListUsersByOrganization retrieves all users for an organization (without pagination)
func (s *Service) ListUsersByOrganization(ctx context.Context, organizationID domain.ID) ([]*domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.user.ListUsersByOrganization")
	defer span.Finish()

	users, _, err := s.repo.ListUsers(ctx, organizationID, 1000, 0)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}
	return result, nil
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, id domain.ID) (*domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.user.GetUser")
	defer span.Finish()

	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByTelegramID retrieves a user by Telegram ID
func (s *Service) GetUserByTelegramID(ctx context.Context, telegramID string) (*domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.user.GetUserByTelegramID")
	defer span.Finish()

	user, err := s.repo.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserRole updates a user's role
func (s *Service) UpdateUserRole(ctx context.Context, id domain.ID, role domain.UserRole) (*domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.user.UpdateUserRole")
	defer span.Finish()

	if err := s.repo.UpdateUserRole(ctx, id, role); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// DeactivateUser deactivates a user
func (s *Service) DeactivateUser(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.user.DeactivateUser")
	defer span.Finish()

	return s.repo.DeactivateUser(ctx, id)
}

// generateToken generates a secure random token
func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetInvitationURL generates invitation URL
func GetInvitationURL(token string, baseURL string) string {
	return fmt.Sprintf("%s/invite/%s", baseURL, token)
}
