package main

import (
	"context"
	"io"
	"log"
	"os"

	"docs-processor/internal/app"
	"docs-processor/internal/config"
	"docs-processor/internal/embeddings"
	"docs-processor/internal/logger"
	"docs-processor/internal/service"
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
	ctx := context.Background()

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
		tracer.WithServiceName(cfg.GetServiceName()),
		tracer.WithCollectorEndpoint(cfg.GetJaegerEndpoint()),
	)
	defer closer.Close()

	vectorDB, err := vectordb.NewOpenSearchClient(
		cfg.GetOpenSearchAddresses(),
		cfg.GetOpenSearchUsername(),
		cfg.GetOpenSearchPassword(),
		cfg.GetOpenSearchIndexName(),
	)
	if err != nil {
		logger.Fatal(ctx, "Failed to create OpenSearch client", "error", err)
	}

	embeddingsCli := embeddings.NewClient(
		cfg.GetEmbeddingsBaseURL(),
		cfg.GetEmbeddingsAPIKey(),
		cfg.GetEmbeddingsModel(),
	)

	searchService := service.NewSearchService(embeddingsCli, vectorDB)

	application := app.New(
		searchService,
		app.WithGrpcPort(cfg.GetGRPCPort()),
		app.WithGatewayPort(cfg.GetHTTPPort()),
		app.WithEnableGateway(cfg.GetEnableGateway()),
		app.WithHTTPPathPrefix(cfg.GetHTTPPathPrefix()),
	)

	logger.Info(ctx, "Starting doc-processor service")

	if err := application.Run(ctx); err != nil {
		logger.Fatal(ctx, "Application error", "error", err)
	}
}
