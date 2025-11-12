package mappers

import (
	"llm-service/internal/domain"
	pb "llm-service/pkg/agent"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// DomainMemoryFactToProto конвертирует domain.UserMemoryFact в proto MemoryFact
func DomainMemoryFactToProto(fact *domain.UserMemoryFact) *pb.MemoryFact {
	if fact == nil {
		return nil
	}

	return &pb.MemoryFact{
		Id:        fact.ID.String(),
		Content:   fact.Content,
		CreatedAt: timestamppb.New(fact.CreatedAt),
	}
}
