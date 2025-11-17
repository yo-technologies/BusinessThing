package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"docs-processor/internal/domain"

	"github.com/ledongthuc/pdf"
	"github.com/opentracing/opentracing-go"
)

type PDFParser struct{}

func NewPDFParser() *PDFParser {
	return &PDFParser{}
}

func (p *PDFParser) Parse(ctx context.Context, reader io.Reader) (string, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "parser.PDFParser.Parse")
	defer span.Finish()

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF: %w", err)
	}

	pdfReader, err := pdf.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("failed to parse PDF: %w", err)
	}

	var text strings.Builder
	numPages := pdfReader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}

		pageText, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}

		text.WriteString(pageText)
		text.WriteString("\n")
	}

	return text.String(), nil
}

func (p *PDFParser) SupportsType(docType domain.DocumentType) bool {
	return docType == domain.DocumentTypePDF || docType == "application/pdf"
}
