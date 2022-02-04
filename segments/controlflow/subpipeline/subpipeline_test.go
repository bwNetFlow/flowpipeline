package subpipeline

import (
	"log"
	"os"
	"testing"

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

// SubPipeline Segment test, passthrough test
func TestSegment_SubPipeline_passthrough(t *testing.T) {
	// TODO FIXME: this is currently not testable using TestSegment, as
	// that does not embed subpipelines into segments

	// result := segments.TestSegment("subpipeline", map[string]string{},
	// 	&flow.FlowMessage{Type: 3})
	// if result.Type != 3 {
	// 	t.Error("Segment SubPipeline is not working.")
	// }
}
