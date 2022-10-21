package clickhouse_segment

import (
	"sync"
	"testing"

	"github.com/bwNetFlow/flowpipeline/pb"
	// "github.com/bwNetFlow/flowpipeline/segments"
)

// Clickhouse Segment test, passthrough test only
func TestSegment_Clickhouse_passtrough(t *testing.T) {
	segment := Clickhouse{}.New(map[string]string{"filename": "test.sqlite"})

	in, out := make(chan *pb.EnrichedFlow), make(chan *pb.EnrichedFlow)
	segment.Rewire(in, out)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)
	in <- &pb.EnrichedFlow{SrcAddr: []byte{192, 168, 88, 1}, DstAddr: []byte{192, 168, 88, 1}, Proto: 1}
	<-out
	in <- &pb.EnrichedFlow{SrcAddr: []byte{192, 168, 88, 2}, DstAddr: []byte{192, 168, 88, 2}, Proto: 2}
	<-out
	close(in)
	wg.Wait()
}
