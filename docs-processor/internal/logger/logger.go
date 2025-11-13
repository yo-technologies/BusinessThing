package logger

import (
	"context"
	"os"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger

func Init() {
	config := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "timestamp",
			LevelKey:     "level",
			MessageKey:   "message",
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeTime:   zapcore.TimeEncoderOfLayout(time.RFC3339),
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	// Добавляем базовые поля, которые будут в каждом логе
	defaultFields := []zap.Field{
		zap.String("service", "llm-agent"),
		zap.String("env", os.Getenv("ENV")),
	}

	var err error
	logger, err = config.Build(zap.Fields(
		defaultFields...,
	))
	if err != nil {
		panic(err)
	}
	sugar = logger.Sugar()
}

func Logger() *zap.Logger {
	return logger
}

// traceField extracts trace id from context (if any) and returns it as zap field
func traceField(ctx context.Context) zap.Field {
	if ctx == nil {
		return zap.Skip()
	}
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return zap.Skip()
	}
	if spanCtx, ok := span.Context().(jaeger.SpanContext); ok {
		return zap.String("trace_id", spanCtx.TraceID().String())
	}
	return zap.Skip()
}

// withCtx returns a logger augmented with context-derived fields (e.g., trace id)
func withCtx(ctx context.Context) *zap.SugaredLogger {
	if logger == nil {
		panic("logger not initialized")
	}
	tf := traceField(ctx)
	if tf.Type != zapcore.SkipType {
		return logger.With(tf).Sugar()
	}
	return sugar
}

func Debug(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Debugw(msg, args...)
}

func Debugf(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Debugf(msg, args...)
}

func Info(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Infow(msg, args...)
}

func Infof(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Infof(msg, args...)
}

func Warn(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Warnw(msg, args...)
}

func Warnf(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Warnf(msg, args...)
}

func Error(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Errorw(msg, args...)
}

func Errorf(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Errorf(msg, args...)
}

func Fatal(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Errorw(msg, args...)
	os.Exit(1)
}

func Fatalf(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Errorf(msg, args...)
	os.Exit(1)
}

func Panic(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Errorw(msg, args...)
	panic(msg)
}

func Panicf(ctx context.Context, msg string, args ...interface{}) {
	withCtx(ctx).Errorf(msg, args...)
	panic(msg)
}
