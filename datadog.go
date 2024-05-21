package oteldatadogtie

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/google/uuid"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

var runtimeID = mkRuntimeID()

func RuntimeID() string {
	return runtimeID
}

func StartDDProfiler(opts ...profiler.Option) error {
	opts = append(opts, profiler.WithTags(ext.RuntimeID+":"+runtimeID))
	if err := profiler.Start(opts...); err != nil {
		return fmt.Errorf("unable to start datadog profiler: %w", err)
	}
	return nil
}

var StopDDProfiler = profiler.Stop

func mkRuntimeID() string {
	if id := os.Getenv("OTEL_TRACER_RUNTIME_ID"); id != "" {
		return id
	}
	id := uuid.New()
	return base64.RawURLEncoding.EncodeToString(id[:])
}
