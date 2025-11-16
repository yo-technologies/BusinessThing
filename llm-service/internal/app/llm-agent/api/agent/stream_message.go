package agent

import (
	"encoding/json"
	"fmt"
	"io"

	"llm-service/internal/app/interceptors"
	"llm-service/internal/app/llm-agent/mappers"
	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"
	"llm-service/internal/logger"
	desc "llm-service/pkg/agent"

	"github.com/samber/lo"
)

// Бидунаправленный стрим: принимаем входящие события клиента и стримим ответы
func (s *Service) StreamMessage(stream desc.AgentService_StreamMessageServer) error {
	ctx := stream.Context()

	// Получаем userID из контекста
	userID, err := interceptors.UserIDFromContext(ctx)
	if err != nil {
		logger.Errorf(ctx, "StreamMessage: failed to get userID from context: %v", err)
		return err
	}

	logger.Infof(ctx, "StreamMessage: started for userID=%s", userID)

	// Адаптер для отправки событий
	streamAdapter := &streamAdapter{stream: stream}

	for {
		req, recvErr := stream.Recv()
		if recvErr == io.EOF {
			logger.Info(ctx, "StreamMessage: client closed stream (EOF)")
			return nil
		}
		if recvErr != nil {
			logger.Errorf(ctx, "StreamMessage: error receiving from stream: %v", recvErr)
			return recvErr
		}

		logger.Infof(ctx, "StreamMessage: received request: %+v", req)

		// Обрабатываем oneof payload
		if nm := req.GetNewMessage(); nm != nil {
			logger.Infof(ctx, "StreamMessage: processing new_message: chatId=%s, orgId=%s, content_len=%d",
				lo.FromPtr(nm.ChatId), nm.GetOrgId(), len(nm.GetContent()))
			// Парсим chat_id если указан
			var chatID *domain.ID
			if nm.ChatId != nil && *nm.ChatId != "" {
				id, err := domain.ParseID(*nm.ChatId)
				if err != nil {
					logger.Errorf(ctx, "StreamMessage: invalid chat ID: %v", err)
					// Сообщаем об ошибке клиенту и продолжаем принимать следующие события
					if sendErr := streamAdapter.SendError(fmt.Errorf("invalid chat ID: %w", err)); sendErr != nil {
						return sendErr
					}
					continue
				}
				chatID = &id
				logger.Infof(ctx, "StreamMessage: parsed chatID=%s", id)
			} else {
				logger.Info(ctx, "StreamMessage: no chatID provided, will create new chat")
			}

			// Парсим org_id
			orgID, err := domain.ParseID(nm.GetOrgId())
			if err != nil {
				logger.Errorf(ctx, "StreamMessage: invalid org ID: %v", err)
				if sendErr := streamAdapter.SendError(fmt.Errorf("invalid org ID: %w", err)); sendErr != nil {
					return sendErr
				}
				continue
			}
			logger.Infof(ctx, "StreamMessage: parsed orgID=%s", orgID)

			// Формируем DTO и запускаем обработку сообщения
			executeDTO := dto.SendMessageDTO{
				ChatID:  chatID,
				UserID:  userID,
				OrgID:   orgID,
				Content: nm.GetContent(),
			}

			logger.Infof(ctx, "StreamMessage: calling agentExecutor.SendMessageStream")

			if err := s.agentExecutor.SendMessageStream(ctx, executeDTO, streamAdapter); err != nil {
				logger.Errorf(ctx, "StreamMessage: agentExecutor failed: %v", err)
				// Передаем ошибку через стрим и продолжаем приёмы
				if sendErr := streamAdapter.SendError(fmt.Errorf("failed to execute agent: %w", err)); sendErr != nil {
					return sendErr
				}
				continue
			}
			logger.Info(ctx, "StreamMessage: agentExecutor completed successfully")
			continue
		}

		// Неподдерживаемый тип запроса
		logger.Errorf(ctx, "StreamMessage: unsupported request payload: %+v", req)
		if sendErr := streamAdapter.SendError(fmt.Errorf("unsupported request payload")); sendErr != nil {
			return sendErr
		}
	}
}

// streamAdapter адаптер для передачи результатов в gRPC stream
type streamAdapter struct {
	stream desc.AgentService_StreamMessageServer
}

func (a *streamAdapter) SendChunk(content string) error {
	return a.stream.Send(&desc.StreamMessageResponse{
		Event: &desc.StreamMessageResponse_Chunk{
			Chunk: &desc.MessageChunk{
				Content: content,
			},
		},
	})
}

func (a *streamAdapter) SendMessage(message *domain.Message) error {
	return a.stream.Send(&desc.StreamMessageResponse{
		Event: &desc.StreamMessageResponse_Message{
			Message: mappers.DomainMessageToProto(message),
		},
	})
}

func (a *streamAdapter) SendToolCall(toolCall *domain.ToolCall) error {
	args, err := json.Marshal(toolCall.Arguments)
	if err != nil {
		return err
	}

	return a.stream.Send(&desc.StreamMessageResponse{
		Event: &desc.StreamMessageResponse_ToolCall{
			ToolCall: &desc.ToolCallEvent{
				ToolName:  toolCall.Name,
				Arguments: string(args),
				Status:    string(toolCall.Status),
			},
		},
	})
}

func (a *streamAdapter) SendUsage(usage *dto.ChatUsageDTO) error {
	return a.stream.Send(&desc.StreamMessageResponse{
		Event: &desc.StreamMessageResponse_Usage{
			Usage: &desc.UsageEvent{
				Usage: &desc.ChatUsage{
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

	sendErr := a.stream.Send(&desc.StreamMessageResponse{
		Event: &desc.StreamMessageResponse_Error{
			Error: &desc.ErrorEvent{
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
	return a.stream.Send(&desc.StreamMessageResponse{
		Event: &desc.StreamMessageResponse_Chat{
			Chat: &desc.ChatEvent{
				ChatId:   chat.ID.String(),
				ChatName: chat.Title,
			},
		},
	})
}

func (a *streamAdapter) SendFinal(chat *domain.Chat, messages []*domain.Message) error {
	pbMessages := make([]*desc.Message, 0, len(messages))
	for _, msg := range messages {
		pbMessages = append(pbMessages, mappers.DomainMessageToProto(msg))
	}

	return a.stream.Send(&desc.StreamMessageResponse{
		Event: &desc.StreamMessageResponse_Final{
			Final: &desc.FinalEvent{
				Chat:          mappers.DomainChatToProto(chat),
				Messages:      pbMessages,
				TotalMessages: int32(len(messages)),
			},
		},
	})
}
