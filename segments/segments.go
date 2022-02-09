// This package is home to all pipeline segment implementations. Generally,
// every segment lives in its own package, implements the Segment interface,
// embeds the BaseSegment to take care of the I/O side of things, and has an
// additional init() function to register itself using RegisterSegment.
package segments

import (
	"log"
	"sync"

	flow "github.com/bwNetFlow/protobuf/go"
)

var (
	registeredSegments    = make(map[string]Segment)
	lock                  = &sync.RWMutex{}
	ContainerVolumePrefix = ""
)

// Used by Segments to register themselves in their init() functions. Errors
// and exits immediately on conflicts.
func RegisterSegment(name string, s Segment) {
	_, ok := registeredSegments[name]
	if ok {
		log.Fatalf("[error] Segments: Tried to register conflicting segment name '%s'.", name)
	}
	lock.Lock()
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
		log.Fatalf("[error] Segments: Could not find a segment named '%s'.", name)
	}
	return segment
}

// Used by the tests to run single flow messages through a segment.
func TestSegment(name string, config map[string]string, msg *flow.FlowMessage) *flow.FlowMessage {
	segment := LookupSegment(name).New(config)
	if segment == nil {
		log.Fatalf("[error] Configured segment '%s' could not be initialized properly, see previous messages.", name)
	}

	in, out := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire([]chan *flow.FlowMessage{in, out}, 0, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	in <- msg
	close(in)
	resultMsg := <-out
	wg.Wait()

	return resultMsg
}

// This interface is central to an Pipeline object, as it operates on a list of
// them. In general, Segments should embed the BaseSegment to provide the
// Rewire function and the associated vars.
type Segment interface {
	New(config map[string]string) Segment                     // for reading the provided config
	Run(wg *sync.WaitGroup)                                   // goroutine, must close(segment.Out) when segment.In is closed
	Rewire(chans []chan *flow.FlowMessage, in uint, out uint) // embed this using BaseSegment
}

// Serves as a basis for any Segment implementations. Segments embedding this
// type only need the New and the Run methods to be compliant to the Segment
// interface.
type BaseSegment struct {
	In  <-chan *flow.FlowMessage
	Out chan<- *flow.FlowMessage
}

// An extended basis for Segment implementations in the filter group. It
// contains the necessities to process filtered (dropped) flows.
type BaseFilterSegment struct {
	BaseSegment
	Drops chan<- *flow.FlowMessage
}

// Set a return channel for dropped flow messages. Segments need to be wary of
// this channel closing when producing messages to this channel. This method is
// only called by the flowpipeline tool from the controlflow/branch segment to
// implement the then/else branches, otherwise this functionality is unused.
func (segment *BaseFilterSegment) SubscribeDrops(drops chan<- *flow.FlowMessage) {
	segment.Drops = drops
}

// This function rewires this Segment with the provided channels. This is
// typically called only by pipeline.New() and present in any Segment
// implementation embedding the BaseSegment.
// The peculiar implementation of passing the full channel list and providing
// indexes is due to the fact that controlflow segments may want to skip
// segments and thus need to have all later references available as well.
func (segment *BaseSegment) Rewire(chans []chan *flow.FlowMessage, in uint, out uint) {
	segment.In = chans[in]
	segment.Out = chans[out]
}
