package pipeline

import (
	"testing"

	"github.com/bwNetFlow/flowpipeline/segments"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/dropfields"
	"github.com/bwNetFlow/flowpipeline/segments/pass"
	flow "github.com/bwNetFlow/protobuf/go"
)

func TestPipelineBuild(t *testing.T) {
	segmentList := []segments.Segment{&pass.Pass{}, &pass.Pass{}}
	pipeline := New(segmentList...)
	pipeline.Start()
	pipeline.In <- &flow.FlowMessage{Type: 3}
	fmsg := <-pipeline.Out
	if fmsg.Type != 3 {
		t.Error("Pipeline Setup is not working.")
	}
}

func TestPipelineTeardown(t *testing.T) {
	segmentList := []segments.Segment{&pass.Pass{}, &pass.Pass{}}
	pipeline := New(segmentList...)
	pipeline.Start()
	pipeline.AutoDrain()
	pipeline.In <- &flow.FlowMessage{Type: 3}
	pipeline.Close() // fail test on halting ;)
}

func TestPipelineConfigSuccess(t *testing.T) {
	pipeline := NewFromConfig([]byte(`---
- segment: pass
  config:
    foo: $baz
    bar: $0`))
	pipeline.Start()
	pipeline.In <- &flow.FlowMessage{Type: 3}
	fmsg := <-pipeline.Out
	if fmsg.Type != 3 {
		t.Error("Pipeline built from config is not working.")
	}
}

func Test_Branch_passthrough(t *testing.T) {
	pipeline := NewFromConfig([]byte(`---
- segment: branch
  if:
  - segment: flowfilter
    config:
      filter: proto tcp
  then:
  - segment: dropfields
    config:
      policy: drop
      fields: InIf
  else:
  - segment: dropfields
    config:
      policy: drop
      fields: OutIf
`))
	pipeline.Start()
	pipeline.In <- &flow.FlowMessage{Proto: 6, InIf: 1, OutIf: 1}
	fmsg := <-pipeline.Out
	if fmsg.Proto != 6 || fmsg.InIf == 1 || fmsg.OutIf != 1 {
		t.Errorf("Branch segment did not work correctly, state is Proto %d, InIf %d, OutIf %d, should be (6, 0, 1).", fmsg.Proto, fmsg.InIf, fmsg.OutIf)
	}
	pipeline.In <- &flow.FlowMessage{Proto: 42, InIf: 1, OutIf: 1}
	fmsg = <-pipeline.Out
	if fmsg.Proto != 42 || fmsg.InIf != 1 || fmsg.OutIf == 1 {
		t.Errorf("Branch segment did not work correctly, state is Proto %d, InIf %d, OutIf %d, should be (42, 1, 0).", fmsg.Proto, fmsg.InIf, fmsg.OutIf)
	}
}
