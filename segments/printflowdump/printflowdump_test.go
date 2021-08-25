package printflowdump

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

// PrintFlowdump Segment test, passthrough test only
func TestSegment_PrintFlowdump_passthrough(t *testing.T) {
	result := segments.TestSegment("printflowdump", map[string]string{},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment PrintFlowDump is not passing through flows.")
	}
}
