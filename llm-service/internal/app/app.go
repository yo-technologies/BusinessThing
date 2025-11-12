package app

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"llm-service/internal/app/interceptors"
	"llm-service/internal/app/llm-agent/api/agent"
	"llm-service/internal/app/llm-agent/api/memory"
	"llm-service/internal/jwt"
	"llm-service/internal/logger"
	desc "llm-service/pkg/agent"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/swaggest/swgui/v5emb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type Options struct {
	grpcPort         int
	gatewayPort      int
	httpPathPrefix   string
	enableGateway    bool
	enableReflection bool
	swaggerFile      []byte
	bypassCors       bool
	jwtProvider      *jwt.Provider
}

var defaultOptions = &Options{
	grpcPort:         50051,
	gatewayPort:      8080,
	enableGateway:    true,
	enableReflection: true,
	httpPathPrefix:   "",
}

type OptionsFunc func(*Options)

func WithGrpcPort(port int) OptionsFunc {
	return func(o *Options) {
		o.grpcPort = port
	}
}

func WithGatewayPort(port int) OptionsFunc {
	return func(o *Options) {
		o.gatewayPort = port
	}
}

func WithEnableReflection(enableReflection bool) OptionsFunc {
	return func(o *Options) {
		o.enableReflection = enableReflection
	}
}

func WithEnableGateway(enableGateway bool) OptionsFunc {
	return func(o *Options) {
		o.enableGateway = enableGateway
	}
}

func WithSwaggerFile(swaggerFile []byte) OptionsFunc {
	return func(o *Options) {
		o.swaggerFile = swaggerFile
	}
}

func WithHTTPPathPrefix(httpPathPrefix string) OptionsFunc {
	return func(o *Options) {
		o.httpPathPrefix = httpPathPrefix
	}
}

func WithBypassCors(bypassCors bool) OptionsFunc {
	return func(o *Options) {
		o.bypassCors = bypassCors
	}
}

type App struct {
	agentService  *agent.Service
	memoryService *memory.Service
	jwtProvider   *jwt.Provider

	options *Options
}

func New(
	agentService *agent.Service,
	memoryService *memory.Service,
	jwtProvider *jwt.Provider,
	options ...OptionsFunc,
) *App {
	opts := defaultOptions
	for _, o := range options {
		o(opts)
	}
	return &App{
		agentService:  agentService,
		memoryService: memoryService,
		jwtProvider:   jwtProvider,
		options:       opts,
	}
}

func (a *App) Run(ctx context.Context) error {
	grpcEndpoint := fmt.Sprintf(":%d", a.options.grpcPort)
	httpEndpoint := fmt.Sprintf(":%d", a.options.gatewayPort)

	// Define unprotected methods (no auth), e.g., gRPC reflection
	unprotected := []string{
		"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.TracingInterceptor,
			interceptors.RecoveryInterceptor,
			interceptors.ErrCodesInterceptor,
			interceptors.NewUnaryAuthInterceptor(a.jwtProvider, unprotected...),
		),
		grpc.ChainStreamInterceptor(
			interceptors.TracingStreamInterceptor,
			interceptors.RecoveryStreamInterceptor,
			interceptors.ErrCodesStreamInterceptor,
			interceptors.NewStreamAuthInterceptor(a.jwtProvider, unprotected...),
		),
	)

	// Register the services
	desc.RegisterAgentServiceServer(srv, a.agentService)
	desc.RegisterMemoryServiceServer(srv, a.memoryService)

	// Reflect the service
	if a.options.enableReflection {
		reflection.Register(srv)
	}

	// Create gateway
	// Ensure Authorization header is forwarded to gRPC metadata
	headerMatcher := func(key string) (string, bool) {
		if strings.EqualFold(key, "Authorization") {
			return "authorization", true
		}
		return runtime.DefaultHeaderMatcher(key)
	}
	gatewayMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(headerMatcher),
	)

	if err := registerGateway(ctx, gatewayMux, grpcEndpoint); err != nil {
		return err
	}

	var corsHandler *cors.Cors
	if a.options.bypassCors {
		corsHandler = cors.AllowAll()
	} else {
		// TODO: Add allowed origins
		corsHandler = cors.AllowAll()
	}
	gatewayMuxWithCORS := corsHandler.Handler(gatewayMux)

	// Create swagger ui
	httpMux := chi.NewRouter()

	// WebSocket endpoint для будущей реализации
	// httpMux.HandleFunc("/ws/v1/chat", wsHandler.HandleChatStream)

	httpMux.HandleFunc("/swagger", func(w http.ResponseWriter, request *http.Request) {
		logger.Info(request.Context(), "serving swagger file")
		w.Header().Set("Content-Type", "application/json")
		w.Write(desc.GetSwaggerJSON())
	})
	httpMux.Mount("/docs/", v5emb.NewHandler(
		"MultiBanking Core Service",
		fmt.Sprintf("%s/swagger", a.options.httpPathPrefix),
		fmt.Sprintf("%s/docs/", a.options.httpPathPrefix),
	))

	httpMux.Handle("/metrics", promhttp.Handler())

	httpMux.Mount("/", http.StripPrefix(a.options.httpPathPrefix, gatewayMuxWithCORS))

	baseMux := chi.NewRouter()
	prefix := a.options.httpPathPrefix
	if prefix == "" {
		prefix = "/"
	}
	baseMux.Mount(prefix, httpMux)

	httpSrv := &http.Server{
		Addr:    httpEndpoint,
		Handler: baseMux,
	}

	// Start the gateway and swagger ui
	go func() {
		logger.Infof(ctx, "http server listening on port %d", a.options.gatewayPort)
		if err := httpSrv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Errorf(ctx, "error starting http server: %v", err)
			}
		}
	}()

	// Handle shutdown signals
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

		<-stop

		logger.Info(ctx, "shutting down server...")

		err := httpSrv.Shutdown(ctx)
		if err != nil {
			logger.Errorf(ctx, "error shutting down http server: %v", err)
		}

		srv.Stop()
	}()

	// Create listener
	lis, err := net.Listen("tcp", grpcEndpoint)
	if err != nil {
		return err
	}

	logger.Infof(ctx, "grpc server listening on port %d", a.options.grpcPort)

	// Start the server
	if err := srv.Serve(lis); err != nil {
		return err
	}

	logger.Infof(ctx, "grpc server stopped")

	return nil
}

func registerGateway(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	err := desc.RegisterAgentServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return err
	}

	err = desc.RegisterMemoryServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return err
	}

	return nil
}
