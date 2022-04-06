package influx

import (
	"testing"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
)

// Influx Segment test, passthrough test only
func TestSegment_Influx_passthrough(t *testing.T) {
	result := segments.TestSegment("influx", map[string]string{"org": "testorg", "bucket": "testbucket", "token": "testtoken"},
		&pb.EnrichedFlow{})
	if result == nil {
		t.Error("Segment Influx is not passing through flows.")
	}
}
