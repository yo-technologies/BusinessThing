package mappers

import (
	"llm-service/internal/domain"
	pb "llm-service/pkg/agent"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// DomainChatToProto конвертирует domain.Chat в proto Chat
func DomainChatToProto(chat *domain.Chat) *pb.Chat {
	if chat == nil {
		return nil
	}

	var parentChatID string
	if chat.ParentChatID != nil {
		parentChatID = chat.ParentChatID.String()
	}

	var parentToolCallID string
	if chat.ParentToolCallID != nil {
		parentToolCallID = chat.ParentToolCallID.String()
	}

	return &pb.Chat{
		Id:               chat.ID.String(),
		OrganizationId:   chat.OrganizationID.String(),
		UserId:           chat.UserID.String(),
		AgentKey:         chat.AgentKey,
		Title:            chat.Title,
		Status:           ChatStatusToProto(chat.Status),
		ParentChatId:     parentChatID,
		ParentToolCallId: parentToolCallID,
		CreatedAt:        timestamppb.New(chat.CreatedAt),
		UpdatedAt:        timestamppb.New(chat.UpdatedAt),
	}
}

// ChatStatusToProto конвертирует domain.ChatStatus в proto ChatStatus
func ChatStatusToProto(status domain.ChatStatus) pb.ChatStatus {
	switch status {
	case domain.ChatStatusActive:
		return pb.ChatStatus_CHAT_STATUS_ACTIVE
	case domain.ChatStatusCompleted:
		return pb.ChatStatus_CHAT_STATUS_COMPLETED
	case domain.ChatStatusFailed:
		return pb.ChatStatus_CHAT_STATUS_FAILED
	case domain.ChatStatusArchived:
		return pb.ChatStatus_CHAT_STATUS_ARCHIVED
	default:
		return pb.ChatStatus_CHAT_STATUS_UNSPECIFIED
	}
}
