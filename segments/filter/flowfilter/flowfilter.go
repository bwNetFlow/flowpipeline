// Runs flows through a filter and forwards only matching flows. Reuses our own
// https://github.com/bwNetFlow/flowfilter project, see the docs there.
package flowfilter

import (
	"log"
	"sync"

	"github.com/bwNetFlow/flowfilter/parser"
	"github.com/bwNetFlow/flowfilter/visitors"
	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
	"go.opentelemetry.io/otel/attribute"
)

// FIXME: the flowfilter project needs to be updated to new protobuf too
type FlowFilter struct {
	segments.BaseFilterSegment
	Filter string // optional, default is empty

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
	if _, err := filter.CheckFlow(newSegment.expression, &pb.EnrichedFlow{}); err != nil {
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
	for fc := range segment.In {
		_, span := segment.Trace(fc)
		if match, _ := filter.CheckFlow(segment.expression, fc.EnrichedFlow); match {
			if span != nil {
				span.SetAttributes(attribute.KeyValue{Key: "dropped", Value: attribute.BoolValue(false)})
				span.End()
			}
			segment.Out <- fc
		} else {
			if span != nil {
				span.SetAttributes(attribute.KeyValue{Key: "dropped", Value: attribute.BoolValue(true)})
				span.End()
			}
			if segment.Drops != nil {
				segment.Drops <- fc
			} else {
				// if dropped flows don't continue somewhere, end their flow span as well
				if fc.FlowSpan != nil {
					fc.FlowSpan.End()
				}
			}
		}
	}
}

func init() {
	segment := &FlowFilter{}
	segments.RegisterSegment("flowfilter", segment)
}
