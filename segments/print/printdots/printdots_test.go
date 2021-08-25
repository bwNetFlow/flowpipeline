package printdots

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

// PrintDots Segment test, passthrough test only
func TestSegment_PrintDots_passthrough(t *testing.T) {
	result := segments.TestSegment("printdots", map[string]string{"flowsPerDot": "100"},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment PrintDots is not passing through flows.")
	}
}
