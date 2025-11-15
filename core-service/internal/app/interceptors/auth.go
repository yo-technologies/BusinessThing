package interceptors

import (
	"context"
	"core-service/internal/domain"
	"core-service/internal/jwt"
	"core-service/internal/logger"
	"strings"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewUnaryAuthInterceptor creates an auth interceptor using provided JWT provider.
// Optionally accepts a list of fully-qualified method names that should bypass auth (unprotected).
func NewUnaryAuthInterceptor(provider *jwt.Provider, unprotected ...string) grpc.UnaryServerInterceptor {
	skip := make(map[string]struct{}, len(unprotected))
	for _, m := range unprotected {
		skip[m] = struct{}{}
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		span, ctx := opentracing.StartSpanFromContext(ctx, info.FullMethod)
		defer span.Finish()

		if _, ok := skip[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		token, err := tokenFromMetadata(ctx)
		if err != nil {
			logger.Warnf(ctx, "auth failed in %s: %v", info.FullMethod, err)
			return nil, domain.ErrUnauthorized
		}

		userID, err := provider.ParseToken(ctx, token)
		if err != nil {
			logger.Warnf(ctx, "token parse failed in %s: %v", info.FullMethod, err)
			return nil, err
		}

		span.SetTag("user_id", userID.String())

		ctx = WithUserID(ctx, userID)
		return handler(ctx, req)
	}
}

// NewStreamAuthInterceptor creates a stream auth interceptor using provided JWT provider.
// Optionally accepts a list of fully-qualified method names that should bypass auth (unprotected).
func NewStreamAuthInterceptor(provider *jwt.Provider, unprotected ...string) grpc.StreamServerInterceptor {
	skip := make(map[string]struct{}, len(unprotected))
	for _, m := range unprotected {
		skip[m] = struct{}{}
	}
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		span, ctx := opentracing.StartSpanFromContext(ss.Context(), info.FullMethod)
		defer span.Finish()

		if _, ok := skip[info.FullMethod]; ok {
			return handler(srv, ss)
		}

		token, err := tokenFromMetadata(ss.Context())
		if err != nil {
			logger.Warnf(ctx, "auth failed in stream %s: %v", info.FullMethod, err)
			return domain.ErrUnauthorized
		}

		userID, err := provider.ParseToken(ss.Context(), token)
		if err != nil {
			logger.Warnf(ctx, "token parse failed in stream %s: %v", info.FullMethod, err)
			return err
		}

		span.SetTag("user_id", userID.String())

		ctx = WithUserID(ctx, userID)
		ctx = WithAccessToken(ctx, token)

		wrapped := &wrappedServerStream{ServerStream: ss, ctx: ctx}

		return handler(srv, wrapped)
	}
}

func tokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Warn(ctx, "no metadata in context")
		return "", domain.ErrUnauthorized
	}

	vals := md.Get("authorization")
	if len(vals) == 0 {
		// grpc-gateway may forward as capitalized header key as well; normalize
		vals = md.Get("Authorization")
	}
	if len(vals) == 0 {
		logger.Warn(ctx, "authorization header not found in metadata")
		return "", domain.ErrUnauthorized
	}

	// Extract token (optional "Bearer " prefix)
	token := vals[0]
	if after, ok := strings.CutPrefix(token, "Bearer "); ok {
		token = after
	}
	if token == "" {
		logger.Warn(ctx, "empty token after removing Bearer prefix")
		return "", domain.ErrUnauthorized
	}

	return token, nil
}
