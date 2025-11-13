package parser

import (
	"context"
	"io"

	"docs-processor/internal/domain"
)

type Parser interface {
	Parse(ctx context.Context, reader io.Reader) (string, error)
	SupportsType(docType domain.DocumentType) bool
}
