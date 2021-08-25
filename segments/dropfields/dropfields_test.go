package dropfields

import (
	"log"
	"os"
	"testing"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
	"github.com/hashicorp/logutils"
)

func TestMain(m *testing.M) {
	log.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"info", "warning", "error"},
		MinLevel: logutils.LogLevel("info"),
		Writer:   os.Stderr,
	})
	code := m.Run()
	os.Exit(code)
}

// DropFields Segment tests are thorough and try every combination
func TestSegment_DropFields_policyKeep(t *testing.T) {
	result := segments.TestSegment("dropfields", map[string]string{"policy": "keep", "fields": "DstAddr"},
		&flow.FlowMessage{SrcAddr: []byte{192, 168, 88, 142}, DstAddr: []byte{192, 168, 88, 143}},
	)
	if len(result.SrcAddr) != 0 || len(result.DstAddr) == 0 {
		t.Error("Segment DropFields is not keeping the proper fields.")
	}
}

func TestSegment_DropFields_policyDrop(t *testing.T) {
	result := segments.TestSegment("dropfields", map[string]string{"policy": "drop", "fields": "SrcAddr"},
		&flow.FlowMessage{SrcAddr: []byte{192, 168, 88, 142}, DstAddr: []byte{192, 168, 88, 143}},
	)
	if len(result.SrcAddr) != 0 || len(result.DstAddr) == 0 {
		t.Error("Segment DropFields is not dropping the proper fields.")
	}
}
