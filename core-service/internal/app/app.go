package app

import (
	"context"
	"core-service/internal/app/core-service/api/auth"
	"core-service/internal/app/core-service/api/contract"
	"core-service/internal/app/core-service/api/document"
	"core-service/internal/app/core-service/api/note"
	"core-service/internal/app/core-service/api/organization"
	"core-service/internal/app/core-service/api/storage"
	"core-service/internal/app/core-service/api/template"
	"core-service/internal/app/core-service/api/user"
	"core-service/internal/app/interceptors"
	"core-service/internal/config"
	"core-service/internal/jwt"
	"core-service/internal/logger"
	pb "core-service/pkg/core"
	"fmt"
	"net"
	"net/http"

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
	httpPathPrefix string
}

var defaultOptions = &Options{
	httpPathPrefix: "",
}

type OptionsFunc func(*Options)

func WithHTTPPathPrefix(httpPathPrefix string) OptionsFunc {
	return func(o *Options) {
		o.httpPathPrefix = httpPathPrefix
	}
}

type App struct {
	cfg             *config.Config
	jwtProvider     *jwt.Provider
	orgService      *organization.Service
	userService     *user.Service
	docService      *document.Service
	noteService     *note.Service
	templateService *template.Service
	contractService *contract.Service
	storageService  *storage.Service
	authService     *auth.Service
	grpcServer      *grpc.Server
	httpServer      *http.Server
	options         *Options
}

type Services struct {
	OrganizationService organization.OrganizationService
	UserService         user.UserService
	DocumentService     document.DocumentService
	NoteService         note.NoteService
	TemplateService     template.TemplateService
	ContractService     contract.ContractService
	StorageService      storage.StorageService
}

func New(cfg *config.Config, jwtProvider *jwt.Provider, services Services, authSvc *auth.Service, opts ...OptionsFunc) *App {
	options := defaultOptions
	for _, opt := range opts {
		opt(options)
	}

	app := &App{
		cfg:             cfg,
		jwtProvider:     jwtProvider,
		orgService:      organization.NewService(services.OrganizationService),
		userService:     user.NewService(services.UserService),
		docService:      document.NewService(services.DocumentService),
		noteService:     note.NewService(services.NoteService),
		templateService: template.NewService(services.TemplateService),
		contractService: contract.NewService(services.ContractService),
		storageService:  storage.NewService(services.StorageService),
		authService:     authSvc,
		options:         options,
	}

	app.setupGRPC()
	return app
}

func (a *App) setupGRPC() {
	// Unprotected methods (public endpoints)
	unprotected := []string{
		"/core.api.core.AuthService/AuthenticateWithTelegram",
		"/core.api.core.AuthService/CompleteRegistration",
	}

	// Setup interceptors
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		interceptors.RecoveryInterceptor,
		interceptors.TracingInterceptor,
		interceptors.ErrCodesInterceptor,
		interceptors.NewUnaryAuthInterceptor(a.jwtProvider, unprotected...),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		interceptors.RecoveryStreamInterceptor,
		interceptors.TracingStreamInterceptor,
		interceptors.ErrCodesStreamInterceptor,
		interceptors.NewStreamAuthInterceptor(a.jwtProvider, unprotected...),
	}

	// Create gRPC server
	a.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	// Register services
	pb.RegisterAuthServiceServer(a.grpcServer, a.authService)
	pb.RegisterOrganizationServiceServer(a.grpcServer, a.orgService)
	pb.RegisterUserServiceServer(a.grpcServer, a.userService)
	pb.RegisterDocumentServiceServer(a.grpcServer, a.docService)
	pb.RegisterNoteServiceServer(a.grpcServer, a.noteService)
	pb.RegisterContractTemplateServiceServer(a.grpcServer, a.templateService)
	pb.RegisterGeneratedContractServiceServer(a.grpcServer, a.contractService)
	pb.RegisterStorageServiceServer(a.grpcServer, a.storageService)

	// Register reflection for grpcurl
	reflection.Register(a.grpcServer)
}

func (a *App) Run(ctx context.Context) error {
	// Start gRPC server
	grpcAddr := fmt.Sprintf(":%d", a.cfg.GRPC.Port)
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	logger.Infof(ctx, "Starting gRPC server on %s", grpcAddr)

	go func() {
		if err := a.grpcServer.Serve(listener); err != nil {
			logger.Errorf(ctx, "gRPC server error: %v", err)
		}
	}()

	// Start HTTP gateway
	if a.cfg.GetHTTPEnabled() {
		if err := a.startHTTPGateway(ctx); err != nil {
			return err
		}
	}

	<-ctx.Done()
	logger.Info(ctx, "Shutting down servers...")

	a.grpcServer.GracefulStop()
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(context.Background()); err != nil {
			logger.Errorf(ctx, "HTTP server shutdown error: %v", err)
		}
	}

	return nil
}

func (a *App) startHTTPGateway(ctx context.Context) error {
	grpcAddr := fmt.Sprintf("localhost:%d", a.cfg.GRPC.Port)
	httpAddr := a.cfg.GetHTTPAddress()

	gatewayMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			switch key {
			case "Authorization", "authorization":
				return key, true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// Register all services with HTTP gateway
	if err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, gatewayMux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register auth handler: %w", err)
	}
	if err := pb.RegisterOrganizationServiceHandlerFromEndpoint(ctx, gatewayMux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register organization handler: %w", err)
	}
	if err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, gatewayMux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register user handler: %w", err)
	}
	if err := pb.RegisterDocumentServiceHandlerFromEndpoint(ctx, gatewayMux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register document handler: %w", err)
	}
	if err := pb.RegisterNoteServiceHandlerFromEndpoint(ctx, gatewayMux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register note handler: %w", err)
	}
	if err := pb.RegisterContractTemplateServiceHandlerFromEndpoint(ctx, gatewayMux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register template handler: %w", err)
	}
	if err := pb.RegisterGeneratedContractServiceHandlerFromEndpoint(ctx, gatewayMux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register contract handler: %w", err)
	}
	if err := pb.RegisterStorageServiceHandlerFromEndpoint(ctx, gatewayMux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register storage handler: %w", err)
	}

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	gatewayMuxWithCORS := corsHandler.Handler(gatewayMux)

	// Setup HTTP router with additional endpoints
	httpMux := chi.NewRouter()

	httpMux.HandleFunc("/swagger", func(w http.ResponseWriter, request *http.Request) {
		logger.Info(request.Context(), "serving swagger file")
		w.Header().Set("Content-Type", "application/json")
		w.Write(pb.GetSwaggerJSON())
	})
	httpMux.Mount("/docs/", v5emb.NewHandler(
		"Core Service",
		fmt.Sprintf("%s/swagger", a.options.httpPathPrefix),
		fmt.Sprintf("%s/docs/", a.options.httpPathPrefix),
	))

	httpMux.Handle("/metrics", promhttp.Handler())
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	httpMux.Mount("/", http.StripPrefix(a.options.httpPathPrefix, gatewayMuxWithCORS))

	baseMux := chi.NewRouter()
	prefix := a.options.httpPathPrefix
	if prefix == "" {
		prefix = "/"
	}
	baseMux.Mount(prefix, httpMux)

	a.httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: baseMux,
	}

	logger.Infof(ctx, "Starting HTTP gateway on %s with prefix '%s'", httpAddr, prefix)

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf(ctx, "HTTP server error: %v", err)
		}
	}()

	return nil
}
