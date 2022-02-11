package printflowdump

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
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

// PrintFlowdump Segment benchmark passthrough
func BenchmarkPrintFlowdump(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	os.Stdout, _ = os.Open(os.DevNull)

	segment := PrintFlowdump{}

	in, out := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire(in, out)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	for n := 0; n < b.N; n++ {
		in <- &flow.FlowMessage{}
		_ = <-out
	}
	close(in)
}
