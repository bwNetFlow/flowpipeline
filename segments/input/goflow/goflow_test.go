package goflow

import (
	"testing"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
)

// Goflow Segment test, passthrough test only, functionality is tested by Goflow package
func TestSegment_Goflow_passthrough(t *testing.T) {
	result := segments.TestSegment("goflow", map[string]string{"port": "2055"},
		&pb.EnrichedFlow{})
	if result == nil {
		t.Error("Segment Goflow is not passing through flows.")
	}
}
