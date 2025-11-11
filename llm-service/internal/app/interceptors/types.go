package interceptors

import (
	"context"
	"errors"

	"llm-service/internal/domain"

	"google.golang.org/grpc"
)

type contextKey string

// wrappedServerStream wraps grpc.ServerStream to allow context overriding
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

const (
	userIDContextKey      contextKey = "auth.user_id"
	accessTokenContextKey contextKey = "auth.access_token"
)

// WithUserID stores domain.ID in context for downstream handlers
func WithUserID(ctx context.Context, userID domain.ID) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

// UserIDFromContext extracts domain.ID from context that was previously stored by auth interceptor
func UserIDFromContext(ctx context.Context) (domain.ID, error) {
	val := ctx.Value(userIDContextKey)
	if val == nil {
		return domain.ID{}, domain.ErrUnauthorized
	}
	id, ok := val.(domain.ID)
	if !ok {
		return domain.ID{}, errors.New("invalid user id in context")
	}
	return id, nil
}

// WithAccessToken stores access token string in context for downstream handlers
func WithAccessToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, accessTokenContextKey, token)
}

// AccessTokenFromContext extracts access token string from context that was previously stored by auth interceptor
func AccessTokenFromContext(ctx context.Context) (string, error) {
	val := ctx.Value(accessTokenContextKey)
	if val == nil {
		return "", domain.ErrUnauthorized
	}
	token, ok := val.(string)
	if !ok {
		return "", errors.New("invalid access token in context")
	}
	return token, nil
}
