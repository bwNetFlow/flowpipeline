package count

import (
	"testing"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

// Count Segment test, passthrough test only
func TestSegment_Count_passthrough(t *testing.T) {
	result := segments.TestSegment("count", map[string]string{"prefix": "Test: "},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment Count is not passing through flows.")
	}
}
