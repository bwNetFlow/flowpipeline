package aggregate

import (
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type Aggregate struct {
	segments.BaseSegment

	cache FlowExporter
}

func (segment Aggregate) New(config map[string]string) segments.Segment {
	return &Aggregate{cache: FlowExporter{}}
}

func (segment *Aggregate) Run(wg *sync.WaitGroup) {
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
			segment.cache.InsertFlow(msg)
		case msg, ok := <-segment.cache.Flows:
			if !ok {
				return
			}
			segment.Out <- msg
		}
	}
}

func init() {
	segment := &Aggregate{}
	segments.RegisterSegment("aggregate", segment)
}
