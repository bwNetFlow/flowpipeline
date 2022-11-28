package json

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
)

// Json Segment test, passthrough test only
func TestSegment_Json_passthrough(t *testing.T) {
	result := segments.TestSegment("json", map[string]string{},
		&pb.EnrichedFlow{})
	if result == nil {
		t.Error("Segment Json is not passing through flows.")
	}
}

// Json Segment benchmark passthrough
func BenchmarkJson(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	os.Stdout, _ = os.Open(os.DevNull)

	segment := Json{}.New(map[string]string{})

	in, out := make(chan *pb.FlowContainer), make(chan *pb.FlowContainer)
	segment.Rewire(in, out)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	for n := 0; n < b.N; n++ {
		in <- &pb.FlowContainer{EnrichedFlow: &pb.EnrichedFlow{}, Context: context.Background()}
		_ = <-out
	}
	close(in)
}
