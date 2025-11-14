package user

import (
	"context"
	pb "core-service/bin/core/api/core"
	"core-service/internal/app/interceptors"
	"core-service/internal/domain"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedUserServiceServer
	userService UserService
}

type UserService interface {
	InviteUser(ctx context.Context, organizationID domain.ID, email string, role domain.UserRole, invitedBy domain.ID) (*domain.Invitation, error)
	AcceptInvitation(ctx context.Context, userID domain.ID, token string) (*domain.User, error)
	ListUsersByOrganization(ctx context.Context, organizationID domain.ID) ([]*domain.User, error)
	GetUser(ctx context.Context, id domain.ID) (*domain.User, error)
	UpdateUserRole(ctx context.Context, id domain.ID, role domain.UserRole) (*domain.User, error)
	DeactivateUser(ctx context.Context, id domain.ID) error
}

func NewService(userService UserService) *Service {
	return &Service{
		userService: userService,
	}
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

func userRoleFromProto(role pb.UserRole) domain.UserRole {
	switch role {
	case pb.UserRole_USER_ROLE_ADMIN:
		return domain.UserRoleAdmin
	case pb.UserRole_USER_ROLE_EMPLOYEE:
		return domain.UserRoleEmployee
	default:
		return domain.UserRoleEmployee
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

func (s *Service) InviteUser(ctx context.Context, req *pb.InviteUserRequest) (*pb.InviteUserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.InviteUser")
	defer span.Finish()

	orgID, err := domain.ParseID(req.OrganizationId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	invitedBy, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	role := userRoleFromProto(req.Role)

	invitation, err := s.userService.InviteUser(ctx, orgID, req.Email, role, invitedBy)
	if err != nil {
		return nil, err
	}

	return &pb.InviteUserResponse{
		InvitationToken: invitation.Token,
		ExpiresAt:       timestamppb.New(invitation.ExpiresAt),
	}, nil
}

func (s *Service) AcceptInvitation(ctx context.Context, req *pb.AcceptInvitationRequest) (*pb.AcceptInvitationResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.AcceptInvitation")
	defer span.Finish()

	// Получаем ID пользователя из контекста (JWT)
	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.userService.AcceptInvitation(ctx, userID, req.Token)
	if err != nil {
		return nil, err
	}

	return &pb.AcceptInvitationResponse{
		User: userToProto(user),
	}, nil
}

func (s *Service) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.ListUsers")
	defer span.Finish()

	orgID, err := domain.ParseID(req.OrganizationId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	users, err := s.userService.ListUsersByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	pbUsers := make([]*pb.User, 0, len(users))
	for _, user := range users {
		pbUsers = append(pbUsers, userToProto(user))
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
	}, nil
}

func (s *Service) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.GetUser")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	user, err := s.userService.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserResponse{
		User: userToProto(user),
	}, nil
}

func (s *Service) UpdateUserRole(ctx context.Context, req *pb.UpdateUserRoleRequest) (*pb.UpdateUserRoleResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.UpdateUserRole")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	role := userRoleFromProto(req.Role)

	user, err := s.userService.UpdateUserRole(ctx, id, role)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateUserRoleResponse{
		User: userToProto(user),
	}, nil
}

func (s *Service) DeactivateUser(ctx context.Context, req *pb.DeactivateUserRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.DeactivateUser")
	defer span.Finish()

	id, err := domain.ParseID(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	err = s.userService.DeactivateUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
