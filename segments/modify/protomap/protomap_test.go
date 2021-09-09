package protomap

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

// Protomap Segment test, passthrough test only
func TestSegment_protomap_passthrough(t *testing.T) {
	result := segments.TestSegment("protomap", map[string]string{},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment protomap is not passing through flows.")
	}
}
