package http

import (
	"testing"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
)

// Http Segment test, passthrough test
func TestSegment_Http_passthrough(t *testing.T) {
	result := segments.TestSegment("http", map[string]string{"url": "http://localhost:8000"},
		&pb.EnrichedFlow{Type: 3})
	if result.Type != 3 {
		t.Error("Segment Http is not working.")
	}
}
