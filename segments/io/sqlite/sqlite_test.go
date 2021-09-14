package sqlite

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

// Sqlite Segment test, passthrough test only
func TestSegment_Sqlite_passthrough(t *testing.T) {
	result := segments.TestSegment("sqlite", map[string]string{"filename": ":memory:"},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment Sqlite is not passing through flows.")
	}
}
