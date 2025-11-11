package main

import (
	_ "embed"
	"net/http"
	"net/url"

	"context"
	"fmt"
	"io"
	"log"
	"os"

	"llm-service/internal/app"
	"llm-service/internal/config"
	"llm-service/internal/db"
	"llm-service/internal/domain"
	"llm-service/internal/jwt"
	openai_llm "llm-service/internal/llm/openai"
	"llm-service/internal/logger"
	"llm-service/internal/repository"
	"llm-service/internal/service/chat"
	"llm-service/internal/service/memory"
	"llm-service/internal/service/quota"
	"llm-service/internal/service/tools"
	"llm-service/internal/tracer"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func init() {
	logger.Init()
	godotenv.Load()
	log.SetOutput(io.Discard)
}

func Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	if err := config.Initialize(configPath); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	cfg := config.Get()

	tracer.MustSetup(
		ctx,
		tracer.WithServiceName(cfg.GetServiceName()),
		tracer.WithCollectorEndpoint(cfg.GetJaegerEndpoint()),
	)

	postgresURL := cfg.GetDBDSN()

	pool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		logger.Fatal(ctx, err.Error())
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Fatal(ctx, err.Error())
	}

	openAIClient, err := getOpenAIClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	llmClient := openai_llm.New(openAIClient, cfg)

	transactor := db.NewContextManager(pool)

	repo := repository.NewPGXRepository(transactor)

	memoryService := memory.New(repo)
	quotaService := quota.New(repo, func(ctx context.Context, userID domain.ID) int {
		return cfg.GetLLMTokenLimit()
	})
	toolsService := tools.New(memoryService)
	chatService := chat.New(
		toolsService,
		repo,
		llmClient,
		quotaService,
		memoryService,
	)

	// Initialize JWT provider from config (fallbacks: env JWT_SECRET -> dev-secret)
	jwtSecret := cfg.GetJWTSecret()
	if jwtSecret == "" {
		if env := os.Getenv("JWT_SECRET"); env != "" {
			jwtSecret = env
			logger.Warn(ctx, "jwt.secret not set in config; using JWT_SECRET from environment")
		} else {
			jwtSecret = "dev-secret"
			logger.Warn(ctx, "jwt.secret not set; using insecure default dev secret")
		}
	}
	jwtProvider := jwt.NewProvider(jwt.WithCredentials(jwt.NewSecretCredentials(jwtSecret)))

	app := app.New(
		chatService,
		memoryService,
		jwtProvider,
		app.WithHTTPPathPrefix(cfg.GetHTTPPathPrefix()),
		app.WithGrpcPort(cfg.GetGRPCPort()),
		app.WithGatewayPort(cfg.GetHTTPPort()),
		app.WithEnableGateway(cfg.GetHTTPEnabled()),
	)

	if err := app.Run(ctx); err != nil {
		return err
	}

	return nil
}

type ProxyRoundTripper struct {
	proxy *url.URL
}

func (t *ProxyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if t.proxy != nil {
		transport.Proxy = http.ProxyURL(t.proxy)
	}

	return transport.RoundTrip(req)
}

func getOpenAIClient(cfg *config.Config) (openai.Client, error) {
	proxyURL := cfg.GetProxyUrl()
	apiKey := cfg.GetLLMAPIKey()
	if apiKey == "" {
		return openai.Client{}, fmt.Errorf("LLM API key is not set")
	}

	baseURL := cfg.GetLLMBaseURL()

	options := []option.RequestOption{
		option.WithAPIKey(apiKey),
		option.WithHTTPClient(&http.Client{
			Transport: &ProxyRoundTripper{
				proxy: proxyURL,
			},
		}),
	}

	if baseURL != "" {
		options = append(options, option.WithBaseURL(baseURL))
	}

	return openai.NewClient(options...), nil
}

func main() {
	if err := Run(); err != nil {
		panic(err)
	}
}
