package pipeline

import (
	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
	"testing"
)

func TestPipelineBuild(t *testing.T) {
	segmentList := []segments.Segment{&segments.NoOp{}, &segments.NoOp{}}
	pipeline := NewPipeline(segmentList...)
	pipeline.In <- &flow.FlowMessage{Type: 3}
	fmsg := <-pipeline.Out
	if fmsg.Type != 3 {
		t.Error("Pipeline Setup is not working.")
	}
}

func TestPipelineTeardown(t *testing.T) {
	segmentList := []segments.Segment{&segments.NoOp{}, &segments.NoOp{}}
	pipeline := NewPipeline(segmentList...)
	pipeline.AutoDrain()
	pipeline.In <- &flow.FlowMessage{Type: 3}
	pipeline.Close() // TODO: fail test on halting ;)
}

func TestPipelineConfigSuccess(t *testing.T) {
	segmentList := SegmentListFromConfig([]byte(`---
- segment: noop
  config:
    foo: $baz`))
	if len(segmentList) != 1 {
		t.Error("Config Parsing is not okay.")
	}
	pipeline := NewPipeline(segmentList...)
	pipeline.In <- &flow.FlowMessage{Type: 3}
	fmsg := <-pipeline.Out
	if fmsg.Type != 3 {
		t.Error("Pipeline built from config is not working.")
	}
}
