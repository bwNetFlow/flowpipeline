// The pipeline package manages segments in Pipeline objects.
package pipeline

import (
	"log"
	"sync"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
	"github.com/bwNetFlow/flowpipeline/segments/filter/drop"
	"github.com/bwNetFlow/flowpipeline/segments/filter/elephant"
	"github.com/bwNetFlow/flowpipeline/segments/filter/flowfilter"
	"github.com/bwNetFlow/flowpipeline/segments/pass"
)

// Basically a list of segments. It further exposes the In and Out channels of
// the Pipeline as a whole, i.e. the ingress channel of the first and the
// egress channel of the last segment in its SegmentList.
type Pipeline struct {
	In          chan *pb.FlowContainer
	Out         <-chan *pb.FlowContainer
	Drop        chan *pb.FlowContainer
	wg          *sync.WaitGroup
	SegmentList []segments.Segment
}

func (pipeline *Pipeline) GetInput() chan *pb.FlowContainer {
	return pipeline.In
}

func (pipeline *Pipeline) GetOutput() <-chan *pb.FlowContainer {
	return pipeline.Out
}

func (pipeline *Pipeline) GetDrop() <-chan *pb.FlowContainer {
	if pipeline.Drop != nil {
		return pipeline.Drop
	}
	pipeline.Drop = make(chan *pb.FlowContainer)
	// Subscribe to drops from special segments, namely all based on
	// BaseFilterSegment grouped in the filter directory.
	for _, segment := range pipeline.SegmentList {
		switch typedSegment := segment.(type) {
		case *drop.Drop:
			typedSegment.SubscribeDrops(pipeline.Drop)
		case *elephant.Elephant:
			typedSegment.SubscribeDrops(pipeline.Drop)
		case *flowfilter.FlowFilter:
			typedSegment.SubscribeDrops(pipeline.Drop)
		}
	}
	// If there are no filter/* segments, this channel will never have
	// messages available.
	return pipeline.Drop
}

// Starts up a goroutine specific to this Pipeline which reads any message from
// the Out channel and discards it. This is a convenience function to enable
// having a segment at the end of the pipeline handle all results, i.e. having
// no post-pipeline processing.
func (pipeline *Pipeline) AutoDrain() {
	go func() {
		for fc := range pipeline.Out {
			fc.FlowSpan.End()
		}
		log.Println("[info] Pipeline closed, auto draining finished.")
	}()
}

// Closes down a Pipeline by closing its In channel and waiting for all
// segments to propagate this close event through the full pipeline,
// terminating all segment goroutines and thus releasing the waitgroup.
// Blocking.
func (pipeline *Pipeline) Close() {
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
	if len(segmentList) == 0 {
		segmentList = []segments.Segment{&pass.Pass{}}
	}
	channels := make([]chan *pb.FlowContainer, len(segmentList)+1)
	channels[0] = make(chan *pb.FlowContainer)
	for i, segment := range segmentList {
		channels[i+1] = make(chan *pb.FlowContainer)
		segment.Rewire(channels[i], channels[i+1])
	}
	return &Pipeline{In: channels[0], Out: channels[len(channels)-1], wg: &sync.WaitGroup{}, SegmentList: segmentList}
}

// Starts the Pipeline by starting all segment goroutines therein.
func (pipeline *Pipeline) Start() {
	for _, segment := range pipeline.SegmentList {
		pipeline.wg.Add(1)
		go segment.Run(pipeline.wg)
	}
}
