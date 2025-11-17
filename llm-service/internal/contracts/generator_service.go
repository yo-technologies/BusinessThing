package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"llm-service/internal/coreservice"
	"llm-service/internal/docx"
	"llm-service/internal/domain"
	"llm-service/internal/storage"

	"github.com/opentracing/opentracing-go"
)

type GeneratorService struct {
	coreServiceClient coreservice.Client
	s3Client          *storage.S3Client
	docxProcessor     *docx.Processor
}

func NewGeneratorService(
	coreServiceClient coreservice.Client,
	s3Client *storage.S3Client,
	docxProcessor *docx.Processor,
) *GeneratorService {
	return &GeneratorService{
		coreServiceClient: coreServiceClient,
		s3Client:          s3Client,
		docxProcessor:     docxProcessor,
	}
}

// GenerateContract генерирует договор из шаблона с заполненными данными
func (s *GeneratorService) GenerateContract(
	ctx context.Context,
	organizationID, templateID, contractName string,
	filledData map[string]interface{},
) (*GeneratedContract, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "contracts.GeneratorService.GenerateContract")
	defer span.Finish()

	// 1. Получаем шаблон из core-service
	template, err := s.coreServiceClient.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// 2. Валидируем filled_data по fields_schema
	if err := s.validateFilledData(template.FieldsSchema, filledData); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 3. Загружаем DOCX шаблон из S3
	if template.S3TemplateKey == "" {
		return nil, fmt.Errorf("template does not have S3 template key")
	}

	templateReader, err := s.s3Client.GetObject(ctx, template.S3TemplateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get template from S3(%s): %w", template.S3TemplateKey, err)
	}
	defer templateReader.Close()

	templateData, err := io.ReadAll(templateReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read template data: %w", err)
	}

	// 4. Заполняем DOCX
	filledDocx, err := s.docxProcessor.FillTemplate(ctx, templateData, filledData)
	if err != nil {
		return nil, fmt.Errorf("failed to fill template: %w", err)
	}

	// 5. Сохраняем в S3
	contractID := domain.NewID()
	s3Key := fmt.Sprintf("contracts/%s/%s.docx", organizationID, contractID)

	if err := s.s3Client.PutObject(ctx, s3Key, filledDocx, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"); err != nil {
		return nil, fmt.Errorf("failed to save contract to S3: %w", err)
	}

	// 6. Регистрируем через core-service
	filledDataJSON, _ := json.Marshal(filledData)
	contract, err := s.coreServiceClient.RegisterContract(ctx, organizationID, templateID, contractName, string(filledDataJSON), s3Key, "docx")
	if err != nil {
		return nil, fmt.Errorf("failed to register contract: %w", err)
	}

	return &GeneratedContract{
		ContractID:   contract.ID,
		Name:         contract.Name,
		DownloadURL:  fmt.Sprintf("/api/v1/contracts/%s/download", contract.ID),
		S3Key:        s3Key,
		CreatedAt:    time.Now(),
		TemplateName: template.Name,
	}, nil
}

func (s *GeneratorService) validateFilledData(fieldsSchemaJSON string, filledData map[string]interface{}) error {
	var schema struct {
		Fields []TemplateField `json:"fields"`
	}

	if err := json.Unmarshal([]byte(fieldsSchemaJSON), &schema); err != nil {
		return fmt.Errorf("failed to parse fields schema: %w", err)
	}

	// Проверяем обязательные поля
	for _, field := range schema.Fields {
		if field.Required {
			if _, ok := filledData[field.Name]; !ok {
				return fmt.Errorf("required field '%s' is missing", field.Name)
			}
		}
	}

	return nil
}

// ListContracts возвращает список сгенерированных договоров
func (s *GeneratorService) ListContracts(ctx context.Context, organizationID string, limit, offset int) ([]*ContractListItem, int, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "contracts.GeneratorService.ListContracts")
	defer span.Finish()

	contracts, total, err := s.coreServiceClient.ListContracts(ctx, organizationID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list contracts: %w", err)
	}

	results := make([]*ContractListItem, 0, len(contracts))
	for _, c := range contracts {
		// Получаем название шаблона
		template, err := s.coreServiceClient.GetTemplate(ctx, c.TemplateID)
		templateName := ""
		if err == nil {
			templateName = template.Name
		}

		results = append(results, &ContractListItem{
			ID:           c.ID,
			Name:         c.Name,
			TemplateName: templateName,
			TemplateID:   c.TemplateID,
			CreatedAt:    c.CreatedAt,
			DownloadURL:  fmt.Sprintf("/api/v1/contracts/%s/download", c.ID),
		})
	}

	return results, total, nil
}
