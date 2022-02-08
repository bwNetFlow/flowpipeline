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
	GetDrop() <-chan *flow.FlowMessage
}

type SubPipeline struct {
	segments.BaseSegment
	subpipeline Pipeline
	conditional Pipeline
}

func (segment SubPipeline) New(config map[string]string) segments.Segment {
	return &SubPipeline{}
}

func (segment *SubPipeline) ImportSubpipeline(pipeline interface{}) {
	segment.subpipeline = pipeline.(Pipeline)
}

func (segment *SubPipeline) ImportConditionalPipeline(pipeline interface{}) {
	segment.conditional = pipeline.(Pipeline)
}

func (segment *SubPipeline) Run(wg *sync.WaitGroup) {
	defer func() {
		segment.subpipeline.Close()
		segment.conditional.Close()
		close(segment.Out)
		wg.Done()
	}()

	go segment.conditional.Start()
	go segment.subpipeline.Start()

	for pre := range segment.In {
		segment.conditional.GetInput() <- pre
		select {
		case msg := <-segment.conditional.GetDrop():
			segment.Out <- msg
		case msg := <-segment.conditional.GetOutput():
			// Using the message from the conditional's output will
			// not revert any changes made to the flow. This can't
			// really be counteracted as the conditional is
			// processed async, so we can't use `pre` and assume
			// that it was the original flow corresponding to the
			// `msg` returned here...
			segment.subpipeline.GetInput() <- msg
		}
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
