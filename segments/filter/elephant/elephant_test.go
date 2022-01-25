package elephant

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

// Elephant Segment test, passthrough test
func TestSegment_Elephant_passthrough(t *testing.T) {
	segment := segments.LookupSegment("elephant").New(map[string]string{})
	if segment == nil {
		log.Fatal("[error] Configured segment 'elephant' could not be initialized properly, see previous messages.")
	}

	in, out := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire([]chan *flow.FlowMessage{in, out}, 0, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	in <- &flow.FlowMessage{Bytes: 10}
	in <- &flow.FlowMessage{Bytes: 100}
	result := <-out
	if result.Bytes != 100 {
		t.Error("Segment Elephant is not working.")
	}
	close(in)
	wg.Wait()
}

// Elephant Segment benchmark passthrough
func BenchmarkElephant(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	os.Stdout, _ = os.Open(os.DevNull)

	segment := Elephant{}

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
