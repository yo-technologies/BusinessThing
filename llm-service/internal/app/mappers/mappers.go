package mappers

import (
	"encoding/json"
	"time"

	"llm-service/internal/domain"
	"llm-service/internal/domain/dto"
	desc "llm-service/pkg/agent"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoChat(c domain.Chat) *desc.Chat {
	return &desc.Chat{
		Id:        c.ID.String(),
		Title:     c.Title,
		CreatedAt: timestamppb.New(c.CreatedAt),
	}
}

func ToProtoMessages(msgs []domain.ChatMessage) []*desc.ChatMessage {
	out := make([]*desc.ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, ToProtoMessage(m))
	}
	return out
}

func ToProtoMessage(m domain.ChatMessage) *desc.ChatMessage {
	pb := &desc.ChatMessage{
		Id:        m.ID.String(),
		ChatId:    m.ChatID.String(),
		Role:      toProtoRole(m.Role),
		Content:   m.Content,
		ToolState: toProtoToolState(m.ToolState),
		CreatedAt: timestamppb.New(m.CreatedAt),
	}

	if m.ToolName.Valid {
		pb.ToolName = m.ToolName.Ptr()
	}
	if m.ToolCallID.Valid {
		pb.ToolCallId = m.ToolCallID.Ptr()
	}
	if len(m.ToolArguments) > 0 {
		var mp map[string]any
		if err := json.Unmarshal(m.ToolArguments, &mp); err == nil {
			if st, err := structpb.NewStruct(mp); err == nil {
				pb.ToolArguments = st
			}
		}
	}
	if m.TokenUsage.Valid {
		pb.TokenUsage = m.TokenUsage.Ptr()
	}
	if m.Error.Valid {
		pb.Error = m.Error.Ptr()
	}
	return pb
}

func toProtoRole(r domain.ChatMessageRole) desc.ChatMessageRole {
	switch r {
	case domain.ChatMessageRoleUser:
		return desc.ChatMessageRole_ROLE_USER
	case domain.ChatMessageRoleAssistant:
		return desc.ChatMessageRole_ROLE_ASSISTANT
	case domain.ChatMessageRoleTool:
		return desc.ChatMessageRole_ROLE_TOOL
	case domain.ChatMessageRoleSystem:
		return desc.ChatMessageRole_ROLE_SYSTEM
	default:
		return desc.ChatMessageRole_ROLE_UNSPECIFIED
	}
}

func toProtoToolState(s domain.ToolState) desc.ToolState {
	switch s {
	case domain.ToolStateProposed:
		return desc.ToolState_TOOL_STATE_PROPOSED
	case domain.ToolStateRejected:
		return desc.ToolState_TOOL_STATE_REJECTED
	case domain.ToolStateApproved:
		return desc.ToolState_TOOL_STATE_APPROVED
	default:
		return desc.ToolState_TOOL_STATE_UNSPECIFIED
	}
}

func ToProtoUsage(u dto.ChatUsage) *desc.ChatUsage {
	return &desc.ChatUsage{PromptTokens: int32(u.PromptTokens), CompletionTokens: int32(u.CompletionTokens), TotalTokens: int32(u.TotalTokens)}
}

func ToProtoCompletion(r dto.ChatCompletionDTO) *desc.ChatCompletion {
	var usage *desc.ChatUsage
	if r.Usage != nil {
		usage = ToProtoUsage(*r.Usage)
	}
	return &desc.ChatCompletion{
		Chat:     ToProtoChat(r.Chat),
		Messages: ToProtoMessages(r.Messages),
		Usage:    usage,
	}
}

func ToProtoToolEvent(e dto.ToolEvent) *desc.ToolEvent {
	return &desc.ToolEvent{
		ToolName:   e.ToolName,
		ToolCallId: e.ToolCallID,
		ArgsJson:   e.ArgsJSON,
		State:      toProtoToolEventState(e.State),
		Error:      e.Error,
	}
}

func toProtoToolEventState(s dto.ToolEventState) desc.ToolEventState {
	switch s {
	case dto.ToolInvoking:
		return desc.ToolEventState_TOOL_EVENT_INVOKING
	case dto.ToolCompleted:
		return desc.ToolEventState_TOOL_EVENT_COMPLETED
	case dto.ToolError:
		return desc.ToolEventState_TOOL_EVENT_ERROR
	default:
		return desc.ToolEventState_TOOL_EVENT_STATE_UNSPECIFIED
	}
}

func ToProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
