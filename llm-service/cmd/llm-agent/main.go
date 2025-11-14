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
	agentapi "llm-service/internal/app/llm-agent/api/agent"
	memoryapi "llm-service/internal/app/llm-agent/api/memory"
	"llm-service/internal/config"
	"llm-service/internal/db"
	"llm-service/internal/domain"
	"llm-service/internal/jwt"
	openai_llm "llm-service/internal/llm/openai"
	"llm-service/internal/logger"
	"llm-service/internal/rag"
	"llm-service/internal/repository"
	"llm-service/internal/service/agent"
	"llm-service/internal/service/chat"
	contextbuilder "llm-service/internal/service/context"
	"llm-service/internal/service/executor"
	"llm-service/internal/service/memory"
	"llm-service/internal/service/quota"
	"llm-service/internal/service/subagent"
	"llm-service/internal/service/tool"
	"llm-service/internal/tracer"
	"llm-service/internal/websearch"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// Initialize services
	memoryService := memory.New(repo)
	quotaService := quota.New(repo, func(ctx context.Context, userID domain.ID) int {
		return cfg.GetLLMTokenLimit()
	})

	// Initialize agent manager
	agentManager, err := agent.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create agent manager: %w", err)
	}

	// Initialize chat manager
	var chatRepo repository.ChatRepository = repo
	var messageRepo repository.MessageRepository = repo
	var toolRepo repository.ToolCallRepository = repo

	chatManager := chat.NewManager(chatRepo, messageRepo, toolRepo)

	// Initialize RAG client
	ragClient, err := rag.NewClient(cfg.GetDocsProcessorAddress())
	if err != nil {
		return fmt.Errorf("failed to create RAG client: %w", err)
	}
	defer ragClient.Close()

	// Initialize context builder
	ctxBuilder := contextbuilder.NewBuilder(ragClient)

	// Initialize subagent manager
	subagentManager := subagent.NewManager(chatManager, agentManager)

	// Initialize Tavily web search client
	tavilyAPIKey := cfg.GetTavilyAPIKey()
	if tavilyAPIKey == "" {
		logger.Warn(ctx, "Tavily API key not set, web search will not be available")
	}
	tavilyClient := websearch.NewTavilyClient(tavilyAPIKey)

	// Initialize tool executor
	toolExecutor := tool.NewExecutor(agentManager, subagentManager, tavilyClient)

	// Initialize agent executor
	agentExecutor := executor.NewExecutor(
		chatManager,
		agentManager,
		ctxBuilder,
		toolExecutor,
		subagentManager,
		llmClient,
		cfg,
	)

	// Create API services
	agentAPIService := agentapi.NewService(chatManager, agentExecutor, quotaService)
	memoryAPIService := memoryapi.NewService(memoryService)

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

	// Create gRPC client connection for WebSocket proxy to localhost gRPC server
	grpcConn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", cfg.GetGRPCPort()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC client connection: %w", err)
	}
	defer grpcConn.Close()

	app := app.New(
		agentAPIService,
		memoryAPIService,
		jwtProvider,
		app.WithHTTPPathPrefix(cfg.GetHTTPPathPrefix()),
		app.WithGrpcPort(cfg.GetGRPCPort()),
		app.WithGatewayPort(cfg.GetHTTPPort()),
		app.WithEnableGateway(cfg.GetHTTPEnabled()),
		app.WithWSGrpcConn(grpcConn),
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
