package drop

import (
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type Drop struct {
	segments.BaseSegment
}

func (segment Drop) New(config map[string]string) segments.Segment {
	return &Drop{}
}

func (segment *Drop) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	for range segment.In {
	}
}

func init() {
	segment := &Drop{}
	segments.RegisterSegment("drop", segment)
}
