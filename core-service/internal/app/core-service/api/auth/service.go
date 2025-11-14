package auth

import (
	"context"
	pb "core-service/bin/core/api/core"
	"core-service/internal/app/interceptors"
	"core-service/internal/domain"

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

	// Получаем ID пользователя из контекста (JWT)
	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.authService.CompleteRegistration(ctx, userID, req.FirstName, req.LastName)
	if err != nil {
		return nil, err
	}

	return &pb.CompleteRegistrationResponse{
		User: userToProto(user),
	}, nil
}

func userToProto(user *domain.User) *pb.User {
	return &pb.User{
		Id:             user.ID.String(),
		OrganizationId: user.OrganizationID.String(),
		Email:          user.Email,
		TelegramId:     user.TelegramID,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Role:           userRoleToProto(user.Role),
		Status:         userStatusToProto(user.Status),
		CreatedAt:      timestamppb.New(user.CreatedAt),
		UpdatedAt:      timestamppb.New(user.UpdatedAt),
	}
}

func userRoleToProto(role domain.UserRole) pb.UserRole {
	switch role {
	case domain.UserRoleAdmin:
		return pb.UserRole_USER_ROLE_ADMIN
	case domain.UserRoleEmployee:
		return pb.UserRole_USER_ROLE_EMPLOYEE
	default:
		return pb.UserRole_USER_ROLE_UNSPECIFIED
	}
}

func userStatusToProto(status domain.UserStatus) pb.UserStatus {
	switch status {
	case domain.UserStatusPending:
		return pb.UserStatus_USER_STATUS_PENDING
	case domain.UserStatusActive:
		return pb.UserStatus_USER_STATUS_ACTIVE
	case domain.UserStatusInactive:
		return pb.UserStatus_USER_STATUS_INACTIVE
	default:
		return pb.UserStatus_USER_STATUS_UNSPECIFIED
	}
}
