package segments

import (
	"log"
	"os"
	"sync"

	flow "github.com/bwNetFlow/protobuf/go"
)

var (
	registeredSegments = make(map[string]Segment)
	lock               = &sync.RWMutex{}
)

func RegisterSegment(name string, s Segment) {
	lock.Lock()
	_, ok := registeredSegments[name]
	if ok {
		log.Printf("[error] Segments: Tried to register conflicting segment name '%s'.", name)
		os.Exit(1)
	}
	registeredSegments[name] = s
	lock.Unlock()
}

func LookupSegment(name string) Segment {
	lock.RLock()
	segment, ok := registeredSegments[name]
	lock.RUnlock()
	if !ok {
		log.Printf("[error] Configured segment %s not found, exiting...", name)
		os.Exit(1)
	}
	return segment
}

type Segment interface {
	New(config map[string]string) Segment
	Run(wg *sync.WaitGroup)
	Rewire(<-chan *flow.FlowMessage, chan<- *flow.FlowMessage)
}

type BaseSegment struct {
	in  <-chan *flow.FlowMessage
	out chan<- *flow.FlowMessage
}

func (segment *BaseSegment) Rewire(in <-chan *flow.FlowMessage, out chan<- *flow.FlowMessage) {
	segment.in = in
	segment.out = out
}

// TODO:
// &segments.KafkaProducerSplit{}
// &segments.SNMP{communities: map[string]string{"default": "public"}},
// &segments.BMPInfo{...},
// &segments.Aggregate/GroupBy{Aggr: "SUM", Step: "5m"},
// &segments.PrometheusMetrics{},
