package addstrings

import (
	"github.com/bwNetFlow/flowpipeline/segments"
	"sync"
)

type AddrStrings struct {
	segments.BaseSegment
}

func (segment AddrStrings) New(config map[string]string) segments.Segment {
	return &AddrStrings{}
}

func (segment *AddrStrings) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	for original := range segment.In {
		segment.Out <- original
	}
}

// register segment
func init() {
	segment := &AddrStrings{}
	segments.RegisterSegment("addrstrings", segment)
}
