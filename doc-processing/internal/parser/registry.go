package parser

import (
	"context"
	"fmt"
	"io"

	"doc-processing/internal/domain"
)

type Registry struct {
	parsers []Parser
}

func NewRegistry() *Registry {
	return &Registry{
		parsers: []Parser{
			NewPDFParser(),
			NewDOCXParser(),
			NewTXTParser(),
		},
	}
}

func (r *Registry) GetParser(docType domain.DocumentType) (Parser, error) {
	for _, parser := range r.parsers {
		if parser.SupportsType(docType) {
			return parser, nil
		}
	}

	return nil, fmt.Errorf("no parser found for document type: %s", docType)
}

func (r *Registry) Parse(ctx context.Context, docType domain.DocumentType, reader io.Reader) (string, error) {
	parser, err := r.GetParser(docType)
	if err != nil {
		return "", fmt.Errorf("parser not found: %w", err)
	}

	return parser.Parse(ctx, reader)
}
