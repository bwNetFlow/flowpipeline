package skip

import (
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

// Skip Segment test, passthrough test
func TestSegment_Skip_passthrough(t *testing.T) {
	result := segments.TestSegment("skip", map[string]string{},
		&flow.FlowMessage{Type: 3})
	if result.Type != 3 {
		t.Error("Segment Skip is not working.")
	}
}

// Skip Segment test, skip 1 test
func TestSegment_Skip_skip(t *testing.T) {
	segment := segments.LookupSegment("skip").New(map[string]string{"skip": "1", "invert": "0", "condition": "port 3"})

	in, out, altout := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire([]chan *flow.FlowMessage{in, out, altout}, 0, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	in <- &flow.FlowMessage{SrcPort: 3}
	close(in)
	select {
	case _ = <-out:
		t.Error("Segment Skip is not skipping when it is supposed to.")
	case result := <-altout:
		if result.SrcPort != 3 {
			t.Error("Segment Skip is skipping correctly but mangling the data.")
		}
	}
	wg.Wait()
}

// Skip Segment test, no skip 1 test
func TestSegment_Skip_noskip(t *testing.T) {
	segment := segments.LookupSegment("skip").New(map[string]string{"skip": "1", "invert": "0", "condition": "port 4"})

	in, out, altout := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire([]chan *flow.FlowMessage{in, out, altout}, 0, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	in <- &flow.FlowMessage{SrcPort: 3}
	close(in)
	select {
	case result := <-out:
		if result.SrcPort != 3 {
			t.Error("Segment Skip is skipping correctly but mangling the data.")
		}
	case _ = <-altout:
		t.Error("Segment Skip is not skipping when it is supposed to.")
	}
	wg.Wait()

}
