package prometheusexporter

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

// PrometheusExporter Segment test, passthrough test only
func TestSegment_PrometheusExporter_passthrough(t *testing.T) {
	result := segments.TestSegment("prometheusexporter", map[string]string{"endpoint": ":8080"},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment PrometheusExporter is not passing through flows.")
	}
}
