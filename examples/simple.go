package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/GuanceCloud/oteldatadogtie"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

var OtelTracer trace.Tracer

func httpExporter(ctx context.Context) (sdktrace.SpanExporter, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("127.0.0.1:9529"),
		otlptracehttp.WithURLPath("/otel/v1/trace"),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		otlptracehttp.WithTimeout(time.Second*15),
		otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
			Enabled:         true,
			InitialInterval: time.Second,
			MaxInterval:     time.Second * 15,
			MaxElapsedTime:  time.Minute * 2,
		}),
	)

	return exporter, err
}

func stdoutExporter(_ context.Context) (sdktrace.SpanExporter, error) {
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

func main() {
	ctx := context.Background()

	exporter, err := httpExporter(ctx)
	if err != nil {
		log.Fatalf("unable to init exporter: %v", err)
	}

	tp := oteldatadogtie.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(time.Second),
			sdktrace.WithExportTimeout(time.Second*15),
		),
	)

	defer tp.Shutdown(ctx)

	otel.SetTracerProvider(tp)

	// If you need to use custom Resource, you should add attribute
	// oteldatadogtie.AttributeRuntimeID to it, then you can call the
	// Open Telemetry original sdktrace.NewTracerProvider function as usual,
	// at last use oteldatadogtie.Wrap function to wrap your TracerProvider
	// instance, for example:
	//
	// res, err := resource.New(ctx,
	// 	resource.WithProcess(),
	// 	resource.WithAttributes(
	// 		attribute.String("foo", "bar"),
	// 		oteldatadogtie.AttributeRuntimeID,
	// 	),
	// )
	// if err != nil {
	// 	log.Fatalf("unable to build resource: %v", err)
	// }
	// if res, err = resource.Merge(resource.Default(), res); err != nil {
	// 	log.Fatalf("unable to merge resource: %v", err)
	// }
	//
	// tp2 := sdktrace.NewTracerProvider(
	// 	sdktrace.WithBatcher(
	// 		exporter,
	// 		sdktrace.WithBatchTimeout(time.Second),
	// 		sdktrace.WithExportTimeout(time.Second*15),
	// 	),
	// 	sdktrace.WithResource(res),
	// )
	//
	// defer tp2.Shutdown(ctx)
	// otel.SetTracerProvider(oteldatadogtie.Wrap(tp2))

	OtelTracer = otel.Tracer("oteldatadogtie")

	// Use our oteldatadogtie.Start wrapper to start Datadog profiler,
	// or use the original Datadog profiler.Start and combine setting
	// tag oteldatadogtie.TagRuntimeID, see oteldatadogtie.StartDDProfiler
	// for details.
	err = oteldatadogtie.StartDDProfiler(
		profiler.WithProfileTypes(profiler.CPUProfile, profiler.HeapProfile, profiler.MutexProfile, profiler.GoroutineProfile),
		profiler.WithService("oteldatadogtie-demo"),
		profiler.WithEnv("testing"),
		profiler.WithVersion("v0.0.1"),
		profiler.WithAgentAddr("127.0.0.1:9529"),
	)
	if err != nil {
		log.Fatalf("unable to start dd profiler: %v", err)
	}

	defer oteldatadogtie.StopDDProfiler()

	n := 27

	for {
		func() {
			newCtx, span := OtelTracer.Start(ctx, "main", trace.WithAttributes(attribute.Int("n", n)))
			defer span.End()
			fmt.Printf("fibonacci(%d) = %d\n", n, fibonacci(newCtx, n))
			time.Sleep(time.Second * 5)
		}()
	}
}

func fibonacci(ctx context.Context, n int) int {
	newCtx := ctx
	if n%6 == 5 {
		var span trace.Span
		newCtx, span = OtelTracer.Start(ctx, "fibonacci", trace.WithAttributes(attribute.Int("n", n)))
		defer span.End()
	}

	if n < 2 {
		return 1
	}

	return fibonacci(newCtx, n-1) + fibonacci(newCtx, n-2)
}
