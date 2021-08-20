package segments

import (
	"log"
	"sync"
)

type Count struct {
	BaseSegment
	count  uint64
	Prefix string
}

func (segment Count) New(config map[string]string) Segment {
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
	RegisterSegment("count", segment)
}
