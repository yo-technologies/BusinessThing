package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"docs-processor/internal/app/api/document"
	"docs-processor/internal/app/interceptors"
	"docs-processor/internal/logger"
	"docs-processor/internal/service"
	desc "docs-processor/pkg/document"

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
	bypassCors       bool
}

var defaultOptions = &Options{
	grpcPort:         50052,
	gatewayPort:      8081,
	enableGateway:    true,
	enableReflection: true,
	httpPathPrefix:   "",
	bypassCors:       true,
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
	documentService *document.Service

	options *Options
}

func New(
	searchService *service.SearchService,
	templateProcessor *service.TemplateProcessor,
	options ...OptionsFunc,
) *App {
	opts := defaultOptions
	for _, o := range options {
		o(opts)
	}
	return &App{
		documentService: document.NewService(searchService, templateProcessor),
		options:         opts,
	}
}

func (a *App) Run(ctx context.Context) error {
	grpcEndpoint := fmt.Sprintf(":%d", a.options.grpcPort)
	httpEndpoint := fmt.Sprintf(":%d", a.options.gatewayPort)

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.TracingInterceptor,
			interceptors.RecoveryInterceptor,
			interceptors.ErrCodesInterceptor,
		),
		grpc.ChainStreamInterceptor(
			interceptors.TracingStreamInterceptor,
			interceptors.RecoveryStreamInterceptor,
			interceptors.ErrCodesStreamInterceptor,
		),
	)

	desc.RegisterDocumentServiceServer(srv, a.documentService)

	if a.options.enableReflection {
		reflection.Register(srv)
	}

	if !a.options.enableGateway {
		lis, err := net.Listen("tcp", grpcEndpoint)
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}

		logger.Info(ctx, "starting grpc server", "port", a.options.grpcPort)

		go func() {
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

			<-stop

			logger.Info(ctx, "shutting down server...")
			srv.Stop()
		}()

		if err := srv.Serve(lis); err != nil {
			return err
		}

		logger.Info(ctx, "grpc server stopped")
		return nil
	}

	gatewayMux := runtime.NewServeMux()

	if err := registerGateway(ctx, gatewayMux, grpcEndpoint); err != nil {
		return err
	}

	var corsHandler *cors.Cors
	if a.options.bypassCors {
		corsHandler = cors.AllowAll()
	} else {
		corsHandler = cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		})
	}
	gatewayMuxWithCORS := corsHandler.Handler(gatewayMux)

	httpMux := chi.NewRouter()

	httpMux.HandleFunc("/swagger", func(w http.ResponseWriter, request *http.Request) {
		logger.Info(request.Context(), "serving swagger file")
		w.Header().Set("Content-Type", "application/json")
		w.Write(desc.GetSwaggerJSON())
	})
	httpMux.Mount("/docs/", v5emb.NewHandler(
		"Document Processing Service",
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

	httpSrv := &http.Server{
		Addr:    httpEndpoint,
		Handler: baseMux,
	}

	go func() {
		logger.Info(ctx, "http server listening", "port", a.options.gatewayPort)
		if err := httpSrv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error(ctx, "error starting http server", "error", err)
			}
		}
	}()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

		<-stop

		logger.Info(ctx, "shutting down server...")

		err := httpSrv.Shutdown(ctx)
		if err != nil {
			logger.Error(ctx, "error shutting down http server", "error", err)
		}

		srv.Stop()
	}()

	lis, err := net.Listen("tcp", grpcEndpoint)
	if err != nil {
		return err
	}

	logger.Info(ctx, "grpc server listening", "port", a.options.grpcPort)

	if err := srv.Serve(lis); err != nil {
		return err
	}

	logger.Info(ctx, "grpc server stopped")

	return nil
}

func registerGateway(ctx context.Context, mux *runtime.ServeMux, grpcEndpoint string) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	err := desc.RegisterDocumentServiceHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
	if err != nil {
		return err
	}

	return nil
}
