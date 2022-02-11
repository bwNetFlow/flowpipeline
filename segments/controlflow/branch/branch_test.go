package branch

import (
	"log"
	"sync"
	"testing"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

// Branch Segment test, passthrough test
// This does not work currently, as segment tests are scoped for segment
// package only, and this specific segment requires some pipeline
// initialization, which would lead to an import cycle. Thus, this test
// confirms that it fails silently, and this segment is instead tested from the
// pipeline package test files.
func TestSegment_Branch_passthrough(t *testing.T) {
	segment := segments.LookupSegment("branch").New(map[string]string{}).(*Branch)
	if segment == nil {
		log.Fatal("[error] Configured segment 'branch' could not be initialized properly, see previous messages.")
	}
	in, out := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire(in, out)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	// this would timeout if it worked properly, instead it logs an error and returns
	segment.Run(wg)
}
