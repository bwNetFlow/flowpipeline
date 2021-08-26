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

func TestSegment_PrintDots_instanciation(t *testing.T) {
	printDots := &PrintDots{}
	result := printDots.New(map[string]string{})
	if result == nil {
		t.Error("Segment PrintDots did not intiate despite good base config.")
	}

	printDots = &PrintDots{}
	result = printDots.New(map[string]string{"flowsperdot": "twelve"})
	if result == nil {
		t.Error("Segment PrintDots did not fallback from bad base config.")
	}

	printDots = &PrintDots{}
	result = printDots.New(map[string]string{"flowsperdot": "12"})
	if result == nil {
		t.Error("Segment PrintDots did not intiate despite good base config.")
	}
}
