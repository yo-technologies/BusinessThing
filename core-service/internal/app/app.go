package app

import (
	"context"
	"core-service/internal/app/core-service/api/contract"
	"core-service/internal/app/core-service/api/document"
	"core-service/internal/app/core-service/api/note"
	"core-service/internal/app/core-service/api/organization"
	"core-service/internal/app/core-service/api/template"
	"core-service/internal/app/core-service/api/user"
	"core-service/internal/app/interceptors"
	"core-service/internal/config"
	"core-service/internal/jwt"
	"core-service/internal/logger"
	pb "core-service/pkg/core/api/core"
	"fmt"
	"net"
	"net/http"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type App struct {
	cfg             *config.Config
	jwtProvider     *jwt.Provider
	orgService      *organization.Service
	userService     *user.Service
	docService      *document.Service
	noteService     *note.Service
	templateService *template.Service
	contractService *contract.Service
	grpcServer      *grpc.Server
	httpServer      *http.Server
}

type Services struct {
	OrganizationService organization.OrganizationService
	UserService         user.UserService
	DocumentService     document.DocumentService
	NoteService         note.NoteService
	TemplateService     template.TemplateService
	ContractService     contract.ContractService
}

func New(cfg *config.Config, jwtProvider *jwt.Provider, services Services) *App {
	app := &App{
		cfg:             cfg,
		jwtProvider:     jwtProvider,
		orgService:      organization.NewService(services.OrganizationService),
		userService:     user.NewService(services.UserService),
		docService:      document.NewService(services.DocumentService),
		noteService:     note.NewService(services.NoteService),
		templateService: template.NewService(services.TemplateService),
		contractService: contract.NewService(services.ContractService),
	}

	app.setupGRPC()
	return app
}

func (a *App) setupGRPC() {
	// Unprotected methods (public endpoints)
	unprotected := []string{
		"/core.UserService/AcceptInvitation",
	}

	// Setup interceptors
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		interceptors.RecoveryInterceptor,
		interceptors.TracingInterceptor,
		interceptors.NewUnaryAuthInterceptor(a.jwtProvider, unprotected...),
		interceptors.ErrCodesInterceptor,
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		interceptors.RecoveryStreamInterceptor,
		interceptors.TracingStreamInterceptor,
		interceptors.NewStreamAuthInterceptor(a.jwtProvider, unprotected...),
		interceptors.ErrCodesStreamInterceptor,
	}

	// Create gRPC server
	a.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaryInterceptors...)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(streamInterceptors...)),
	)

	// Register services
	pb.RegisterOrganizationServiceServer(a.grpcServer, a.orgService)
	pb.RegisterUserServiceServer(a.grpcServer, a.userService)
	pb.RegisterDocumentServiceServer(a.grpcServer, a.docService)
	pb.RegisterNoteServiceServer(a.grpcServer, a.noteService)
	pb.RegisterContractTemplateServiceServer(a.grpcServer, a.templateService)
	pb.RegisterGeneratedContractServiceServer(a.grpcServer, a.contractService)

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

	mux := runtime.NewServeMux(
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
	if err := pb.RegisterOrganizationServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register organization handler: %w", err)
	}
	if err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register user handler: %w", err)
	}
	if err := pb.RegisterDocumentServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register document handler: %w", err)
	}
	if err := pb.RegisterNoteServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register note handler: %w", err)
	}
	if err := pb.RegisterContractTemplateServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register template handler: %w", err)
	}
	if err := pb.RegisterGeneratedContractServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register contract handler: %w", err)
	}

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	a.httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: corsHandler.Handler(mux),
	}

	logger.Infof(ctx, "Starting HTTP gateway on %s", httpAddr)

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf(ctx, "HTTP server error: %v", err)
		}
	}()

	return nil
}
