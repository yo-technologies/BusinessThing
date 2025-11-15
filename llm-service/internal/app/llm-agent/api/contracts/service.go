package contracts

import (
	"context"
	"encoding/json"
	"fmt"

	"llm-service/internal/domain"
	"llm-service/internal/logger"
	"llm-service/internal/service"
	pb "llm-service/pkg/agent"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	contractGeneratorService service.ContractGeneratorService
	pb.UnimplementedContractsServiceServer
}

func NewService(contractGeneratorService service.ContractGeneratorService) *Service {
	return &Service{
		contractGeneratorService: contractGeneratorService,
	}
}

// TestGenerateContract тестовая ручка для генерации контракта из шаблона
func (s *Service) TestGenerateContract(ctx context.Context, req *pb.TestGenerateContractRequest) (*pb.TestGenerateContractResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "api.TestGenerateContract")
	defer span.Finish()

	logger.Info(ctx, "Test generate contract request received",
		"org_id", req.OrgId,
		"template_id", req.TemplateId,
		"contract_name", req.ContractName,
		"fields_count", len(req.FilledData),
	)

	// Конвертируем map<string, string> в map[string]interface{}
	filledData := make(map[string]interface{}, len(req.FilledData))
	for key, value := range req.FilledData {
		filledData[key] = value
		logger.Debug(ctx, "Field mapping",
			"field", key,
			"value", value,
		)
	}

	// Вызываем сервис генерации
	result, err := s.contractGeneratorService.GenerateContract(
		ctx,
		req.OrgId,
		req.TemplateId,
		req.ContractName,
		filledData,
	)
	if err != nil {
		logger.Error(ctx, "Failed to generate contract", "error", err)
		return nil, err
	}

	// Сериализуем результат в JSON и обратно в map для универсальности
	resultJSON, err := json.Marshal(result)
	if err != nil {
		logger.Error(ctx, "Failed to marshal result", "error", err)
		return nil, domain.NewInternalError("failed to marshal result", err)
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(resultJSON, &resultMap); err != nil {
		logger.Error(ctx, "Failed to unmarshal result", "error", err)
		return nil, domain.NewInternalError("failed to unmarshal result", err)
	}

	logger.Info(ctx, "Contract generated successfully",
		"contract_id", getStringFromMap(resultMap, "contract_id"),
	)

	response := &pb.TestGenerateContractResponse{
		ContractId:   getStringFromMap(resultMap, "contract_id"),
		ContractName: getStringFromMap(resultMap, "name"),
		DownloadUrl:  getStringFromMap(resultMap, "download_url"),
		S3Key:        getStringFromMap(resultMap, "s3_key"),
		TemplateName: getStringFromMap(resultMap, "template_name"),
		CreatedAt:    timestamppb.Now(),
	}

	return response, nil
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}
