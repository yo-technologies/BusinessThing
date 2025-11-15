package auth

import (
	"context"
	"core-service/internal/app/interceptors"
	"core-service/internal/domain"
	pb "core-service/pkg/core"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedAuthServiceServer
	authService AuthService
}

type AuthService interface {
	AuthenticateWithTelegram(ctx context.Context, initData string) (string, *domain.User, bool, error)
	CompleteRegistration(ctx context.Context, userID domain.ID, firstName, lastName string) (*domain.User, error)
	RefreshToken(ctx context.Context, userID domain.ID) (string, error)
}

func NewService(authService AuthService) *Service {
	return &Service{
		authService: authService,
	}
}

func (s *Service) AuthenticateWithTelegram(ctx context.Context, req *pb.AuthenticateWithTelegramRequest) (*pb.AuthenticateWithTelegramResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.auth.AuthenticateWithTelegram")
	defer span.Finish()

	token, user, isNewUser, err := s.authService.AuthenticateWithTelegram(ctx, req.GetInitData())
	if err != nil {
		return nil, err
	}

	return &pb.AuthenticateWithTelegramResponse{
		AccessToken: token,
		User:        userToProto(user),
		IsNewUser:   isNewUser,
	}, nil
}

func (s *Service) CompleteRegistration(ctx context.Context, req *pb.CompleteRegistrationRequest) (*pb.CompleteRegistrationResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.auth.CompleteRegistration")
	defer span.Finish()

	userID, err := domain.ParseID(req.GetUserId())
	if err != nil {
		return nil, domain.NewInvalidArgumentError("invalid user_id")
	}

	user, err := s.authService.CompleteRegistration(ctx, userID, req.FirstName, req.LastName)
	if err != nil {
		return nil, err
	}

	return &pb.CompleteRegistrationResponse{
		User: userToProto(user),
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.auth.RefreshToken")
	defer span.Finish()

	// Получаем userID из контекста (из JWT токена)
	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	token, err := s.authService.RefreshToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &pb.RefreshTokenResponse{
		AccessToken: token,
	}, nil
}

func userToProto(user *domain.User) *pb.User {
	return &pb.User{
		Id:         user.ID.String(),
		TelegramId: user.TelegramID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		CreatedAt:  timestamppb.New(user.CreatedAt),
		UpdatedAt:  timestamppb.New(user.UpdatedAt),
	}
}
