package goflow

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

// Goflow Segment test, passthrough test only, functionality is tested by Goflow package
func TestSegment_Goflow_passthrough(t *testing.T) {
	result := segments.TestSegment("goflow", map[string]string{"port": "2055"},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment Goflow is not passing through flows.")
	}
}
