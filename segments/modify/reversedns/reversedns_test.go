package reversedns

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

// TODO: create test with a fixes reverse dns address

// ReverseDns Segment test, passthrough
func TestSegment_ReverseDns_passthrough(t *testing.T) {
	result := segments.TestSegment("reverseDns", map[string]string{},
		&flow.FlowMessage{Bytes: 1})
	if result.Bytes != 1 {
		t.Error("Segment ReverseDns is not working correctly.")
	}
}

// ReverseDns Segment benchmark passthrough
func BenchmarkReverseDns(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	os.Stdout, _ = os.Open(os.DevNull)

	segment := ReverseDns{}.New(map[string]string{})

	in, out := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire(in, out)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	for n := 0; n < b.N; n++ {
		in <- &flow.FlowMessage{SamplingRate: 0, Bytes: 1}
		_ = <-out
	}
	close(in)
}
