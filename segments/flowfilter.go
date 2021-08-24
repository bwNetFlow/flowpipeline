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

	expression *parser.Expression
}

func (segment FlowFilter) New(config map[string]string) Segment {
	var err error

	newSegment := &FlowFilter{
		Filter: config["filter"],
	}

	newSegment.expression, err = parser.Parse(config["filter"])
	if err != nil {
		log.Printf("[error] FlowFilter: Syntax error in filter expression: %v", err)
		return nil
	}
	return newSegment
}

func (segment *FlowFilter) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()

	log.Printf("[info] FlowFilter: Using filter expression: %s", segment.Filter)

	filter := &visitors.Filter{}
	for msg := range segment.in {
		if match, err := filter.CheckFlow(segment.expression, msg); match {
			if err != nil {
				log.Printf("[error] FlowFilter: Semantic error in filter expression: %v", err)
				continue // TODO: introduce option on-error action, current state equals 'drop'
			}
			segment.out <- msg
		}
	}
}

func init() {
	segment := &FlowFilter{}
	RegisterSegment("flowfilter", segment)
}
