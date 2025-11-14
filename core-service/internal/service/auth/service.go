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

// Result contains authentication output payload.
type Result struct {
    Token string
    User  domain.User
}

// AuthenticateWithTelegram validates Telegram init data and returns JWT token.
func (s *Service) AuthenticateWithTelegram(ctx context.Context, initData string) (*Result, error) {
    data, err := s.validator.Validate(initData)
    if err != nil {
        return nil, domain.ErrUnauthorized
    }

    telegramID := strconv.FormatInt(data.User.ID, 10)
    user, err := s.users.GetUserByTelegramID(ctx, telegramID)
    if err != nil {
        if errors.Is(err, domain.ErrNotFound) {
            return nil, domain.ErrUnauthorized
        }
        return nil, err
    }

    token, err := s.jwt.GenerateAccessToken(ctx, user.ID, user.OrganizationID, user.Role)
    if err != nil {
        logger.Errorf(ctx, "failed to generate token: %v", err)
        return nil, err
    }

    return &Result{Token: token, User: user}, nil
}

