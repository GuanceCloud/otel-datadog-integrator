package oteldatadogtie

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/GuanceCloud/oteldatadogtie/profiler"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	ddprofiler "gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func newTraceProvider() trace.TracerProvider {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		panic(err)
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(time.Second)),
	)
}

var tracer trace.Tracer

func TestWrap(t *testing.T) {
	err := profiler.Start(
		ddprofiler.WithProfileTypes(ddprofiler.CPUProfile, ddprofiler.HeapProfile),
		ddprofiler.WithTags("foo:bar", "hello:world"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer profiler.Stop()

	otel.SetTracerProvider(Wrap(newTraceProvider()))
	tracer = otel.Tracer("testing")

	_, span := tracer.Start(context.Background(), "test-wrap")
	defer span.End()

	// your code here...
}
