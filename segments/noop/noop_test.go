package noop

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

// NoOp Segment test, passthrough test
func TestSegment_NoOp_passthrough(t *testing.T) {
	result := segments.TestSegment("noop", map[string]string{},
		&flow.FlowMessage{Type: 3})
	if result.Type != 3 {
		t.Error("Segment NoOp is not working.")
	}
}
