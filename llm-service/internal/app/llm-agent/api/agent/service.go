package agent

import (
	"context"

	"llm-service/internal/domain"
	"llm-service/internal/logger"
	"llm-service/internal/service"
	pb "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	chatManager   service.ChatManager
	agentExecutor service.AgentExecutor
	quotaService  QuotaService

	pb.UnimplementedAgentServiceServer
}

type QuotaService interface {
	GetLimits(ctx context.Context, userID domain.ID) (domain.LLMLimits, error)
}

func NewService(
	chatManager service.ChatManager,
	agentExecutor service.AgentExecutor,
	quotaService QuotaService,
) *Service {
	return &Service{
		chatManager:   chatManager,
		agentExecutor: agentExecutor,
		quotaService:  quotaService,
	}
}

func (s *Service) GetLLMLimits(ctx context.Context, _ *emptypb.Empty) (*pb.GetLLMLimitsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.agent.GetLLMLimits")
	defer span.Finish()

	userID, ok := ctx.Value("user_id").(domain.ID)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	limits, err := s.quotaService.GetLimits(ctx, userID)
	if err != nil {
		logger.Error(ctx, "failed to get limits", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.GetLLMLimitsResponse{
		DailyLimit: int32(limits.DailyLimit),
		Used:       int32(limits.Used),
		Remaining:  int32(limits.Remaining()),
	}, nil
}
