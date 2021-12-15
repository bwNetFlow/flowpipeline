package json

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

// Json Segment test, passthrough test only
func TestSegment_Json_passthrough(t *testing.T) {
	result := segments.TestSegment("stdout", map[string]string{},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment Json is not passing through flows.")
	}
}

// Json Segment benchmark passthrough
func BenchmarkJson(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	os.Stdout, _ = os.Open(os.DevNull)

	segment := Json{}.New(map[string]string{})

	in, out := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire([]chan *flow.FlowMessage{in, out}, 0, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	for n := 0; n < b.N; n++ {
		in <- &flow.FlowMessage{}
		_ = <-out
	}
	close(in)
}
