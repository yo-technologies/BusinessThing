package service

import (
	"context"

	pb "docs-processor/pkg/document"

	"github.com/opentracing/opentracing-go"
)

type TemplateSearchService struct {
	pb.UnimplementedDocumentServiceServer
	templateProcessor *TemplateProcessor
}

func NewTemplateSearchService(templateProcessor *TemplateProcessor) *TemplateSearchService {
	return &TemplateSearchService{
		templateProcessor: templateProcessor,
	}
}

func (s *TemplateSearchService) SearchTemplates(ctx context.Context, req *pb.SearchTemplatesRequest) (*pb.SearchTemplatesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.SearchTemplates")
	defer span.Finish()

	limit := int(req.Limit)
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	results, err := s.templateProcessor.SearchTemplates(ctx, req.Query, limit)
	if err != nil {
		return nil, err
	}

	pbResults := make([]*pb.TemplateSearchResult, 0, len(results))
	for _, r := range results {
		pbResults = append(pbResults, &pb.TemplateSearchResult{
			TemplateId:   r.TemplateID.String(),
			Name:         r.Name,
			Description:  r.Description,
			TemplateType: r.TemplateType,
			FieldsCount:  int32(r.FieldsCount),
			Score:        r.Score,
		})
	}

	return &pb.SearchTemplatesResponse{
		Templates: pbResults,
	}, nil
}
