package repository

import (
	"context"
	"core-service/internal/domain"
	"core-service/internal/logger"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
)

// CreateUser inserts a new user
func (r *PGXRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateUser")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        INSERT INTO users (id, telegram_id, first_name, last_name, is_active, registration_completed, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, telegram_id, first_name, last_name, is_active, registration_completed, created_at, updated_at
    `

	var created domain.User
	err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(user.ID),
		user.TelegramID,
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.RegistrationCompleted,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		logger.Errorf(ctx, "failed to create user: %v", err)
		return domain.User{}, err
	}

	return created, nil
}

// GetUser retrieves a user by ID
func (r *PGXRepository) GetUser(ctx context.Context, id domain.ID) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetUser")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, telegram_id, first_name, last_name, is_active, registration_completed, created_at, updated_at
        FROM users
        WHERE id = $1
    `

	var user domain.User
	err := pgxscan.Get(ctx, engine, &user, query, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to get user: %v", err)
		return domain.User{}, err
	}

	return user, nil
}

// GetUserByTelegramID retrieves a user by Telegram ID
func (r *PGXRepository) GetUserByTelegramID(ctx context.Context, telegramID string) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetUserByTelegramID")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, telegram_id, first_name, last_name, is_active, registration_completed, created_at, updated_at
        FROM users
        WHERE telegram_id = $1
    `

	var user domain.User
	err := pgxscan.Get(ctx, engine, &user, query, telegramID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to get user by telegram_id: %v", err)
		return domain.User{}, err
	}

	return user, nil
}

// ListUsers retrieves users for an organization with pagination
func (r *PGXRepository) ListUsers(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.UserWithMembership, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListUsers")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	// Get total count
	countQuery := `
		SELECT COUNT(*) 
		FROM users u
		INNER JOIN organization_members om ON u.id = om.user_id
		WHERE om.organization_id = $1
	`
	var total int
	err := pgxscan.Get(ctx, engine, &total, countQuery, uuidToPgtype(organizationID))
	if err != nil {
		logger.Errorf(ctx, "failed to count users: %v", err)
		return nil, 0, err
	}

	// Get users with membership info
	query := `
        SELECT 
            u.id, u.telegram_id, u.first_name, u.last_name, u.is_active, u.registration_completed, u.created_at, u.updated_at,
            om.id as "organization_member.id",
            om.organization_id as "organization_member.organization_id",
            om.user_id as "organization_member.user_id",
            om.email as "organization_member.email",
            om.role as "organization_member.role",
            om.status as "organization_member.status",
            om.joined_at as "organization_member.joined_at",
            om.updated_at as "organization_member.updated_at"
        FROM users u
        INNER JOIN organization_members om ON u.id = om.user_id
        WHERE om.organization_id = $1
        ORDER BY om.joined_at DESC
        LIMIT $2 OFFSET $3
    `

	var users []domain.UserWithMembership
	err = pgxscan.Select(ctx, engine, &users, query, uuidToPgtype(organizationID), limit, offset)
	if err != nil {
		logger.Errorf(ctx, "failed to list users: %v", err)
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUser updates an existing user
func (r *PGXRepository) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateUser")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        UPDATE users
        SET telegram_id = $2, first_name = $3, last_name = $4, is_active = $5, registration_completed = $6, updated_at = $7
        WHERE id = $1
        RETURNING id, telegram_id, first_name, last_name, is_active, registration_completed, created_at, updated_at
    `

	var updated domain.User
	err := pgxscan.Get(ctx, engine, &updated, query,
		uuidToPgtype(user.ID),
		user.TelegramID,
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.RegistrationCompleted,
		user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to update user: %v", err)
		return domain.User{}, err
	}

	return updated, nil
}

// UpdateUserRole updates a user's role in an organization (deprecated - use UpdateOrganizationMemberRole)
func (r *PGXRepository) UpdateUserRole(ctx context.Context, id domain.ID, role domain.UserRole) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateUserRole")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `UPDATE organization_members SET role = $2, updated_at = NOW() WHERE user_id = $1`

	tag, err := engine.Exec(ctx, query, uuidToPgtype(id), role)
	if err != nil {
		logger.Errorf(ctx, "failed to update user role: %v", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// DeactivateUser deactivates a user
func (r *PGXRepository) DeactivateUser(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.DeactivateUser")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1`

	tag, err := engine.Exec(ctx, query, uuidToPgtype(id))
	if err != nil {
		logger.Errorf(ctx, "failed to deactivate user: %v", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// CreateInvitation inserts a new invitation
func (r *PGXRepository) CreateInvitation(ctx context.Context, invitation domain.Invitation) (domain.Invitation, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateInvitation")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        INSERT INTO invitations (id, organization_id, token, role, expires_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, organization_id, token, role, expires_at, used_at, created_at
    `

	var created domain.Invitation
	err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(invitation.ID),
		uuidToPgtype(invitation.OrganizationID),
		invitation.Token,
		invitation.Role,
		invitation.ExpiresAt,
		invitation.CreatedAt,
	)
	if err != nil {
		logger.Errorf(ctx, "failed to create invitation: %v", err)
		return domain.Invitation{}, err
	}

	return created, nil
}

// GetInvitationByToken retrieves an invitation by token
func (r *PGXRepository) GetInvitationByToken(ctx context.Context, token string) (domain.Invitation, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetInvitationByToken")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, organization_id, token, role, expires_at, used_at, created_at
        FROM invitations
        WHERE token = $1
    `

	var invitation domain.Invitation
	err := pgxscan.Get(ctx, engine, &invitation, query, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Invitation{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to get invitation: %v", err)
		return domain.Invitation{}, err
	}

	return invitation, nil
}

// ListInvitations retrieves invitations for an organization
func (r *PGXRepository) ListInvitations(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.Invitation, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListInvitations")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM invitations
		WHERE organization_id = $1
	`
	err := pgxscan.Get(ctx, engine, &total, countQuery, uuidToPgtype(organizationID))
	if err != nil {
		logger.Errorf(ctx, "failed to count invitations: %v", err)
		return nil, 0, err
	}

	// Get invitations
	query := `
		SELECT id, organization_id, token, role, expires_at, used_at, created_at
		FROM invitations
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var invitations []domain.Invitation
	err = pgxscan.Select(ctx, engine, &invitations, query,
		uuidToPgtype(organizationID),
		limit,
		offset,
	)
	if err != nil {
		logger.Errorf(ctx, "failed to list invitations: %v", err)
		return nil, 0, err
	}

	if invitations == nil {
		invitations = []domain.Invitation{}
	}

	return invitations, total, nil
}

// MarkInvitationAsUsed marks an invitation as used
func (r *PGXRepository) MarkInvitationAsUsed(ctx context.Context, id domain.ID) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.MarkInvitationAsUsed")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `UPDATE invitations SET used_at = NOW() WHERE id = $1 AND used_at IS NULL`

	tag, err := engine.Exec(ctx, query, uuidToPgtype(id))
	if err != nil {
		logger.Errorf(ctx, "failed to mark invitation as used: %v", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// CreateOrganizationMember creates a new organization membership
func (r *PGXRepository) CreateOrganizationMember(ctx context.Context, member domain.OrganizationMember) (domain.OrganizationMember, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.CreateOrganizationMember")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        INSERT INTO organization_members (id, organization_id, user_id, email, role, status, joined_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, organization_id, user_id, email, role, status, joined_at, updated_at
    `

	var created domain.OrganizationMember
	err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(member.ID),
		uuidToPgtype(member.OrganizationID),
		uuidToPgtype(member.UserID),
		member.Email,
		member.Role,
		member.Status,
		member.JoinedAt,
		member.UpdatedAt,
	)
	if err != nil {
		logger.Errorf(ctx, "failed to create organization member: %v", err)
		return domain.OrganizationMember{}, err
	}

	return created, nil
}

// GetOrganizationMember retrieves a membership by user and organization
func (r *PGXRepository) GetOrganizationMember(ctx context.Context, userID, organizationID domain.ID) (*domain.OrganizationMember, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.GetOrganizationMember")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, organization_id, user_id, email, role, status, joined_at, updated_at
        FROM organization_members
        WHERE user_id = $1 AND organization_id = $2
    `

	var member domain.OrganizationMember
	err := pgxscan.Get(ctx, engine, &member, query, uuidToPgtype(userID), uuidToPgtype(organizationID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to get organization member: %v", err)
		return nil, err
	}

	return &member, nil
}

// UpdateOrganizationMember updates an organization membership
func (r *PGXRepository) UpdateOrganizationMember(ctx context.Context, member domain.OrganizationMember) (domain.OrganizationMember, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateOrganizationMember")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        UPDATE organization_members
        SET email = $2, role = $3, status = $4, updated_at = $5
        WHERE id = $1
        RETURNING id, organization_id, user_id, email, role, status, joined_at, updated_at
    `

	var updated domain.OrganizationMember
	err := pgxscan.Get(ctx, engine, &updated, query,
		uuidToPgtype(member.ID),
		member.Email,
		member.Role,
		member.Status,
		member.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OrganizationMember{}, domain.ErrNotFound
		}
		logger.Errorf(ctx, "failed to update organization member: %v", err)
		return domain.OrganizationMember{}, err
	}

	return updated, nil
}

// ListOrganizationMembersByUser lists all organization memberships for a user
func (r *PGXRepository) ListOrganizationMembersByUser(ctx context.Context, userID domain.ID) ([]*domain.OrganizationMember, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListOrganizationMembersByUser")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `
        SELECT id, organization_id, user_id, email, role, status, joined_at, updated_at
        FROM organization_members
        WHERE user_id = $1
        ORDER BY joined_at DESC
    `

	var members []*domain.OrganizationMember
	err := pgxscan.Select(ctx, engine, &members, query, uuidToPgtype(userID))
	if err != nil {
		logger.Errorf(ctx, "failed to list organization members by user: %v", err)
		return nil, err
	}

	return members, nil
}
