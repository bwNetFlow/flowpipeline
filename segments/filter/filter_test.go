package flowfilter

import (
	"io/ioutil"
	"log"
	"math/rand"
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

// FlowFilter Segment testing is basic, the filtering itself is tested in the flowfilter repo
func TestSegment_FlowFilter_accept(t *testing.T) {
	result := segments.TestSegment("flowfilter", map[string]string{"filter": "proto 4"},
		&flow.FlowMessage{Proto: 4})
	if result == nil {
		t.Error("Segment FlowFilter dropped a flow incorrectly.")
	}
}

func TestSegment_FlowFilter_deny(t *testing.T) {
	result := segments.TestSegment("flowfilter", map[string]string{"filter": "proto 5"},
		&flow.FlowMessage{Proto: 4})
	if result != nil {
		t.Error("Segment FlowFilter accepted a flow incorrectly.")
	}
}

func TestSegment_FlowFilter_syntax(t *testing.T) {
	filter := &FlowFilter{}
	result := filter.New(map[string]string{"filter": "protoo 4"})
	if result != nil {
		t.Error("Segment FlowFilter did something with a syntax error present.")
	}
}

// FlowFilter Segment benchmark passthrough
func BenchmarkFlowFilter(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	os.Stdout, _ = os.Open(os.DevNull)

	segment := FlowFilter{}.New(map[string]string{"filter": "port <50"})

	in, out := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire([]chan *flow.FlowMessage{in, out}, 0, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	for n := 0; n < b.N; n++ {
		in <- &flow.FlowMessage{SrcPort: uint32(rand.Intn(100))}
		_ = <-out
	}
	close(in)
}
