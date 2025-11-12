package main

import (
	"context"
	"database/sql"
	"os"

	"llm-service/internal/config"
	"llm-service/internal/logger"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func init() {
	logger.Init()
	godotenv.Load()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config.yaml"
	}

	if err := config.Initialize(cfgPath); err != nil {
		logger.Fatalf(ctx, "failed to initialize config: %v", err)
	}

	cfg := config.Get()

	postgresURL := cfg.GetDBDSN()

	sqlDB, err := sql.Open("postgres", postgresURL)
	if err != nil {
		logger.Fatalf(ctx, "failed to connect to database: %v", err)
	}

	if err := goose.RunContext(ctx, "up", sqlDB, "migrations"); err != nil {
		logger.Fatalf(ctx, "failed to apply migrations: %v", err)
	}

	logger.Info(ctx, "migrations applied")
}
