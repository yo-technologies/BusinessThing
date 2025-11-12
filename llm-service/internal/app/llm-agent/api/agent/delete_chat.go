package agent

import (
	"context"
	"fmt"
	"llm-service/internal/app/interceptors"
	"llm-service/internal/domain"
	desc "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) DeleteChat(ctx context.Context, req *desc.DeleteChatRequest) (*emptypb.Empty, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.agent.DeleteChat")
	defer span.Finish()

	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	chatID, err := domain.ParseID(req.ChatId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid chat_id")
	}

	orgID, err := domain.ParseID(req.GetOrgId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid organization_id")
	}

	err = s.chatManager.DeleteChat(ctx, chatID, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete chat: %w", err)
	}

	return &emptypb.Empty{}, nil
}
