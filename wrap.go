package oteldatadogtie

import (
	"context"
	"runtime/pprof"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

// These constants are copied from https://github.com/DataDog/dd-trace-go/blob/main/internal/traceprof/traceprof.go
const (
	SpanID          = "span id"
	LocalRootSpanID = "local root span id"
	TraceEndpoint   = "trace endpoint"
)

// KeyRuntimeID is used to mark a process uniquely
const KeyRuntimeID attribute.Key = ext.RuntimeID

var AttributeRuntimeID = KeyRuntimeID.String(runtimeID)

func NewTracerProvider(opts ...sdktrace.TracerProviderOption) *TracerProviderWrapper {
	res := resource.Default()
	if merged, err := resource.Merge(res, resource.NewSchemaless(AttributeRuntimeID)); err == nil {
		res = merged
	}
	opts = append(opts, sdktrace.WithResource(res))
	return Wrap(sdktrace.NewTracerProvider(opts...))
}

func Wrap(tp *sdktrace.TracerProvider) *TracerProviderWrapper {
	tp.RegisterSpanProcessor(new(pprofSpanProcessor))

	return &TracerProviderWrapper{
		TracerProvider: tp,
	}
}

type TracerProviderWrapper struct {
	*sdktrace.TracerProvider
}

func (t *TracerProviderWrapper) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	tracer := t.TracerProvider.Tracer(name, options...)
	return &tracerWrapper{
		tracer,
	}
}

type tracerWrapper struct {
	trace.Tracer
}

func (o *tracerWrapper) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	newCtx, span := o.Tracer.Start(ctx, spanName, opts...)
	return newCtx, &spanWrapper{
		Span:      span,
		ParentCtx: ctx,
	}
}

type spanWrapper struct {
	trace.Span
	ParentCtx context.Context
}

func (o *spanWrapper) End(options ...trace.SpanEndOption) {
	if o.ParentCtx != nil {
		defer pprof.SetGoroutineLabels(o.ParentCtx)
	}
	o.Span.End(options...)
}

type pprofSpanProcessor struct{}

func (p *pprofSpanProcessor) OnStart(parent context.Context, s sdktrace.ReadWriteSpan) {
	if !s.Resource().Set().HasValue(KeyRuntimeID) {
		s.SetAttributes(AttributeRuntimeID)
	}

	labels := pprof.Labels(
		SpanID, s.SpanContext().SpanID().String(),
		LocalRootSpanID, s.SpanContext().TraceID().String(),
		TraceEndpoint, s.Name(),
	)

	pprof.SetGoroutineLabels(pprof.WithLabels(parent, labels))
}

func (p *pprofSpanProcessor) OnEnd(sdktrace.ReadOnlySpan) {}

func (p *pprofSpanProcessor) Shutdown(context.Context) error {
	return nil
}

func (p *pprofSpanProcessor) ForceFlush(context.Context) error {
	return nil
}

var _ sdktrace.SpanProcessor = (*pprofSpanProcessor)(nil)
