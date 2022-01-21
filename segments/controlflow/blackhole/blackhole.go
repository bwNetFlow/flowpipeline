package skip

import (
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type Blackhole struct {
	segments.BaseSegment
}

func (segment Blackhole) New(config map[string]string) segments.Segment {
	return &Blackhole{}
}

func (segment *Blackhole) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	for range segment.In {
	}
}

func init() {
	segment := &Blackhole{}
	segments.RegisterSegment("skip", segment)
}
