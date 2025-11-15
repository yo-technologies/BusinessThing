package organization

import (
	"context"
	"core-service/internal/domain"
	"strings"

	"github.com/opentracing/opentracing-go"
)

type repository interface {
	CreateOrganization(ctx context.Context, org domain.Organization) (domain.Organization, error)
	GetOrganization(ctx context.Context, id domain.ID) (domain.Organization, error)
	GetOrganizationsByUserID(ctx context.Context, userID domain.ID) ([]domain.Organization, error)
	UpdateOrganization(ctx context.Context, org domain.Organization) (domain.Organization, error)
	DeleteOrganization(ctx context.Context, id domain.ID) error
}

type Service struct {
	repo repository
}

func New(repo repository) *Service {
	return &Service{repo: repo}
}

// CreateOrganization creates a new organization
func (s *Service) CreateOrganization(ctx context.Context, name, industry, region, description, profileData string) (domain.Organization, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.organization.CreateOrganization")
	defer span.Finish()

	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Organization{}, domain.NewInvalidArgumentError("organization name is required")
	}

	org := domain.NewOrganization(name, industry, region, description, profileData)

	created, err := s.repo.CreateOrganization(ctx, org)
	if err != nil {
		return domain.Organization{}, err
	}

	return created, nil
}

// GetOrganization retrieves an organization by ID
func (s *Service) GetOrganization(ctx context.Context, id domain.ID) (domain.Organization, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.organization.GetOrganization")
	defer span.Finish()

	return s.repo.GetOrganization(ctx, id)
}

// UpdateOrganization updates an existing organization
func (s *Service) UpdateOrganization(ctx context.Context, id domain.ID, name, industry, region, description, profileData *string) (domain.Organization, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.organization.UpdateOrganization")
	defer span.Finish()

	// Get existing organization
	org, err := s.repo.GetOrganization(ctx, id)
	if err != nil {
		return domain.Organization{}, err
	}

	// Update fields if provided
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return domain.Organization{}, domain.NewInvalidArgumentError("organization name cannot be empty")
		}
		org.Name = trimmed
	}
	if industry != nil {
		org.Industry = *industry
	}
	if region != nil {
		org.Region = *region
	}
	if description != nil {
		org.Description = *description
	}
	if profileData != nil {
		org.ProfileData = *profileData
	}

	updated, err := s.repo.UpdateOrganization(ctx, org)
	if err != nil {
		return domain.Organization{}, err
	}

	return updated, nil
}

// DeleteOrganization soft deletes an organization
func (s *Service) DeleteOrganization(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.organization.DeleteOrganization")
	defer span.Finish()

	return s.repo.DeleteOrganization(ctx, id)
}

// ListMyOrganizations retrieves all organizations where the user is a member
func (s *Service) ListMyOrganizations(ctx context.Context, userID domain.ID) ([]domain.Organization, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.organization.ListMyOrganizations")
	defer span.Finish()

	return s.repo.GetOrganizationsByUserID(ctx, userID)
}
