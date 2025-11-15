package grpcutils

import (
	"context"
	"strings"

	"docs-processor/internal/logger"

	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// metadataReaderWriter удовлетворяет интерфейсам opentracing.TextMapReader и opentracing.TextMapWriter
// для передачи trace контекста через gRPC metadata
type metadataReaderWriter struct {
	metadata.MD
}

func (w metadataReaderWriter) Set(key, val string) {
	// gRPC HPACK отклоняет любые ключи в верхнем регистре
	// Поскольку HTTP_HEADERS формат case-insensitive, мы приводим к нижнему регистру
	key = strings.ToLower(key)
	w.MD[key] = append(w.MD[key], val)
}

func (w metadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// UnaryClientInterceptor добавляет trace_id в метаданные gRPC запроса
func UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md := metadata.New(nil)

	// Инжектируем tracing context в metadata через правильный carrier
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		carrier := metadataReaderWriter{md}
		if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier); err != nil {
			logger.Warnf(ctx, "failed to inject tracing context: %v", err)
		}
	}

	logger.Debugf(ctx, "gRPC Client Interceptor - Method: %s, Metadata: %v", method, md)

	// Если есть метаданные для добавления, добавляем их в контекст
	if len(md) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}
