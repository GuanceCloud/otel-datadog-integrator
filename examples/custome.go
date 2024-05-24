package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/sdk/resource"
	"log"
	"time"

	"github.com/GuanceCloud/oteldatadogtie"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func main() {
	ctx := context.Background()

	exporter, err := httpExporter(ctx)
	if err != nil {
		log.Fatalf("unable to init exporter: %v", err)
	}

	res, err := resource.New(ctx,
		resource.WithProcess(),
		resource.WithAttributes(
			attribute.String("foo", "bar"),
			oteldatadogtie.AttributeRuntimeID,
		),
	)
	if err != nil {
		log.Fatalf("unable to build resource: %v", err)
	}
	if res, err = resource.Merge(resource.Default(), res); err != nil {
		log.Fatalf("unable to merge resource: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithBatchTimeout(time.Second),
			sdktrace.WithExportTimeout(time.Second*15),
		),
		sdktrace.WithResource(res),
	)

	defer tp.Shutdown(ctx)

	wrapped := oteldatadogtie.Wrap(tp)

	otel.SetTracerProvider(wrapped)

	OtelTracer = otel.Tracer("oteldatadogtie")

	err = profiler.Start(
		profiler.WithProfileTypes(profiler.CPUProfile, profiler.HeapProfile, profiler.MutexProfile, profiler.GoroutineProfile),
		profiler.WithService("oteldatadogtie-demo"),
		profiler.WithEnv("testing"),
		profiler.WithVersion("v0.0.1"),
		profiler.WithAgentAddr("127.0.0.1:9529"),
		profiler.WithTags(
			"other-tag:foo",
			"another-tag:bar",
			oteldatadogtie.TagRuntimeID,
		),
	)
	if err != nil {
		log.Fatalf("unable to start dd profiler: %v", err)
	}

	defer profiler.Stop()

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
