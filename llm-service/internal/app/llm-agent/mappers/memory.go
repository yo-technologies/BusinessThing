package mappers

import (
	"llm-service/internal/domain"
	pb "llm-service/pkg/agent"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// DomainOrganizationMemoryFactToProto конвертирует domain.OrganizationMemoryFact в proto MemoryFact
func DomainOrganizationMemoryFactToProto(fact *domain.OrganizationMemoryFact) *pb.MemoryFact {
	if fact == nil {
		return nil
	}

	return &pb.MemoryFact{
		Id:        fact.ID.String(),
		Content:   fact.Content,
		CreatedAt: timestamppb.New(fact.CreatedAt),
	}
}
