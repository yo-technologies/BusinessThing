package auth

import (
	"context"
	"errors"
	"strconv"

	"core-service/internal/domain"
	"core-service/internal/jwt"
	"core-service/internal/logger"
	"core-service/internal/telegram"
)

type userRepository interface {
	GetUserByTelegramID(ctx context.Context, telegramID string) (domain.User, error)
	GetUser(ctx context.Context, id domain.ID) (domain.User, error)
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetOrganizationMember(ctx context.Context, organizationID, userID domain.ID) (*domain.OrganizationMember, error)
	ListOrganizationMembersByUser(ctx context.Context, userID domain.ID) ([]*domain.OrganizationMember, error)
}

// Service handles authentication flows.
type Service struct {
	users     userRepository
	jwt       *jwt.Provider
	validator *telegram.Validator
}

func New(users userRepository, jwtProvider *jwt.Provider, validator *telegram.Validator) *Service {
	return &Service{
		users:     users,
		jwt:       jwtProvider,
		validator: validator,
	}
}

// AuthenticateWithTelegram validates Telegram init data and returns JWT token.
// If user is new or hasn't completed registration, returns is_new_user=true flag without token.
// If user has completed registration, returns token with all active organizations (may be empty list).
func (s *Service) AuthenticateWithTelegram(ctx context.Context, initData string) (string, *domain.User, bool, error) {
	data, err := s.validator.Validate(initData)
	if err != nil {
		logger.Warnf(ctx, "telegram init data validation failed: %v", err)
		return "", nil, false, domain.ErrUnauthorized
	}

	telegramID := strconv.FormatInt(data.User.ID, 10)
	user, err := s.users.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Создаем нового пользователя
			logger.Infof(ctx, "creating new user with telegram ID: %s", telegramID)
			newUser := domain.NewUser(telegramID)
			created, err := s.users.CreateUser(ctx, newUser)
			if err != nil {
				logger.Errorf(ctx, "failed to create new user: %v", err)
				return "", nil, false, err
			}

			// Для нового пользователя не генерируем токен - он должен завершить регистрацию
			logger.Infof(ctx, "new user created, registration incomplete: %s", created.ID)
			return "", &created, true, nil
		}
		logger.Errorf(ctx, "failed to get user by telegram ID: %v", err)
		return "", nil, false, err
	}

	// Проверяем, завершена ли регистрация
	if !user.RegistrationCompleted {
		logger.Infof(ctx, "user %s found but registration not completed", user.ID)
		return "", &user, true, nil
	}

	// Существующий пользователь с завершенной регистрацией - берем все его организации
	memberships, err := s.users.ListOrganizationMembersByUser(ctx, user.ID)
	if err != nil {
		logger.Errorf(ctx, "failed to get user memberships: %v", err)
		return "", nil, false, err
	}

	// Фильтруем только активные membership
	activeMemberships := make([]*domain.OrganizationMember, 0)
	for _, m := range memberships {
		if m.Status == domain.UserStatusActive {
			activeMemberships = append(activeMemberships, m)
		}
	}

	// Генерируем токен со всеми организациями (может быть пустой список)
	token, err := s.jwt.GenerateAccessToken(ctx, user.ID, activeMemberships)
	if err != nil {
		logger.Errorf(ctx, "failed to generate token: %v", err)
		return "", nil, false, err
	}

	if len(activeMemberships) == 0 {
		logger.Infof(ctx, "user %s authenticated without organizations", user.ID)
	} else {
		logger.Infof(ctx, "user %s authenticated with %d organizations", user.ID, len(activeMemberships))
	}
	return token, &user, false, nil
}

// CompleteRegistration завершает регистрацию нового пользователя (добавление ФИО).
func (s *Service) CompleteRegistration(ctx context.Context, userID domain.ID, firstName, lastName string) (*domain.User, error) {
	user, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return nil, domain.NewNotFoundError("user not found")
	}

	user.CompleteProfile(firstName, lastName)

	updated, err := s.users.UpdateUser(ctx, user)
	if err != nil {
		logger.Errorf(ctx, "failed to update user profile: %v", err)
		return nil, err
	}

	return &updated, nil
}
