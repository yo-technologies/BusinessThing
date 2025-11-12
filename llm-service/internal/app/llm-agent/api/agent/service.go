package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"llm-service/internal/app/interceptors"
	"llm-service/internal/app/llm-agent/mappers"
	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"
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

func (s *Service) StreamMessage(req *pb.StreamMessageRequest, stream pb.AgentService_StreamMessageServer) error {
	ctx := stream.Context()

	// Получаем userID из контекста
	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Парсим chat_id если указан
	var chatID *domain.ID
	if req.ChatId != nil && *req.ChatId != "" {
		id, err := domain.ParseID(*req.ChatId)
		if err != nil {
			return fmt.Errorf("invalid chat ID: %w", err)
		}
		chatID = &id
	}

	// Создаем DTO для выполнения
	executeDTO := dto.SendMessageDTO{
		ChatID:  chatID,
		UserID:  userID,
		Content: req.Content,
	}

	// Создаем stream adapter
	streamAdapter := &streamAdapter{
		stream: stream,
	}

	// Выполняем агента
	err = s.agentExecutor.SendMessageStream(ctx, executeDTO, streamAdapter)
	if err != nil {
		return fmt.Errorf("failed to execute agent: %w", err)
	}

	return nil
}

// streamAdapter адаптер для передачи результатов в gRPC stream
type streamAdapter struct {
	stream pb.AgentService_StreamMessageServer
}

func (a *streamAdapter) SendChunk(content string) error {
	return a.stream.Send(&pb.StreamMessageResponse{
		Event: &pb.StreamMessageResponse_Chunk{
			Chunk: &pb.MessageChunk{
				Content: content,
			},
		},
	})
}

func (a *streamAdapter) SendMessage(message *domain.Message) error {
	return a.stream.Send(&pb.StreamMessageResponse{
		Event: &pb.StreamMessageResponse_Message{
			Message: mappers.DomainMessageToProto(message),
		},
	})
}

func (a *streamAdapter) SendToolCall(toolCall *domain.ToolCall) error {
	args, err := json.Marshal(toolCall.Arguments)
	if err != nil {
		return err
	}

	return a.stream.Send(&pb.StreamMessageResponse{
		Event: &pb.StreamMessageResponse_ToolCall{
			ToolCall: &pb.ToolCallEvent{
				ToolName:  toolCall.Name,
				Arguments: string(args),
				Status:    string(toolCall.Status),
			},
		},
	})
}

func (a *streamAdapter) SendUsage(usage *dto.ChatUsageDTO) error {
	return a.stream.Send(&pb.StreamMessageResponse{
		Event: &pb.StreamMessageResponse_Usage{
			Usage: &pb.UsageEvent{
				Usage: &pb.ChatUsage{
					PromptTokens:     int32(usage.PromptTokens),
					CompletionTokens: int32(usage.CompletionTokens),
					TotalTokens:      int32(usage.TotalTokens),
				},
			},
		},
	})
}

func (a *streamAdapter) SendError(err error) error {
	code := "internal_error"
	message := err.Error()

	if domain.IsQuotaExceededError(err) {
		code = "quota_exceeded"
	} else if domain.IsNotFoundError(err) {
		code = "not_found"
	}

	sendErr := a.stream.Send(&pb.StreamMessageResponse{
		Event: &pb.StreamMessageResponse_Error{
			Error: &pb.ErrorEvent{
				Code:    code,
				Message: message,
			},
		},
	})

	if sendErr != nil {
		if sendErr == io.EOF {
			return nil
		}
		return sendErr
	}

	return err
}

func (a *streamAdapter) SendChat(chat *domain.Chat) error {
	return a.stream.Send(&pb.StreamMessageResponse{
		Event: &pb.StreamMessageResponse_Chat{
			Chat: &pb.ChatEvent{
				ChatId:   chat.ID.String(),
				ChatName: chat.Title,
			},
		},
	})
}
