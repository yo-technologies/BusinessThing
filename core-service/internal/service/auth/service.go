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
// If user is new, creates a user with pending status and returns is_new_user=true flag.
func (s *Service) AuthenticateWithTelegram(ctx context.Context, initData string) (string, *domain.User, bool, error) {
	data, err := s.validator.Validate(initData)
	if err != nil {
		return "", nil, false, domain.ErrUnauthorized
	}

	telegramID := strconv.FormatInt(data.User.ID, 10)
	user, err := s.users.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Создаем нового пользователя со статусом pending
			newUser := domain.NewUser(telegramID)
			created, err := s.users.CreateUser(ctx, newUser)
			if err != nil {
				logger.Errorf(ctx, "failed to create new user: %v", err)
				return "", nil, false, err
			}

			// Для нового пользователя генерируем временный токен
			token, err := s.jwt.GenerateAccessToken(ctx, created.ID, created.OrganizationID, created.Role)
			if err != nil {
				logger.Errorf(ctx, "failed to generate token for new user: %v", err)
				return "", nil, false, err
			}

			return token, &created, true, nil
		}
		return "", nil, false, err
	}

	// Существующий пользователь
	token, err := s.jwt.GenerateAccessToken(ctx, user.ID, user.OrganizationID, user.Role)
	if err != nil {
		logger.Errorf(ctx, "failed to generate token: %v", err)
		return "", nil, false, err
	}

	return token, &user, false, nil
}

// CompleteRegistration завершает регистрацию нового пользователя (добавление ФИО).
func (s *Service) CompleteRegistration(ctx context.Context, userID domain.ID, firstName, lastName string) (*domain.User, error) {
	user, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return nil, domain.NewNotFoundError("user not found")
	}

	if user.Status != domain.UserStatusPending {
		return nil, domain.NewInvalidArgumentError("user registration already completed")
	}

	user.CompleteProfile(firstName, lastName)

	updated, err := s.users.UpdateUser(ctx, user)
	if err != nil {
		logger.Errorf(ctx, "failed to update user profile: %v", err)
		return nil, err
	}

	return &updated, nil
}
