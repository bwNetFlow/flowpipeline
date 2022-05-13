package count

import (
	"testing"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
)

// Count Segment test, passthrough test only
func TestSegment_Count_passthrough(t *testing.T) {
	result := segments.TestSegment("count", map[string]string{"prefix": "Test: "},
		&pb.EnrichedFlow{})
	if result == nil {
		t.Error("Segment Count is not passing through flows.")
	}
}
