package repository

import (
	"context"
	"core-service/internal/db"
	"core-service/internal/domain"

	"github.com/jackc/pgx/v5/pgtype"
)

// Repository aggregates all repository interfaces
type Repository interface {
	OrganizationRepository
	UserRepository
	InvitationRepository
	DocumentRepository
	NoteRepository
	ContractTemplateRepository
	GeneratedContractRepository
}

// OrganizationRepository defines methods for organization data access
type OrganizationRepository interface {
	CreateOrganization(ctx context.Context, org domain.Organization) (domain.Organization, error)
	GetOrganization(ctx context.Context, id domain.ID) (domain.Organization, error)
	GetOrganizationsByUserID(ctx context.Context, userID domain.ID) ([]domain.Organization, error)
	UpdateOrganization(ctx context.Context, org domain.Organization) (domain.Organization, error)
	DeleteOrganization(ctx context.Context, id domain.ID) error
}

// UserRepository defines methods for user data access
type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, id domain.ID) (domain.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID string) (domain.User, error)
	ListUsers(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.UserWithMembership, int, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	UpdateUserRole(ctx context.Context, id domain.ID, role domain.UserRole) error
	DeactivateUser(ctx context.Context, id domain.ID) error
}

// InvitationRepository defines methods for invitation data access
type InvitationRepository interface {
	CreateInvitation(ctx context.Context, invitation domain.Invitation) (domain.Invitation, error)
	GetInvitationByToken(ctx context.Context, token string) (domain.Invitation, error)
	ListInvitations(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.Invitation, int, error)
	MarkInvitationAsUsed(ctx context.Context, id domain.ID) error
}

// DocumentRepository defines methods for document data access
type DocumentRepository interface {
	CreateDocument(ctx context.Context, doc domain.Document) (domain.Document, error)
	GetDocument(ctx context.Context, id domain.ID) (domain.Document, error)
	ListDocuments(ctx context.Context, organizationID domain.ID, status *domain.DocumentStatus, limit, offset int) ([]domain.Document, int, error)
	UpdateDocumentStatus(ctx context.Context, id domain.ID, status domain.DocumentStatus, errorMessage string) error
	DeleteDocument(ctx context.Context, id domain.ID) error
}

// NoteRepository defines methods for note data access
type NoteRepository interface {
	CreateNote(ctx context.Context, note domain.Note) (domain.Note, error)
	GetNote(ctx context.Context, id domain.ID) (domain.Note, error)
	ListNotes(ctx context.Context, organizationID domain.ID, limit int) ([]domain.Note, error)
	DeleteNote(ctx context.Context, id domain.ID) error
}

// ContractTemplateRepository defines methods for contract template data access
type ContractTemplateRepository interface {
	CreateTemplate(ctx context.Context, template domain.ContractTemplate) (domain.ContractTemplate, error)
	GetTemplate(ctx context.Context, id domain.ID) (domain.ContractTemplate, error)
	ListTemplates(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.ContractTemplate, int, error)
	UpdateTemplate(ctx context.Context, template domain.ContractTemplate) (domain.ContractTemplate, error)
	DeleteTemplate(ctx context.Context, id domain.ID) error
}

// GeneratedContractRepository defines methods for generated contract data access
type GeneratedContractRepository interface {
	CreateContract(ctx context.Context, contract domain.GeneratedContract) (domain.GeneratedContract, error)
	GetContract(ctx context.Context, id domain.ID) (domain.GeneratedContract, error)
	ListContracts(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.GeneratedContract, int, error)
	ListContractsByTemplate(ctx context.Context, templateID domain.ID) ([]domain.GeneratedContract, error)
	DeleteContract(ctx context.Context, id domain.ID) error
}

// PGXRepository implements Repository using PostgreSQL
type PGXRepository struct {
	engineFactory db.EngineFactory
	transactioner db.Transactioner
}

// NewPGXRepository creates a new PostgreSQL repository
func NewPGXRepository(cm *db.ContextManager) *PGXRepository {
	return &PGXRepository{
		engineFactory: cm,
		transactioner: cm,
	}
}

// Helper function to convert domain.ID to pgtype.UUID
func uuidToPgtype(id domain.ID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: [16]byte(id),
		Valid: true,
	}
}
