package prometheus

import (
	"testing"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

// Prometheus Segment test, passthrough test only
func TestSegment_PrometheusExporter_passthrough(t *testing.T) {
	result := segments.TestSegment("prometheus", map[string]string{"endpoint": ":8080"},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment Prometheus is not passing through flows.")
	}
}
