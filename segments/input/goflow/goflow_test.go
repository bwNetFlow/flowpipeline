package goflow

import (
	"testing"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

// Goflow Segment test, passthrough test only, functionality is tested by Goflow package
func TestSegment_Goflow_passthrough(t *testing.T) {
	result := segments.TestSegment("goflow", map[string]string{"port": "2055"},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment Goflow is not passing through flows.")
	}
}
