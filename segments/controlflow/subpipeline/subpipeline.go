package subpipeline

import (
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
)

// This mirrors the proper implementation in the pipeline package. This
// duplication is to avoid the import cycle.
type Pipeline interface {
	Start()
	Close()
	GetInput() chan *flow.FlowMessage
	GetOutput() <-chan *flow.FlowMessage
}

type SubPipeline struct {
	segments.BaseSegment
	subpipeline Pipeline
}

func (segment SubPipeline) New(config map[string]string) segments.Segment {
	return &SubPipeline{}
}

func (segment *SubPipeline) ImportPipeline(pipeline interface{}) {
	segment.subpipeline = pipeline.(Pipeline)
}

func (segment *SubPipeline) Run(wg *sync.WaitGroup) {
	defer func() {
		segment.subpipeline.Close()
		close(segment.Out)
		wg.Done()
	}()

	go segment.subpipeline.Start()

	for pre := range segment.In {
		segment.subpipeline.GetInput() <- pre
		post := <-segment.subpipeline.GetOutput()
		segment.Out <- post
	}
}

// Override default Rewire method for this segment.
func (segment *SubPipeline) Rewire(chans []chan *flow.FlowMessage, in uint, out uint) {
	segment.In = chans[in]
	segment.Out = chans[out]
}

func init() {
	segment := &SubPipeline{}
	segments.RegisterSegment("subpipeline", segment)
}
