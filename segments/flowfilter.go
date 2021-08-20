package segments

import (
	"log"
	"sync"

	"github.com/bwNetFlow/flowfilter/parser"
	"github.com/bwNetFlow/flowfilter/visitors"
)

type FlowFilter struct {
	BaseSegment
	Filter string
}

func (segment FlowFilter) New(config map[string]string) Segment {
	return &FlowFilter{
		Filter: config["filter"],
	}
}

func (segment *FlowFilter) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	expr, err := parser.Parse(segment.Filter)
	if err != nil {
		log.Printf("[error] FlowFilter: Syntax error in filter expression: %v", err)
		return
	}
	log.Printf("[info] FlowFilter: Using filter expression: %s", segment.Filter)

	filter := &visitors.Filter{}
	for msg := range segment.In {
		if match, err := filter.CheckFlow(expr, msg); match {
			if err != nil {
				log.Printf("[error] FlowFilter: Semantic error in filter expression: %v", err)
				return
			}
			segment.Out <- msg
		}
	}
}

func init() {
	segment := &FlowFilter{}
	RegisterSegment("flowfilter", segment)
}
