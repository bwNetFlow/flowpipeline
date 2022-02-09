package drop

import (
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

type Drop struct {
	segments.BaseSegment
	drops chan<- *flow.FlowMessage
}

func (segment Drop) New(config map[string]string) segments.Segment {
	return &Drop{}
}

func (segment *Drop) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	for msg := range segment.In {
		if segment.drops != nil {
			segment.drops <- msg
			if r := recover(); r != nil {
				segment.drops = nil
			}
		}
	}
}

func (segment *Drop) SubscribeDrops(drops chan<- *flow.FlowMessage) {
	segment.drops = drops
}

func init() {
	segment := &Drop{}
	segments.RegisterSegment("drop", segment)
}
