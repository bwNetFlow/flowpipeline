package addcid

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
	"github.com/hashicorp/logutils"
)

func TestMain(m *testing.M) {
	log.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"info", "warning", "error"},
		MinLevel: logutils.LogLevel("info"),
		Writer:   os.Stderr,
	})
	code := m.Run()
	os.Exit(code)
}

// AddCid Segment tests are thorough and try every combination
func TestSegment_AddCid_noLocalAddrKeep(t *testing.T) {
	result := segments.TestSegment("addcid", map[string]string{"filename": "../../../examples/enricher/customer_subnets.csv"},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{192, 168, 88, 142}})
	if result.Cid != 0 {
		t.Error("Segment AddCid is adding a Cid when the local address is undetermined.")
	}
}

func TestSegment_AddCid_noLocalAddrDrop(t *testing.T) {
	result := segments.TestSegment("addcid", map[string]string{"filename": "../../../examples/enricher/customer_subnets.csv", "dropunmatched": "true"},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{192, 168, 88, 142}})
	if result != nil {
		t.Error("Segment AddCid is not dropping the flow as instructed if the local address is undetermined.")
	}
}

func TestSegment_AddCid_localAddrIsDst(t *testing.T) {
	result := segments.TestSegment("addcid", map[string]string{"filename": "../../../examples/enricher/customer_subnets.csv"},
		&flow.FlowMessage{RemoteAddr: 1, DstAddr: []byte{192, 168, 88, 42}})
	if result.Cid != 1 {
		t.Error("Segment AddCid is not adding a Cid when the local address is the destination address.")
	}
}

func TestSegment_AddCid_localAddrIsSrc(t *testing.T) {
	result := segments.TestSegment("addcid", map[string]string{"filename": "../../../examples/enricher/customer_subnets.csv"},
		&flow.FlowMessage{RemoteAddr: 2, SrcAddr: []byte{192, 168, 88, 142}})
	if result.Cid != 1 {
		t.Error("Segment AddCid is not adding a Cid when the local address is the source address.")
	}
}

func TestSegment_AddCid_bothAddrs(t *testing.T) {
	result := segments.TestSegment("addcid", map[string]string{"matchboth": "1", "filename": "../../../examples/enricher/customer_subnets.csv"},
		&flow.FlowMessage{SrcAddr: []byte{192, 168, 88, 142}})
	if result.Cid != 1 {
		t.Error("Segment AddCid is not adding a Cid when the local address is the source address.")
	}
}

// AddCid Segment benchmark passthrough
func BenchmarkAddCid(b *testing.B) {
	log.SetOutput(ioutil.Discard)
	os.Stdout, _ = os.Open(os.DevNull)

	segment := AddCid{}.New(map[string]string{"filename": "../../../examples/enricher/customer_subnets.csv"})

	in, out := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	segment.Rewire([]chan *flow.FlowMessage{in, out}, 0, 1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go segment.Run(wg)

	for n := 0; n < b.N; n++ {
		in <- &flow.FlowMessage{SrcAddr: []byte{192, 168, 88, 142}}
		_ = <-out
	}
	close(in)
}
