package segments

import "sync"

// The NoOp Segment is considered a template for any additional Segments, as it
// showcases the exact implementation.
type NoOp struct {
	BaseSegment // always embed this, no need to repeat I/O chan code
	// add any additional fields here
}

// Every Segment must implement a New method, even if there isn't any config
// it is interested in.
func (segment NoOp) New(config map[string]string) Segment {
	// do config stuff here, add it to fields maybe
	return &NoOp{}
}

// The main goroutine of any Segment. Any Run method must:
// 1. close(segment.Out) when done, usually when segment.In is closed by the previous segment or the Pipeline itself.
// 2. call wg.Done() before exiting
func (segment *NoOp) Run(wg *sync.WaitGroup) {
	defer func() {
		// This defer clause is important and needs to be present in
		// any Segment.Run method in some form, but with at least the
		// following two statements.
		close(segment.Out)
		wg.Done()
	}()
	for msg := range segment.In {
		// Work with the flow messages here.
		segment.Out <- msg
	}
}

// Every Segment needs an init() function of some form in its file to be
// callable from config. An unregistered Segment will only be available using
// the API.
func init() {
	segment := &NoOp{}
	RegisterSegment("noop", segment)
}
