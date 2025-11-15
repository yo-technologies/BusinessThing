package service

import (
	"context"
	"fmt"

	"docs-processor/internal/domain"
	"docs-processor/internal/logger"

	"github.com/opentracing/opentracing-go"
)

type JobProcessor struct {
	documentProcessor *DocumentProcessor
	templateProcessor *TemplateProcessor
}

func NewJobProcessor(
	documentProcessor *DocumentProcessor,
	templateProcessor *TemplateProcessor,
) *JobProcessor {
	return &JobProcessor{
		documentProcessor: documentProcessor,
		templateProcessor: templateProcessor,
	}
}

func (p *JobProcessor) ProcessJob(ctx context.Context, job *domain.ProcessingJob) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "service.JobProcessor.ProcessJob")
	defer span.Finish()

	span.SetTag("job_type", string(job.JobType))

	logger.Info(ctx, "Processing job",
		"job_type", string(job.JobType),
		"retry_count", job.RetryCount,
	)

	var err error
	switch job.JobType {
	case domain.JobTypeDocument:
		err = p.documentProcessor.ProcessDocument(ctx, job)
	case domain.JobTypeTemplateIndex:
		err = p.templateProcessor.IndexTemplate(ctx, job)
	case domain.JobTypeTemplateDelete:
		err = p.templateProcessor.DeleteTemplate(ctx, job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.JobType)
	}

	if err != nil {
		logger.Error(ctx, "Job processing failed",
			"job_type", string(job.JobType),
			"error", err,
			"retry_count", job.RetryCount,
		)
		return err
	}

	logger.Info(ctx, "Job processed successfully",
		"job_type", string(job.JobType),
	)

	return nil
}
