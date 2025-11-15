package docx

import (
	"bytes"
	"context"
	"fmt"

	"llm-service/internal/logger"

	"github.com/lukasjarosch/go-docx"
	"github.com/opentracing/opentracing-go"
)

type Processor struct{}

func New() *Processor {
	return &Processor{}
}

// FillTemplate заполняет DOCX шаблон значениями из values
// Использует библиотеку go-docx для корректной обработки фрагментированных плейсхолдеров
// Поддерживает плейсхолдеры вида {field_name}
func (p *Processor) FillTemplate(ctx context.Context, templateData []byte, values map[string]interface{}) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "docx.Processor.FillTemplate")
	defer span.Finish()

	logger.Info(ctx, "Starting DOCX template filling",
		"template_size_bytes", len(templateData),
		"values_count", len(values),
	)

	// Открываем DOCX из байтов
	doc, err := docx.OpenBytes(templateData)
	if err != nil {
		logger.Error(ctx, "Failed to open DOCX template", "error", err)
		return nil, fmt.Errorf("failed to open DOCX template: %w", err)
	}

	logger.Debug(ctx, "DOCX template opened successfully")

	// Конвертируем values в PlaceholderMap
	// Библиотека использует плейсхолдеры без фигурных скобок
	replaceMap := make(docx.PlaceholderMap)
	for key, value := range values {
		replaceMap[key] = fmt.Sprintf("%v", value)
		logger.Debug(ctx, "Added placeholder mapping",
			"placeholder", key,
			"value", fmt.Sprintf("%v", value),
		)
	}

	logger.Info(ctx, "Prepared placeholder replacement map", "placeholders_count", len(replaceMap))

	// Заменяем все плейсхолдеры
	err = doc.ReplaceAll(replaceMap)
	if err != nil {
		logger.Error(ctx, "Failed to replace placeholders", "error", err)
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}

	logger.Info(ctx, "Placeholders replaced successfully")

	// Записываем результат в буфер
	var buf bytes.Buffer
	err = doc.Write(&buf)
	if err != nil {
		logger.Error(ctx, "Failed to write filled DOCX", "error", err)
		return nil, fmt.Errorf("failed to write filled DOCX: %w", err)
	}

	resultSize := buf.Len()
	logger.Info(ctx, "DOCX template filled successfully",
		"result_size_bytes", resultSize,
		"size_change", resultSize-len(templateData),
	)

	return buf.Bytes(), nil
}

// ExtractPlaceholders извлекает все плейсхолдеры из DOCX шаблона
func (p *Processor) ExtractPlaceholders(ctx context.Context, templateData []byte) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "docx.Processor.ExtractPlaceholders")
	defer span.Finish()

	logger.Info(ctx, "Extracting placeholders from DOCX template",
		"template_size_bytes", len(templateData),
	)

	// Открываем DOCX из байтов
	doc, err := docx.OpenBytes(templateData)
	if err != nil {
		logger.Error(ctx, "Failed to open DOCX template for placeholder extraction", "error", err)
		return nil, fmt.Errorf("failed to open DOCX template: %w", err)
	}

	logger.Debug(ctx, "DOCX template opened successfully for placeholder extraction")

	// Получаем все плейсхолдеры из документа
	// Библиотека парсит документ и находит все {placeholder} включая фрагментированные
	placeholders, err := doc.GetPlaceHoldersList()
	if err != nil {
		logger.Error(ctx, "Failed to extract placeholders", "error", err)
		return nil, fmt.Errorf("failed to extract placeholders: %w", err)
	}

	logger.Info(ctx, "Placeholders extracted successfully", "count", len(placeholders))

	for i, placeholder := range placeholders {
		logger.Debug(ctx, "Found placeholder",
			"index", i,
			"name", placeholder,
		)
	}

	return placeholders, nil
}
