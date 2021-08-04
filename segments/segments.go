package segments

import (
	flow "github.com/bwNetFlow/protobuf/go"
	"sync"
)

type Segment interface {
	New(config map[string]string) Segment
	Run(wg *sync.WaitGroup)
	Rewire(<-chan *flow.FlowMessage, chan<- *flow.FlowMessage)
}

type BaseSegment struct {
	in  <-chan *flow.FlowMessage
	out chan<- *flow.FlowMessage
}

func (segment *BaseSegment) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()
	for msg := range segment.in {
		segment.out <- msg
	}
}

func (segment *BaseSegment) Rewire(in <-chan *flow.FlowMessage, out chan<- *flow.FlowMessage) {
	segment.in = in
	segment.out = out
}

type NoOp struct {
	BaseSegment
}

func (segment NoOp) New(config map[string]string) Segment {
	return &NoOp{}
}

// TODO:
// &segments.KafkaProducerSplit{}
// &segments.SNMP{communities: map[string]string{"default": "public"}},
// &segments.BMPInfo{...},
// &segments.Aggregate/GroupBy{Aggr: "SUM", Step: "5m"},
// &segments.PrometheusMetrics{},
