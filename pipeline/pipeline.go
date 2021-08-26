// The pipeline package manages segments in Pipeline objects.
package pipeline

import (
	"log"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

// Basically a list of segments. It further exposes the In and Out channels of
// the Pipeline as a whole, i.e. the ingress channel of the first and the
// egress channel of the last segment in its SegmentList.
type Pipeline struct {
	In          chan *flow.FlowMessage
	Out         <-chan *flow.FlowMessage
	wg          *sync.WaitGroup
	SegmentList []segments.Segment `yaml: segments`
}

// Starts up a goroutine specific to this Pipeline which reads any message from
// the Out channel and discards it. This is a convenience function to enable
// having a segment at the end of the pipeline handle all results, i.e. having
// no post-pipeline processing.
func (pipeline Pipeline) AutoDrain() {
	go func() {
		for _ = range pipeline.Out {
		}
		log.Println("[info] Pipeline closed, auto draining finished.")
	}()
}

// Closes down a Pipeline by closing its In channel and waiting for all
// segments to propagate this close event through the full pipeline,
// terminating all segment goroutines and thus releasing the waitgroup.
// Blocking.
func (pipeline Pipeline) Close() {
	defer func() {
		recover() // in case In is already closed
		pipeline.wg.Wait()
	}()
	close(pipeline.In)
}

// Initializes a new Pipeline object and then starts all segment goroutines
// therein. Initialization includes creating any intermediate channels and
// wiring up the segments in the segmentList with them.
func New(segmentList ...segments.Segment) *Pipeline {
	channels := make([]chan *flow.FlowMessage, len(segmentList)+1)
	channels[0] = make(chan *flow.FlowMessage)
	wg := sync.WaitGroup{}
	for i, segment := range segmentList {
		channels[i+1] = make(chan *flow.FlowMessage)
		segment.Rewire(channels[i], channels[i+1])
		wg.Add(1)
		go segment.Run(&wg)
	}
	return &Pipeline{In: channels[0], Out: channels[len(channels)-1], wg: &wg, SegmentList: segmentList}
}
