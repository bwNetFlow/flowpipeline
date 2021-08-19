// This package is home to all pipeline segment implementations. Generally,
// every segment lives in its own file, implements the Segment interface,
// embeds the BaseSegment to take care of the I/O side of things, and has an
// additional init() function to register itself using RegisterSegment.
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

// Used by Segments to register themselves in their init() functions. Errors
// and exits immediately on conflicts.
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

// Used by the pipeline package to convert segment names in configuration to
// actual Segment objects.
func LookupSegment(name string) Segment {
	lock.RLock()
	segment, ok := registeredSegments[name]
	lock.RUnlock()
	if !ok {
		// think about having it log the error and introduce NoOp instead of quitting
		log.Printf("[error] Configured segment %s not found, exiting...", name)
		os.Exit(1)
	}
	return segment
}

// This interface is central to an Pipeline object, as it operates on a list of
// them. In general, Segments should embed the BaseSegment to provide the
// Rewire function and the associated vars.
type Segment interface {
	New(config map[string]string) Segment                      // for reading the provided config
	Run(wg *sync.WaitGroup)                                    // goroutine, must close(segment.out) when segment.in is closed
	Rewire(<-chan *flow.FlowMessage, chan<- *flow.FlowMessage) // embed this using BaseSegment
}

// Serves as a basis for any Segment implementations. Segments embedding this
// type only need the New and the Run methods to be compliant to the Segment
// interface.
type BaseSegment struct {
	in  <-chan *flow.FlowMessage
	out chan<- *flow.FlowMessage
}

// This function rewires this Segment with the provided channels. This is
// typically called only by pipeline.New() and present in any Segment
// implementation.
func (segment *BaseSegment) Rewire(in <-chan *flow.FlowMessage, out chan<- *flow.FlowMessage) {
	segment.in = in
	segment.out = out
}

// TODO:
// &segments.KafkaProducerSplit{}
// &segments.SNMP{communities: map[string]string{"default": "public"}},
// &segments.BMPInfo{...},
// &segments.Aggregate/GroupBy{Aggr: "SUM", Step: "5m"},
