package user

import (
	"context"
	"core-service/internal/app/interceptors"
	"core-service/internal/domain"
	pb "core-service/pkg/core"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	pb.UnimplementedUserServiceServer
	userService UserService
	miniAppURL  string
}

type UserService interface {
	InviteUser(ctx context.Context, organizationID domain.ID, role domain.UserRole, invitedBy domain.ID) (*domain.Invitation, error)
	AcceptInvitation(ctx context.Context, userID domain.ID, token string) (*domain.User, error)
	ListUsersByOrganization(ctx context.Context, organizationID domain.ID) ([]*domain.UserWithMembership, error)
	ListInvitations(ctx context.Context, organizationID domain.ID, limit, offset int) ([]domain.Invitation, int, error)
	GetUser(ctx context.Context, id domain.ID) (*domain.User, error)
	UpdateUserRole(ctx context.Context, id domain.ID, role domain.UserRole) (*domain.User, error)
	DeactivateUser(ctx context.Context, id domain.ID) error
}

func NewService(userService UserService, miniAppURL string) *Service {
	return &Service{
		userService: userService,
		miniAppURL:  miniAppURL,
	}
}

func userToProto(user *domain.UserWithMembership) *pb.User {
	return &pb.User{
		Id:         user.User.ID.String(),
		TelegramId: user.User.TelegramID,
		FirstName:  user.User.FirstName,
		LastName:   user.User.LastName,
		Role:       userRoleToProto(user.OrganizationMember.Role),
		Status:     userStatusToProto(user.OrganizationMember.Status),
		CreatedAt:  timestamppb.New(user.User.CreatedAt),
		UpdatedAt:  timestamppb.New(user.User.UpdatedAt),
	}
}

func simpleUserToProto(user *domain.User) *pb.User {
	return &pb.User{
		Id:         user.ID.String(),
		TelegramId: user.TelegramID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Role:       pb.UserRole_USER_ROLE_UNSPECIFIED,
		Status:     pb.UserStatus_USER_STATUS_UNSPECIFIED,
		CreatedAt:  timestamppb.New(user.CreatedAt),
		UpdatedAt:  timestamppb.New(user.UpdatedAt),
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

	invitation, err := s.userService.InviteUser(ctx, orgID, role, invitedBy)
	if err != nil {
		return nil, err
	}

	// Generate Telegram Mini App URL with invitation token
	invitationURL := fmt.Sprintf("%s?startapp=invitation_%s", s.miniAppURL, invitation.Token)

	return &pb.InviteUserResponse{
		InvitationToken: invitation.Token,
		InvitationUrl:   invitationURL,
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
		User: simpleUserToProto(user),
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
		User: simpleUserToProto(user),
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
		User: simpleUserToProto(user),
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

func (s *Service) ListInvitations(ctx context.Context, req *pb.ListInvitationsRequest) (*pb.ListInvitationsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.ListInvitations")
	defer span.Finish()

	orgID, err := domain.ParseID(req.OrganizationId)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}

	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	invitations, total, err := s.userService.ListInvitations(ctx, orgID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	pbInvitations := make([]*pb.Invitation, len(invitations))
	for i, inv := range invitations {
		pbInvitations[i] = invitationToProto(&inv)
	}

	return &pb.ListInvitationsResponse{
		Invitations: pbInvitations,
		Total:       int32(total),
		Page:        int32(page),
	}, nil
}

func invitationToProto(inv *domain.Invitation) *pb.Invitation {
	result := &pb.Invitation{
		Id:             inv.ID.String(),
		OrganizationId: inv.OrganizationID.String(),
		Token:          inv.Token,
		Role:           userRoleToProto(inv.Role),
		ExpiresAt:      timestamppb.New(inv.ExpiresAt),
		CreatedAt:      timestamppb.New(inv.CreatedAt),
	}
	if inv.UsedAt != nil {
		result.UsedAt = timestamppb.New(*inv.UsedAt)
	}
	return result
}
