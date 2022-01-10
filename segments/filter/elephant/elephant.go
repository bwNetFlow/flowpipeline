// Filters out the bulky average of flows.
package elephant

import (
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type Elephant struct {
	segments.BaseSegment
}

func (segment Elephant) New(config map[string]string) segments.Segment {
	return &Elephant{}
}

func (segment *Elephant) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	for msg := range segment.In {
		segment.Out <- msg
	}
}

func init() {
	segment := &Elephant{}
	segments.RegisterSegment("elephant", segment)
}
