package parser

import (
	"context"
	"fmt"
	"io"

	"docs-processor/internal/domain"

	"github.com/opentracing/opentracing-go"
)

type TXTParser struct{}

func NewTXTParser() *TXTParser {
	return &TXTParser{}
}

func (p *TXTParser) Parse(ctx context.Context, reader io.Reader) (string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "parser.TXTParser.Parse")
	defer span.Finish()

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read TXT: %w", err)
	}

	return string(content), nil
}

func (p *TXTParser) SupportsType(docType domain.DocumentType) bool {
	return docType == domain.DocumentTypeTXT
}
