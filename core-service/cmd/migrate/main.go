package main

import (
	"context"
	"core-service/internal/config"
	"core-service/internal/logger"
	"errors"
	"flag"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var direction string
	var configPath string
	flag.StringVar(&direction, "direction", "up", "Migration direction: up or down")
	flag.StringVar(&configPath, "config", "config.yaml", "Path to config file")
	flag.Parse()

	ctx := context.Background()

	// Initialize config
	if err := config.Initialize(configPath); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	cfg := config.Get()

	// Setup logger
	logger.Init()
	logger.Info(ctx, "Starting migration")

	// Create migrator
	m, err := migrate.New(
		"file://migrations",
		cfg.GetDBDSN(),
	)
	if err != nil {
		logger.Fatalf(ctx, "Failed to create migrator: %v", err)
	}
	defer m.Close()

	// Run migration
	switch direction {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			logger.Fatalf(ctx, "Failed to migrate up: %v", err)
		}
		logger.Info(ctx, "Migration up completed successfully")
	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			logger.Fatalf(ctx, "Failed to migrate down: %v", err)
		}
		logger.Info(ctx, "Migration down completed successfully")
	default:
		logger.Fatalf(ctx, "Unknown direction: %s", direction)
	}
}
