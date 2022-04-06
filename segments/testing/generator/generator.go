package generator

import (
	"sync"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
)

type Generator struct {
	segments.BaseSegment
}

func (segment Generator) New(config map[string]string) segments.Segment {
	return &Generator{}
}

func (segment *Generator) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	for {
		select {
		case msg, ok := <-segment.In:
			if !ok {
				return
			}
			segment.Out <- msg
		default:
			segment.Out <- &pb.EnrichedFlow{Proto: 6, Bytes: 42, Note: "generated test flow"}
		}
	}
}

func init() {
	segment := &Generator{}
	segments.RegisterSegment("generator", segment)
}
