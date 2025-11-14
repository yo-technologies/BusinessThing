package interceptors

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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

func TracingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Пытаемся извлечь trace_id из метаданных
	var span opentracing.Span
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		carrier := metadataReaderWriter{md}
		spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
		if err == nil {
			span = opentracing.GlobalTracer().StartSpan(info.FullMethod, ext.RPCServerOption(spanCtx))
			ctx = opentracing.ContextWithSpan(ctx, span)
		}
	}

	// Если span не был создан выше, создаем обычный span
	if span == nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, info.FullMethod)
	}
	defer span.Finish()

	result, err := handler(ctx, req)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.message", err)
	}

	return result, err
}

func TracingStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	span, ctx := opentracing.StartSpanFromContext(ss.Context(), info.FullMethod)
	defer span.Finish()

	wrapped := &wrappedServerStream{ServerStream: ss, ctx: ctx}
	err := handler(srv, wrapped)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.message", err)
	}

	return err
}
