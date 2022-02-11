package http

import (
	"testing"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

// Http Segment test, passthrough test
func TestSegment_Http_passthrough(t *testing.T) {
	result := segments.TestSegment("http", map[string]string{"url": "http://localhost:8000"},
		&flow.FlowMessage{Type: 3})
	if result.Type != 3 {
		t.Error("Segment Http is not working.")
	}
}
