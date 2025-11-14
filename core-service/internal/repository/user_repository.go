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
        INSERT INTO users (id, organization_id, email, telegram_id, first_name, last_name, role, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, organization_id, email, telegram_id, first_name, last_name, role, status, created_at, updated_at
    `

	var created domain.User
	err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(user.ID),
		uuidToPgtype(user.OrganizationID),
		user.Email,
		user.TelegramID,
		user.FirstName,
		user.LastName,
		user.Role,
		user.Status,
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
        SELECT id, organization_id, email, telegram_id, first_name, last_name, role, status, created_at, updated_at
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
        SELECT id, organization_id, email, telegram_id, first_name, last_name, role, status, created_at, updated_at
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
func (r *PGXRepository) ListUsers(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.User, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.ListUsers")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)

	// Get total count
	countQuery := `SELECT COUNT(*) FROM users WHERE organization_id = $1`
	var total int
	err := pgxscan.Get(ctx, engine, &total, countQuery, uuidToPgtype(organizationID))
	if err != nil {
		logger.Errorf(ctx, "failed to count users: %v", err)
		return nil, 0, err
	}

	// Get users
	query := `
        SELECT id, organization_id, email, telegram_id, first_name, last_name, role, status, created_at, updated_at
        FROM users
        WHERE organization_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `

	var users []domain.User
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
        SET email = $2, telegram_id = $3, first_name = $4, last_name = $5, role = $6, status = $7, updated_at = $8
        WHERE id = $1
        RETURNING id, organization_id, email, telegram_id, first_name, last_name, role, status, created_at, updated_at
    `

	var updated domain.User
	err := pgxscan.Get(ctx, engine, &updated, query,
		uuidToPgtype(user.ID),
		user.Email,
		user.TelegramID,
		user.FirstName,
		user.LastName,
		user.Role,
		user.Status,
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

// UpdateUserRole updates a user's role
func (r *PGXRepository) UpdateUserRole(ctx context.Context, id domain.ID, role domain.UserRole) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "repository.UpdateUserRole")
	defer span.Finish()

	engine := r.engineFactory.Get(ctx)
	query := `UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1`

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
	query := `UPDATE users SET status = 'inactive', updated_at = NOW() WHERE id = $1`

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
        INSERT INTO invitations (id, organization_id, email, token, role, expires_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, organization_id, email, token, role, expires_at, used_at, created_at
    `

	var created domain.Invitation
	err := pgxscan.Get(ctx, engine, &created, query,
		uuidToPgtype(invitation.ID),
		uuidToPgtype(invitation.OrganizationID),
		invitation.Email,
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
        SELECT id, organization_id, email, token, role, expires_at, used_at, created_at
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
