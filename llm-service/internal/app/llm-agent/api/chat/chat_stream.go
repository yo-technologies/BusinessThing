package chat

import (
	"fmt"
	"sync/atomic"

	"llm-service/internal/app/interceptors"
	"llm-service/internal/app/mappers"
	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"
	"llm-service/internal/logger"
	desc "llm-service/pkg/agent"

	"github.com/guregu/null/v6"
	"github.com/opentracing/opentracing-go"
)

func (s *Service) ChatStream(stream desc.ChatService_ChatStreamServer) error {
	span, ctx := opentracing.StartSpanFromContext(stream.Context(), "api.chat.ChatStream")
	defer span.Finish()

	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		return err
	}

	var stopFlag atomic.Bool

	callbacks := dto.ChatStreamCallbacks{
		OnStatus: func(status string) error {
			return stream.Send(&desc.ChatStreamResponse{Event: &desc.ChatStreamResponse_Status{Status: &desc.StatusEvent{Status: status}}})
		},
		OnUsage: func(u dto.ChatUsage) error {
			return stream.Send(&desc.ChatStreamResponse{Event: &desc.ChatStreamResponse_Usage{Usage: &desc.UsageEvent{Usage: mappers.ToProtoUsage(u)}}})
		},
		OnContentDelta: func(delta string) error {
			return stream.Send(&desc.ChatStreamResponse{Event: &desc.ChatStreamResponse_ContentDelta{ContentDelta: &desc.ContentDeltaEvent{Delta: delta}}})
		},
		OnToolEvent: func(e dto.ToolEvent) error {
			return stream.Send(&desc.ChatStreamResponse{Event: &desc.ChatStreamResponse_ToolEvent{ToolEvent: mappers.ToProtoToolEvent(e)}})
		},
		OnFinalResponse: func(result dto.ChatCompletionDTO) error {
			return stream.Send(&desc.ChatStreamResponse{Event: &desc.ChatStreamResponse_Final{Final: &desc.FinalEvent{Result: mappers.ToProtoCompletion(result)}}})
		},
		OnError: func(e error) error {
			return stream.Send(&desc.ChatStreamResponse{Event: &desc.ChatStreamResponse_Error{Error: &desc.ErrorEvent{Message: e.Error()}}})
		},
		ShouldStop: func() bool { return stopFlag.CompareAndSwap(true, false) },
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		switch p := req.GetPayload().(type) {
		case *desc.ChatStreamRequest_SendMessage:
			send := p.SendMessage

			if err := send.Validate(); err != nil {
				_ = callbacks.OnError(fmt.Errorf("%w: %w", err, domain.ErrInvalidArgument))
				continue
			}

			chatReq := dto.SendChatMessageRequest{}
			chatReq.ChatID = null.StringFromPtr(send.ChatId)
			chatReq.Content = send.GetContent()

			_, err := s.chatService.SendChatMessageStream(ctx, userID, chatReq, callbacks)
			if err != nil {
				logger.Errorf(ctx, "SendChatMessageStream error: %v", err)
				// Error already sent via callbacks.OnError, continue listening
			}

		case *desc.ChatStreamRequest_ToolDecision:
			td := p.ToolDecision
			chatID, err := domain.ParseID(td.GetChatId())
			if err != nil {
				_ = callbacks.OnError(fmt.Errorf("%w: %w", err, domain.ErrInvalidArgument))
				continue
			}

			_, err = s.chatService.HandleToolCallDecisionStream(ctx, userID, chatID, td.GetToolCallId(), td.GetApprove(), callbacks)
			if err != nil {
				_ = callbacks.OnError(fmt.Errorf("failed to handle tool call decision: %w", err))
				logger.Errorf(ctx, "HandleToolCallDecisionStream error: %v", err)
				// Error already sent via callbacks.OnError, continue listening
			}

		case *desc.ChatStreamRequest_Stop:
			stopFlag.Store(true)
		default:
			logger.Warn(ctx, "received empty/unknown payload in ChatStream")
		}
	}
}
