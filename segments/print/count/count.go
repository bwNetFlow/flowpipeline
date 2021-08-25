// Counts the number of passing flows and prints the result on termination.
package count

import (
	"log"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type Count struct {
	segments.BaseSegment
	count  uint64
	Prefix string
}

func (segment Count) New(config map[string]string) segments.Segment {
	return &Count{
		Prefix: config["prefix"],
	}
}

func (segment *Count) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	for msg := range segment.In {
		segment.count += 1
		segment.Out <- msg
	}
	// use log without level to print to stderr but never filter it
	log.Printf("%s%d", segment.Prefix, segment.count)
}

func init() {
	segment := &Count{}
	segments.RegisterSegment("count", segment)
}
