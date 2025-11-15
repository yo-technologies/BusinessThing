package docx

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go"
)

type Processor struct{}

func New() *Processor {
	return &Processor{}
}

// FillTemplate заполняет DOCX шаблон значениями из values
// Поддерживает плейсхолдеры вида {{field_name}}
func (p *Processor) FillTemplate(ctx context.Context, templateData []byte, values map[string]interface{}) ([]byte, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "docx.Processor.FillTemplate")
	defer span.Finish()

	// Открываем DOCX как ZIP архив
	zipReader, err := zip.NewReader(bytes.NewReader(templateData), int64(len(templateData)))
	if err != nil {
		return nil, fmt.Errorf("failed to open DOCX as ZIP: %w", err)
	}

	// Создаем новый буфер для результата
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	// Регулярное выражение для поиска плейсхолдеров
	placeholderRegex := regexp.MustCompile(`\{\{([a-zA-Z0-9_]+)\}\}`)

	// Обрабатываем каждый файл в архиве
	for _, file := range zipReader.File {
		// Открываем файл из архива
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", file.Name, err)
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file.Name, err)
		}

		// Если это XML файл с содержимым, заменяем плейсхолдеры
		if strings.HasSuffix(file.Name, ".xml") {
			contentStr := string(content)
			contentStr = placeholderRegex.ReplaceAllStringFunc(contentStr, func(match string) string {
				// Извлекаем имя поля из {{field_name}}
				fieldName := placeholderRegex.FindStringSubmatch(match)[1]
				if value, ok := values[fieldName]; ok {
					return fmt.Sprintf("%v", value)
				}
				// Если значение не найдено, оставляем плейсхолдер
				return match
			})
			content = []byte(contentStr)
		}

		// Создаем файл в новом архиве
		w, err := zipWriter.CreateHeader(&file.FileHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to create file %s in result: %w", file.Name, err)
		}

		if _, err := w.Write(content); err != nil {
			return nil, fmt.Errorf("failed to write file %s to result: %w", file.Name, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close ZIP writer: %w", err)
	}

	return buf.Bytes(), nil
}

// ExtractPlaceholders извлекает все плейсхолдеры из DOCX шаблона
func (p *Processor) ExtractPlaceholders(ctx context.Context, templateData []byte) ([]string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "docx.Processor.ExtractPlaceholders")
	defer span.Finish()

	zipReader, err := zip.NewReader(bytes.NewReader(templateData), int64(len(templateData)))
	if err != nil {
		return nil, fmt.Errorf("failed to open DOCX as ZIP: %w", err)
	}

	placeholderRegex := regexp.MustCompile(`\{\{([a-zA-Z0-9_]+)\}\}`)
	placeholdersMap := make(map[string]struct{})

	// Ищем плейсхолдеры во всех XML файлах
	for _, file := range zipReader.File {
		if !strings.HasSuffix(file.Name, ".xml") {
			continue
		}

		rc, err := file.Open()
		if err != nil {
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}

		matches := placeholderRegex.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) > 1 {
				placeholdersMap[match[1]] = struct{}{}
			}
		}
	}

	// Преобразуем map в slice
	placeholders := make([]string, 0, len(placeholdersMap))
	for placeholder := range placeholdersMap {
		placeholders = append(placeholders, placeholder)
	}

	return placeholders, nil
}
