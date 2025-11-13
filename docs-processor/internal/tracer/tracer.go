package tracer

import (
	"context"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

type options struct {
	serviceName       string
	collectorEndpoint string
}

type OptionsFunc func(*options)

func WithServiceName(serviceName string) OptionsFunc {
	return func(o *options) {
		o.serviceName = serviceName
	}
}

func WithCollectorEndpoint(endpoint string) OptionsFunc {
	return func(o *options) {
		o.collectorEndpoint = endpoint
	}
}

func MustSetup(ctx context.Context, opts ...OptionsFunc) io.Closer {
	o := &options{
		serviceName:       "docs-processor",
		collectorEndpoint: "http://localhost:14268/api/traces",
	}

	for _, opt := range opts {
		opt(o)
	}

	cfg := &config.Configuration{
		ServiceName: o.serviceName,
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:          false,
			CollectorEndpoint: o.collectorEndpoint,
		},
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		panic(err)
	}

	opentracing.SetGlobalTracer(tracer)

	return closer
}
