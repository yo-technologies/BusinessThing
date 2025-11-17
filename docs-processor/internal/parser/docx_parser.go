package parser

import (
	"context"
	"fmt"
	"io"

	"docs-processor/internal/domain"

	"code.sajari.com/docconv/v2"
	"github.com/opentracing/opentracing-go"
)

type DOCXParser struct{}

func NewDOCXParser() *DOCXParser {
	return &DOCXParser{}
}

func (p *DOCXParser) Parse(ctx context.Context, reader io.Reader) (string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "parser.DOCXParser.Parse")
	defer span.Finish()

	res, _, err := docconv.ConvertDocx(reader)
	if err != nil {
		return "", fmt.Errorf("failed to convert DOCX: %w", err)
	}

	return res, nil
}

func (p *DOCXParser) SupportsType(docType domain.DocumentType) bool {
	return docType == domain.DocumentTypeDOCX || 
		docType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
}
