package parser

import (
	"context"
	"io"

	"doc-processing/internal/domain"
)

type Parser interface {
	Parse(ctx context.Context, reader io.Reader) (string, error)
	SupportsType(docType domain.DocumentType) bool
}
