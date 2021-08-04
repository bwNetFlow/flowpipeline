package pipeline

import (
	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
	"log"
	"sync"
)

type Pipeline struct {
	In          chan *flow.FlowMessage
	Out         chan *flow.FlowMessage
	wg          *sync.WaitGroup
	SegmentList []segments.Segment `yaml: segments`
}

func (pipeline Pipeline) AutoDrain() {
	go func() {
		for _ = range pipeline.Out {
		}
		log.Println("Pipeline closed.")
	}()
}

func (pipeline Pipeline) Close() {
	defer func() {
		recover() // in case In is already closed
		pipeline.wg.Wait()
	}()
	close(pipeline.In)
}

func NewPipeline(segmentList ...segments.Segment) *Pipeline {
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
