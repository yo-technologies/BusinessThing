package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"docs-processor/internal/chunker"
	"docs-processor/internal/config"
	"docs-processor/internal/coreservice"
	"docs-processor/internal/embeddings"
	"docs-processor/internal/logger"
	"docs-processor/internal/parser"
	"docs-processor/internal/queue"
	"docs-processor/internal/service"
	"docs-processor/internal/storage"
	"docs-processor/internal/tracer"
	"docs-processor/internal/vectordb"

	"github.com/joho/godotenv"
)

func init() {
	logger.Init()
	godotenv.Load()
	log.SetOutput(io.Discard)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	if err := config.Initialize(configPath); err != nil {
		logger.Fatal(ctx, "Failed to initialize config", "error", err)
	}

	cfg := config.Get()

	closer := tracer.MustSetup(
		ctx,
		tracer.WithServiceName(cfg.GetServiceName()+"-worker"),
		tracer.WithCollectorEndpoint(cfg.GetJaegerEndpoint()),
	)
	defer closer.Close()

	s3Client, err := storage.NewS3Client(
		cfg.GetS3Endpoint(),
		cfg.GetS3AccessKey(),
		cfg.GetS3SecretKey(),
		cfg.GetS3Region(),
		cfg.GetS3Bucket(),
		cfg.GetS3UseSSL(),
	)
	if err != nil {
		logger.Fatal(ctx, "Failed to create S3 client", "error", err)
	}

	vectorDB, err := vectordb.NewOpenSearchClient(
		cfg.GetOpenSearchAddresses(),
		cfg.GetOpenSearchUsername(),
		cfg.GetOpenSearchPassword(),
		cfg.GetOpenSearchIndexName(),
	)
	if err != nil {
		logger.Fatal(ctx, "Failed to create OpenSearch client", "error", err)
	}

	templatesDB, err := vectordb.NewTemplatesClient(
		cfg.GetOpenSearchAddresses(),
		cfg.GetOpenSearchUsername(),
		cfg.GetOpenSearchPassword(),
		cfg.GetOpenSearchTemplatesIndex(),
	)
	if err != nil {
		logger.Fatal(ctx, "Failed to create templates OpenSearch client", "error", err)
	}

	embeddingsCli := embeddings.NewClient(
		cfg.GetEmbeddingsBaseURL(),
		cfg.GetEmbeddingsAPIKey(),
		cfg.GetEmbeddingsModel(),
	)

	parserRegistry := parser.NewRegistry()
	textChunker := chunker.New(
		cfg.GetChunkingMaxChunkSize(),
		cfg.GetChunkingOverlapSize(),
	)

	coreClient, err := coreservice.NewClient(cfg.GetCoreServiceAddress())
	if err != nil {
		logger.Fatal(ctx, "Failed to create core-service client", "error", err)
	}
	defer coreClient.Close()

	documentProcessor := service.NewDocumentProcessor(
		s3Client,
		parserRegistry,
		textChunker,
		embeddingsCli,
		vectorDB,
		coreClient,
		cfg.GetEmbeddingsBatchSize(),
	)

	templateProcessor := service.NewTemplateProcessor(
		embeddingsCli,
		templatesDB,
	)

	jobProcessor := service.NewJobProcessor(
		documentProcessor,
		templateProcessor,
		coreClient,
	)

	rabbitMQ, err := queue.NewRabbitMQClient(
		cfg.GetRabbitMQURL(),
		cfg.GetRabbitMQQueueName(),
	)
	if err != nil {
		logger.Fatal(ctx, "Failed to create RabbitMQ client", "error", err)
	}
	defer rabbitMQ.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info(ctx, "Received shutdown signal")
		cancel()
	}()

	logger.Info(ctx, "Starting document processing worker")

	if err := rabbitMQ.ConsumeJobs(
		ctx,
		cfg.GetRabbitMQConsumerTag(),
		cfg.GetRabbitMQPrefetchCount(),
		jobProcessor.ProcessJob,
	); err != nil {
		logger.Fatal(ctx, "Worker error", "error", err)
	}

	logger.Info(ctx, "Worker stopped")
}
