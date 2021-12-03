package stdin

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

// StdIn Segment test, passthrough test only
func TestSegment_StdIn_passthrough(t *testing.T) {
	result := segments.TestSegment("stdin", map[string]string{},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment StdIn is not passing through flows.")
	}
}
