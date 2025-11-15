package agent

import (
	"context"
	"fmt"

	"llm-service/internal/app/interceptors"
	"llm-service/internal/domain"
	"llm-service/internal/logger"
	"llm-service/internal/service"
	pb "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
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

	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		logger.Warnf(ctx, "unauthorized access to GetLLMLimits: %v", err)
		return nil, domain.ErrUnauthorized
	}

	limits, err := s.quotaService.GetLimits(ctx, userID)
	if err != nil {
		logger.Error(ctx, "failed to get limits", "error", err)
		return nil, fmt.Errorf("failed to get limits: %w", err)
	}

	return &pb.GetLLMLimitsResponse{
		DailyLimit: int32(limits.DailyLimit),
		Used:       int32(limits.Used),
		Remaining:  int32(limits.Remaining()),
	}, nil
}
