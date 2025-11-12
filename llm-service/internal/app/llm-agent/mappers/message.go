package mappers

import (
	"encoding/json"
	"llm-service/internal/domain"
	pb "llm-service/pkg/agent"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// DomainMessageToProto конвертирует domain.Message в proto Message
func DomainMessageToProto(msg *domain.Message) *pb.Message {
	if msg == nil {
		return nil
	}

	var toolCallID string
	if msg.ToolCallID != nil {
		toolCallID = msg.ToolCallID.String()
	}

	toolCalls := make([]*pb.ToolCall, 0, len(msg.ToolCalls))
	for _, tc := range msg.ToolCalls {
		toolCalls = append(toolCalls, DomainToolCallToProto(tc))
	}

	var sender string
	if msg.Sender != nil {
		sender = *msg.Sender
	}

	return &pb.Message{
		Id:         msg.ID.String(),
		ChatId:     msg.ChatID.String(),
		Role:       MessageRoleToProto(msg.Role),
		Content:    msg.Content,
		Sender:     sender,
		ToolCalls:  toolCalls,
		ToolCallId: toolCallID,
		CreatedAt:  timestamppb.New(msg.CreatedAt),
	}
}

// MessageRoleToProto конвертирует domain.MessageRole в proto MessageRole
func MessageRoleToProto(role domain.MessageRole) pb.MessageRole {
	switch role {
	case domain.MessageRoleSystem:
		return pb.MessageRole_MESSAGE_ROLE_SYSTEM
	case domain.MessageRoleUser:
		return pb.MessageRole_MESSAGE_ROLE_USER
	case domain.MessageRoleAssistant:
		return pb.MessageRole_MESSAGE_ROLE_ASSISTANT
	case domain.MessageRoleTool:
		return pb.MessageRole_MESSAGE_ROLE_TOOL
	default:
		return pb.MessageRole_MESSAGE_ROLE_UNSPECIFIED
	}
}

// DomainToolCallToProto конвертирует domain.ToolCall в proto ToolCall
func DomainToolCallToProto(tc *domain.ToolCall) *pb.ToolCall {
	if tc == nil {
		return nil
	}

	var arguments string
	if tc.Arguments != nil {
		argBytes, _ := json.Marshal(tc.Arguments)
		arguments = string(argBytes)
	}

	var result string
	if tc.Result != nil {
		resBytes, _ := json.Marshal(tc.Result)
		result = string(resBytes)
	}

	var completedAt *timestamppb.Timestamp
	if tc.CompletedAt != nil {
		completedAt = timestamppb.New(*tc.CompletedAt)
	}

	return &pb.ToolCall{
		Id:          tc.ID.String(),
		Name:        string(tc.Name),
		Arguments:   arguments,
		Result:      result,
		Status:      ToolCallStatusToProto(tc.Status),
		CreatedAt:   timestamppb.New(tc.CreatedAt),
		CompletedAt: completedAt,
	}
}

// ToolCallStatusToProto конвертирует domain.ToolCallStatus в proto ToolCallStatus
func ToolCallStatusToProto(status domain.ToolCallStatus) pb.ToolCallStatus {
	switch status {
	case domain.ToolCallStatusPending:
		return pb.ToolCallStatus_TOOL_CALL_STATUS_PENDING
	case domain.ToolCallStatusExecuting:
		return pb.ToolCallStatus_TOOL_CALL_STATUS_EXECUTING
	case domain.ToolCallStatusCompleted:
		return pb.ToolCallStatus_TOOL_CALL_STATUS_COMPLETED
	case domain.ToolCallStatusFailed:
		return pb.ToolCallStatus_TOOL_CALL_STATUS_FAILED
	default:
		return pb.ToolCallStatus_TOOL_CALL_STATUS_UNSPECIFIED
	}
}
