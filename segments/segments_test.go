package segments

import (
	"log"
	"os"
	"sync"
	"testing"

	flow "github.com/bwNetFlow/protobuf/go"
	"github.com/hashicorp/logutils"
)

func TestMain(m *testing.M) {
	log.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"info", "warning", "error"},
		MinLevel: logutils.LogLevel("warn"),
		Writer:   os.Stderr,
	})
	code := m.Run()
	os.Exit(code)
}

func testSegmentWithFlow(s Segment, f *flow.FlowMessage) (result *flow.FlowMessage) {
	in, out := make(chan *flow.FlowMessage), make(chan *flow.FlowMessage)
	s.Rewire(in, out)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go s.Run(wg)

	in <- f
	close(in)

	result = <-out
	wg.Wait()

	return
}

// NoOp Segment test, passthrough test
func TestSegment_NoOp_passthrough(t *testing.T) {
	result := testSegmentWithFlow(&NoOp{}, &flow.FlowMessage{Type: 3})
	if result.Type != 3 {
		t.Error("Segment NoOp is not working.")
	}
}

// DropFields Segment tests are thorough and try every combination
func TestSegment_DropFields_policyKeep(t *testing.T) {
	result := testSegmentWithFlow(&DropFields{Policy: "keep", Keep: "DstAddr"},
		&flow.FlowMessage{SrcAddr: []byte{192, 168, 88, 142}, DstAddr: []byte{192, 168, 88, 143}})
	if len(result.SrcAddr) != 0 || len(result.DstAddr) == 0 {
		t.Error("Segment DropFields is not keeping the proper fields.")
	}
}

func TestSegment_DropFields_policyDrop(t *testing.T) {
	result := testSegmentWithFlow(&DropFields{Policy: "drop", Drop: "SrcAddr"},
		&flow.FlowMessage{SrcAddr: []byte{192, 168, 88, 142}, DstAddr: []byte{192, 168, 88, 143}})
	if len(result.SrcAddr) != 0 || len(result.DstAddr) == 0 {
		t.Error("Segment DropFields is not dropping the proper fields.")
	}
}

// AddCid Segment tests are thorough and try every combination
func TestSegment_AddCid_noLocalAddrKeep(t *testing.T) {
	result := testSegmentWithFlow(&AddCid{FileName: "../example_subnets.csv"},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{192, 168, 88, 142}})
	if result.Cid != 0 {
		t.Error("Segment AddCid is adding a Cid when the local address is undetermined.")
	}
}

func TestSegment_AddCid_noLocalAddrDrop(t *testing.T) {
	result := testSegmentWithFlow(&AddCid{FileName: "../example_subnets.csv", DropUnmatched: true},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{192, 168, 88, 142}})
	if result != nil {
		t.Error("Segment AddCid is not dropping the flow as instructed if the local address is undetermined.")
	}
}

func TestSegment_AddCid_localAddrIsDst(t *testing.T) {
	result := testSegmentWithFlow(&AddCid{FileName: "../example_subnets.csv"},
		&flow.FlowMessage{RemoteAddr: 1, DstAddr: []byte{192, 168, 88, 42}})
	if result.Cid != 1 {
		t.Error("Segment AddCid is not adding a Cid when the local address is the destination address.")
	}
}

func TestSegment_AddCid_localAddrIsSrc(t *testing.T) {
	result := testSegmentWithFlow(&AddCid{FileName: "../example_subnets.csv"},
		&flow.FlowMessage{RemoteAddr: 2, SrcAddr: []byte{192, 168, 88, 142}})
	if result.Cid != 1 {
		t.Error("Segment AddCid is not adding a Cid when the local address is the source address.")
	}
}

// GeoLocation Segment tests are thorough and try every combination
func TestSegment_GeoLocation_noRemoteAddrKeep(t *testing.T) {
	result := testSegmentWithFlow(&GeoLocation{FileName: "../GeoLite2-Country-Test.mmdb"},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{2, 125, 160, 218}, DstAddr: []byte{2, 125, 160, 218}})
	if result.RemoteCountry != "" {
		t.Error("Segment GeoLocation is adding a RemoteCountry when the remote address is undetermined.")
	}
}

func TestSegment_GeoLocation_noRemoteAddrDrop(t *testing.T) {
	result := testSegmentWithFlow(&GeoLocation{
		FileName: "../GeoLite2-Country-Test.mmdb", DropUnmatched: true},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{2, 125, 160, 218}, DstAddr: []byte{2, 125, 160, 218}})
	if result != nil {
		t.Error("Segment GeoLocation is not dropping the flow as instructed if the remote address is undetermined.")
	}
}

func TestSegment_GeoLocation_remoteAddrIsSrc(t *testing.T) {
	result := testSegmentWithFlow(&GeoLocation{FileName: "../GeoLite2-Country-Test.mmdb"},
		&flow.FlowMessage{RemoteAddr: 1, SrcAddr: []byte{2, 125, 160, 218}})
	if result.RemoteCountry != "GB" {
		t.Error("Segment GeoLocation is not adding RemoteCountry when the remote address is the source address.")
	}
}

func TestSegment_GeoLocation_remoteAddrIsDst(t *testing.T) {
	result := testSegmentWithFlow(&GeoLocation{FileName: "../GeoLite2-Country-Test.mmdb"},
		&flow.FlowMessage{RemoteAddr: 2, DstAddr: []byte{2, 125, 160, 218}})
	if result.RemoteCountry != "GB" {
		t.Error("Segment GeoLocation is not adding RemoteCountry when the remote address is the destination address.")
	}
}

// RemoteAddress Segment testing is basically checking whether switch/case is working okay...
func TestSegment_RemoteAddress(t *testing.T) {
	result := testSegmentWithFlow(&RemoteAddress{FlowSrc: "border"}, &flow.FlowMessage{FlowDirection: 0})
	if result.RemoteAddr != 1 {
		t.Error("Segment RemoteAddress is not determining RemoteAddr correctly.")
	}
}

// FlowFilter Segment testing is basic, the filtering itself is tested in the flowfilter repo
func TestSegment_FlowFilter_accept(t *testing.T) {
	result := testSegmentWithFlow(&FlowFilter{Filter: "proto 4"}, &flow.FlowMessage{Proto: 4})
	if result == nil {
		t.Error("Segment FlowFilter dropped a flow incorrectly.")
	}
}

func TestSegment_FlowFilter_deny(t *testing.T) {
	result := testSegmentWithFlow(&FlowFilter{Filter: "proto 5"}, &flow.FlowMessage{Proto: 4})
	if result != nil {
		t.Error("Segment FlowFilter accepted a flow incorrectly.")
	}
}

// Goflow Segment test, passthrough test only, functionality is tested by Goflow package
func TestSegment_Goflow_passthrough(t *testing.T) {
	result := testSegmentWithFlow(&Goflow{Port: 2055}, &flow.FlowMessage{})
	if result == nil {
		t.Error("Segment Goflow is not passing through flows.")
	}
}

// KafkaConsumer Segment test
func TestSegment_KafkaConsumer_todo(t *testing.T) {
	// TODO: figure out how to test this or mock up Kafka
	return
}

// KafkaProducer Segment test
func TestSegment_KafkaProducer_todo(t *testing.T) {
	// TODO: figure out how to test this or mock up Kafka
	return
}

// PrintDots Segment test, passthrough test only
func TestSegment_PrintDots_passthrough(t *testing.T) {
	result := testSegmentWithFlow(&PrintDots{FlowsPerDot: 100}, &flow.FlowMessage{})
	if result == nil {
		t.Error("Segment PrintDots is not passing through flows.")
	}
}

// PrintFlowdump Segment test, passthrough test only
func TestSegment_PrintFlowdump_passthrough(t *testing.T) {
	result := testSegmentWithFlow(&PrintFlowdump{}, &flow.FlowMessage{})
	if result == nil {
		t.Error("Segment PrintFlowDump is not passing through flows.")
	}
}

// Count Segment test, passthrough test only
func TestSegment_Count_passthrough(t *testing.T) {
	result := testSegmentWithFlow(&Count{Prefix: "Test: "}, &flow.FlowMessage{})
	if result == nil {
		t.Error("Segment Count is not passing through flows.")
	}
}

// StdIn Segment test, passthrough test only
func TestSegment_StdIn_passthrough(t *testing.T) {
	result := testSegmentWithFlow(&StdIn{}, &flow.FlowMessage{})
	if result == nil {
		t.Error("Segment StdIn is not passing through flows.")
	}
}

// StdOut Segment test, passthrough test only
func TestSegment_StdOut_passthrough(t *testing.T) {
	result := testSegmentWithFlow(&StdOut{}, &flow.FlowMessage{})
	if result == nil {
		t.Error("Segment StdOut is not passing through flows.")
	}
}

// Prometheus Exporter test, passthrough test only
func TestSegment_Prometheus_exporter(t *testing.T) {
	result := testSegmentWithFlow(&StdOut{}, &flow.FlowMessage{})
	if result == nil {
		t.Error("Segment PrometheusExporter is not passing through flows.")
	}

}
