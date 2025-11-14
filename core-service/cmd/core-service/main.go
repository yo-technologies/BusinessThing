package main

import (
	"context"
	"core-service/internal/app"
	"core-service/internal/config"
	"core-service/internal/db"
	"core-service/internal/jwt"
	"core-service/internal/logger"
	"core-service/internal/repository"
	"core-service/internal/service/auth"
	"core-service/internal/service/contract"
	"core-service/internal/service/document"
	"core-service/internal/service/note"
	"core-service/internal/service/organization"
	"core-service/internal/service/template"
	"core-service/internal/service/user"
	"core-service/internal/telegram"
	"core-service/internal/tracer"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var configPath string
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
	logger.Info(ctx, "Starting Core Service")

	// Setup tracer
	tracer.MustSetup(ctx,
		tracer.WithServiceName(cfg.GetServiceName()),
		tracer.WithCollectorEndpoint(cfg.GetJaegerEndpoint()),
	)

	// Initialize database
	pool, err := pgxpool.New(ctx, cfg.GetDBDSN())
	if err != nil {
		logger.Fatalf(ctx, "Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Fatalf(ctx, "Failed to ping database: %v", err)
	}
	logger.Info(ctx, "Database connection established")

	// Initialize context manager (transactor)
	contextManager := db.NewContextManager(pool)

	// Initialize repositories
	repo := repository.NewPGXRepository(contextManager)

	// Initialize services
	orgService := organization.New(repo)
	userService := user.New(repo)
	docService := document.New(repo)
	noteService := note.New(repo)
	templateService := template.New(repo)
	contractService := contract.New(repo)

	// Initialize JWT provider
	jwtSecret := cfg.GetJWTSecret()
	if jwtSecret == "" {
		logger.Warn(ctx, "JWT secret not set in config, using insecure default")
		jwtSecret = "dev-secret-change-in-production"
	}
	jwtProvider := jwt.NewProvider(
		jwt.WithCredentials(jwt.NewSecretCredentials(jwtSecret)),
		jwt.WithAccessTTL(cfg.GetJWTAccessTTL()),
	)

	// Initialize Telegram validator
	telegramValidator := telegram.NewValidator(
		cfg.GetTelegramBotToken(),
		cfg.GetTelegramInitDataTTL(),
	)

	// Initialize auth service
	authService := auth.New(repo, jwtProvider, telegramValidator)

	// Create app with all services
	application := app.New(cfg, jwtProvider, app.Services{
		OrganizationService: orgService,
		UserService:         userService,
		DocumentService:     docService,
		NoteService:         noteService,
		TemplateService:     templateService,
		ContractService:     contractService,
		AuthService:         authService,
	})

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Infof(ctx, "Received signal: %v", sig)
		cancel()
	}()

	// Run app
	if err := application.Run(ctx); err != nil {
		logger.Fatalf(ctx, "Application error: %v", err)
	}

	logger.Info(ctx, "Core Service stopped")
}
