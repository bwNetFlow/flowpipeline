package protomap

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

// Protomap Segment test, passthrough test only
func TestSegment_protomap_passthrough(t *testing.T) {
	result := segments.TestSegment("protomap", map[string]string{},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment protomap is not passing through flows.")
	}
}

// Protomap Segment test, passthrough test only
func TestSegment_protomap_tcp(t *testing.T) {
	result := segments.TestSegment("protomap", map[string]string{},
		&flow.FlowMessage{Proto: 6})
	if result.ProtoName != "TCP" {
		t.Error("Segment protomap is not tagging ProtoName correctly.")
	}
}

// Protomap Segment benchmark passthrough
func BenchmarkProtomap(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	os.Stdout, _ = os.Open(os.DevNull)

	segment := Protomap{}.New(map[string]string{})

	in, out := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire([]chan *flow.FlowMessage{in, out}, 0, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	for n := 0; n < b.N; n++ {
		in <- &flow.FlowMessage{Proto: 6}
		_ = <-out
	}
	close(in)
}
