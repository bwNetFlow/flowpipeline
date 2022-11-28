package csv

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"
	"testing"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
)

// Csv Segment test, passthrough test
func TestSegment_Csv_passthrough(t *testing.T) {
	result := segments.TestSegment("csv", map[string]string{},
		&pb.EnrichedFlow{Type: 3, SamplerAddress: net.ParseIP("192.0.2.1")})

	if result.Type != 3 {
		t.Error("Segment Csv is not working.")
	}
}

// Csv Segment benchmark passthrough
func BenchmarkCsv(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	os.Stdout, _ = os.Open(os.DevNull)

	segment := Csv{}.New(map[string]string{})

	in, out := make(chan *pb.FlowContainer), make(chan *pb.FlowContainer)
	segment.Rewire(in, out)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	for n := 0; n < b.N; n++ {
		in <- &pb.FlowContainer{EnrichedFlow: &pb.EnrichedFlow{Proto: 45}, Context: context.Background()}
		_ = <-out
	}
	close(in)
}
