package influx

import (
	"testing"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

// Influx Segment test, passthrough test only
func TestSegment_Influx_passthrough(t *testing.T) {
	result := segments.TestSegment("influx", map[string]string{"org": "testorg", "bucket": "testbucket", "token": "testtoken"},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment Influx is not passing through flows.")
	}
}
