package oteldatadogtie

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/google/uuid"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

var (
	runtimeID    = mkRuntimeID()
	TagRuntimeID = ext.RuntimeID + ":" + runtimeID
)

func RuntimeID() string {
	return runtimeID
}

// StartDDProfiler is a wrapper of Datadog profiler.Start function,
// the purpose is simply to set the tag runtime-id to our custom value, you
// can also use the original Datadog profiler.Start function (if required) and
// combining setting the TagRuntimeID tag.
func StartDDProfiler(opts ...profiler.Option) error {
	opts = append(opts, profiler.WithTags(TagRuntimeID))
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
