// Runs flows through a filter and forwards only matching flows. Reuses our own
// https://github.com/bwNetFlow/flowfilter project, see the docs there.
package flowfilter

import (
	"log"
	"sync"

	"github.com/bwNetFlow/flowfilter/parser"
	"github.com/bwNetFlow/flowfilter/visitors"
	"github.com/bwNetFlow/flowpipeline/segments"

	flow "github.com/bwNetFlow/protobuf/go"
)

type FlowFilter struct {
	segments.BaseSegment
	Filter string // optional, default is empty

	drops      chan<- *flow.FlowMessage
	expression *parser.Expression
}

func (segment FlowFilter) New(config map[string]string) segments.Segment {
	var err error

	newSegment := &FlowFilter{
		Filter: config["filter"],
	}

	newSegment.expression, err = parser.Parse(config["filter"])
	if err != nil {
		log.Printf("[error] FlowFilter: Syntax error in filter expression: %v", err)
		return nil
	}
	filter := &visitors.Filter{}
	if _, err := filter.CheckFlow(newSegment.expression, &flow.FlowMessage{}); err != nil {
		log.Printf("[error] FlowFilter: Semantic error in filter expression: %v", err)
		return nil
	}
	return newSegment
}

func (segment *FlowFilter) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	log.Printf("[info] FlowFilter: Using filter expression: %s", segment.Filter)

	filter := &visitors.Filter{}
	for msg := range segment.In {
		if match, _ := filter.CheckFlow(segment.expression, msg); match {
			segment.Out <- msg
		} else if segment.drops != nil {
			segment.drops <- msg
			if r := recover(); r != nil {
				segment.drops = nil
			}
		}
	}
}

func (segment *FlowFilter) SubscribeDrops(drops chan<- *flow.FlowMessage) {
	segment.drops = drops
}

func init() {
	segment := &FlowFilter{}
	segments.RegisterSegment("flowfilter", segment)
}
