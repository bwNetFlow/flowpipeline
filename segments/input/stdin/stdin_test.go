package stdin

import (
	"bufio"
	"context"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
)

// StdIn Segment test, passthrough test only
func TestSegment_StdIn_passthrough(t *testing.T) {
	result := segments.TestSegment("stdin", map[string]string{},
		&pb.EnrichedFlow{})
	if result == nil {
		t.Error("Segment StdIn is not passing through flows.")
	}
}

// Stdin Segment benchmark passthrough
func BenchmarkStdin(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	os.Stdout, _ = os.Open(os.DevNull)

	segment := StdIn{
		scanner: bufio.NewScanner(os.Stdin),
	}

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
