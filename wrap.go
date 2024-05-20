package oteldatadogtie

import (
	"context"
	"github.com/GuanceCloud/oteldatadogtie/profiler"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"runtime/pprof"
)

// These constants are copied from https://github.com/DataDog/dd-trace-go/blob/main/internal/traceprof/traceprof.go
const (
	SpanID          = "span id"
	LocalRootSpanID = "local root span id"
	TraceEndpoint   = "trace endpoint"
)

// Below tags are used to mark a process uniquely
const (
	RuntimeIDHyphen    = "runtime-id"
	RuntimeIDUnderline = "runtime_id"
)

func Wrap(tp trace.TracerProvider) trace.TracerProvider {
	return &tracerProviderWrapper{
		TracerProvider: tp,
	}
}

type tracerProviderWrapper struct {
	trace.TracerProvider
}

func (t *tracerProviderWrapper) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
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
	labels := pprof.Labels(
		SpanID, span.SpanContext().SpanID().String(),
		LocalRootSpanID, span.SpanContext().TraceID().String(),
		TraceEndpoint, spanName,
	)

	pprof.SetGoroutineLabels(pprof.WithLabels(ctx, labels))
	return newCtx, &spanWrapper{
		Span:        span,
		PreviousCtx: ctx,
	}
}

type spanWrapper struct {
	trace.Span
	PreviousCtx context.Context
}

func (o *spanWrapper) End(options ...trace.SpanEndOption) {
	if o.PreviousCtx != nil {
		defer pprof.SetGoroutineLabels(o.PreviousCtx)
	}

	attrs := make([]attribute.KeyValue, 0, 2)
	attrs = append(attrs, attribute.String(ext.RuntimeID, profiler.RuntimeID()))
	if ext.RuntimeID != RuntimeIDHyphen {
		attrs = append(attrs, attribute.String(RuntimeIDHyphen, profiler.RuntimeID()))
	}
	if ext.RuntimeID != RuntimeIDUnderline {
		attrs = append(attrs, attribute.String(RuntimeIDUnderline, profiler.RuntimeID()))
	}

	o.Span.SetAttributes(attrs...)
	o.Span.End(options...)
}
