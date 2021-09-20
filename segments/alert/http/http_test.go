package http

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

// Http Segment test, passthrough test
func TestSegment_Http_passthrough(t *testing.T) {
	result := segments.TestSegment("http", map[string]string{"url": "http://localhost:8000"},
		&flow.FlowMessage{Type: 3})
	if result.Type != 3 {
		t.Error("Segment Http is not working.")
	}
}
