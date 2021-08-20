package segments

import (
	"log"
	"os"
	"testing"

	flow "github.com/bwNetFlow/protobuf/go"
	"github.com/hashicorp/logutils"
)

func TestMain(m *testing.M) {
	log.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"info", "warning", "error"},
		MinLevel: logutils.LogLevel("warning"),
		Writer:   os.Stderr,
	})
	code := m.Run()
	os.Exit(code)
}

// NoOp Segment test, passthrough test
func TestSegment_NoOp_passthrough(t *testing.T) {
	result := TestSegment("noop", map[string]string{},
		&flow.FlowMessage{Type: 3})
	if result.Type != 3 {
		t.Error("Segment NoOp is not working.")
	}
}

// DropFields Segment tests are thorough and try every combination
func TestSegment_DropFields_policyKeep(t *testing.T) {
	result := TestSegment("dropfields", map[string]string{"policy": "keep", "fields": "DstAddr"},
		&flow.FlowMessage{SrcAddr: []byte{192, 168, 88, 142}, DstAddr: []byte{192, 168, 88, 143}},
	)
	if len(result.SrcAddr) != 0 || len(result.DstAddr) == 0 {
		t.Error("Segment DropFields is not keeping the proper fields.")
	}
}

func TestSegment_DropFields_policyDrop(t *testing.T) {
	result := TestSegment("dropfields", map[string]string{"policy": "drop", "fields": "SrcAddr"},
		&flow.FlowMessage{SrcAddr: []byte{192, 168, 88, 142}, DstAddr: []byte{192, 168, 88, 143}},
	)
	if len(result.SrcAddr) != 0 || len(result.DstAddr) == 0 {
		t.Error("Segment DropFields is not dropping the proper fields.")
	}
}

// AddCid Segment tests are thorough and try every combination
func TestSegment_AddCid_noLocalAddrKeep(t *testing.T) {
	result := TestSegment("addcid", map[string]string{"filename": "../examples/enricher/customer_subnets.csv"},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{192, 168, 88, 142}})
	if result.Cid != 0 {
		t.Error("Segment AddCid is adding a Cid when the local address is undetermined.")
	}
}

func TestSegment_AddCid_noLocalAddrDrop(t *testing.T) {
	result := TestSegment("addcid", map[string]string{"filename": "../examples/enricher/customer_subnets.csv", "dropunmatched": "false"},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{192, 168, 88, 142}})
	if result != nil {
		t.Error("Segment AddCid is not dropping the flow as instructed if the local address is undetermined.")
	}
}

func TestSegment_AddCid_localAddrIsDst(t *testing.T) {
	result := TestSegment("addcid", map[string]string{"filename": "../examples/enricher/customer_subnets.csv"},
		&flow.FlowMessage{RemoteAddr: 1, DstAddr: []byte{192, 168, 88, 42}})
	if result.Cid != 1 {
		t.Error("Segment AddCid is not adding a Cid when the local address is the destination address.")
	}
}

func TestSegment_AddCid_localAddrIsSrc(t *testing.T) {
	result := TestSegment("addcid", map[string]string{"filename": "../examples/enricher/customer_subnets.csv"},
		&flow.FlowMessage{RemoteAddr: 2, SrcAddr: []byte{192, 168, 88, 142}})
	if result.Cid != 1 {
		t.Error("Segment AddCid is not adding a Cid when the local address is the source address.")
	}
}

// GeoLocation Segment tests are thorough and try every combination
func TestSegment_GeoLocation_noRemoteAddrKeep(t *testing.T) {
	result := TestSegment("geolocation", map[string]string{"filename": "../examples/enricher/GeoLite2-Country-Test.mmdb"},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{2, 125, 160, 218}, DstAddr: []byte{2, 125, 160, 218}})
	if result.RemoteCountry != "" {
		t.Error("Segment GeoLocation is adding a RemoteCountry when the remote address is undetermined.")
	}
}

func TestSegment_GeoLocation_noRemoteAddrDrop(t *testing.T) {
	result := TestSegment("geolocation", map[string]string{"filename": "../examples/enricher/GeoLite2-Country-Test.mmdb", "dropunmatched": "1"},
		&flow.FlowMessage{RemoteAddr: 0, SrcAddr: []byte{2, 125, 160, 218}, DstAddr: []byte{2, 125, 160, 218}})
	if result != nil {
		t.Error("Segment GeoLocation is not dropping the flow as instructed if the remote address is undetermined.")
	}
}

func TestSegment_GeoLocation_remoteAddrIsSrc(t *testing.T) {
	result := TestSegment("geolocation", map[string]string{"fileName": "../examples/enricher/GeoLite2-Country-Test.mmdb"},
		&flow.FlowMessage{RemoteAddr: 1, SrcAddr: []byte{2, 125, 160, 218}})
	if result.RemoteCountry != "GB" {
		t.Error("Segment GeoLocation is not adding RemoteCountry when the remote address is the source address.")
	}
}

func TestSegment_GeoLocation_remoteAddrIsDst(t *testing.T) {
	result := TestSegment("geolocation", map[string]string{"fileName": "../examples/enricher/GeoLite2-Country-Test.mmdb"},
		&flow.FlowMessage{RemoteAddr: 2, DstAddr: []byte{2, 125, 160, 218}})
	if result == nil || result.RemoteCountry != "GB" {
		t.Error("Segment GeoLocation is not adding RemoteCountry when the remote address is the destination address.")
	}
}

// RemoteAddress Segment testing is basically checking whether switch/case is working okay...
func TestSegment_RemoteAddress(t *testing.T) {
	result := TestSegment("remoteaddress", map[string]string{"flowSrc": "border"},
		&flow.FlowMessage{FlowDirection: 0})
	if result.RemoteAddr != 1 {
		t.Error("Segment RemoteAddress is not determining RemoteAddr correctly.")
	}
}

// FlowFilter Segment testing is basic, the filtering itself is tested in the flowfilter repo
func TestSegment_FlowFilter_accept(t *testing.T) {
	result := TestSegment("flowfilter", map[string]string{"filter": "proto 4"},
		&flow.FlowMessage{Proto: 4})
	if result == nil {
		t.Error("Segment FlowFilter dropped a flow incorrectly.")
	}
}

func TestSegment_FlowFilter_deny(t *testing.T) {
	result := TestSegment("flowfilter", map[string]string{"filter": "proto 5"},
		&flow.FlowMessage{Proto: 4})
	if result != nil {
		t.Error("Segment FlowFilter accepted a flow incorrectly.")
	}
}

// Goflow Segment test, passthrough test only, functionality is tested by Goflow package
func TestSegment_Goflow_passthrough(t *testing.T) {
	result := TestSegment("goflow", map[string]string{"port": "2055"},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment Goflow is not passing through flows.")
	}
}

// PrintDots Segment test, passthrough test only
func TestSegment_PrintDots_passthrough(t *testing.T) {
	result := TestSegment("printdots", map[string]string{"flowsPerDot": "100"},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment PrintDots is not passing through flows.")
	}
}

// PrintFlowdump Segment test, passthrough test only
func TestSegment_PrintFlowdump_passthrough(t *testing.T) {
	result := TestSegment("printflowdump", map[string]string{},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment PrintFlowDump is not passing through flows.")
	}
}

// Count Segment test, passthrough test only
func TestSegment_Count_passthrough(t *testing.T) {
	result := TestSegment("count", map[string]string{"prefix": "Test: "},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment Count is not passing through flows.")
	}
}

// StdIn Segment test, passthrough test only
func TestSegment_StdIn_passthrough(t *testing.T) {
	result := TestSegment("stdin", map[string]string{},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment StdIn is not passing through flows.")
	}
}

// StdOut Segment test, passthrough test only
func TestSegment_StdOut_passthrough(t *testing.T) {
	result := TestSegment("stdout", map[string]string{},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment StdOut is not passing through flows.")
	}
}

// PrometheusExporter Segment test, passthrough test only
func TestSegment_PrometheusExporter_passthrough(t *testing.T) {
	result := TestSegment("stdout", map[string]string{},
		&flow.FlowMessage{})
	if result == nil {
		t.Error("Segment PrometheusExporter is not passing through flows.")
	}
}

// Normalize Segment test, in-flow SampleingRate test
func TestSegment_Normalize_inFlowSamplingRate(t *testing.T) {
	result := TestSegment("normalize", map[string]string{},
		&flow.FlowMessage{SamplingRate: 32, Bytes: 1})
	if result.Bytes != 32 {
		t.Error("Segment Normalize is not working with in-flow SamplingRate.")
	}
}

// Normalize Segment test, fallback SampleingRate test
func TestSegment_Normalize_fallbackSamplingRate(t *testing.T) {
	result := TestSegment("normalize", map[string]string{"fallback": "42"},
		&flow.FlowMessage{SamplingRate: 0, Bytes: 1})
	if result.Bytes != 42 {
		t.Error("Segment Normalize is not working with fallback SamplingRate.")
	}
}

// Normalize Segment test, no fallback SampleingRate test
func TestSegment_Normalize_noFallbackSamplingRate(t *testing.T) {
	result := TestSegment("normalize", map[string]string{},
		&flow.FlowMessage{SamplingRate: 0, Bytes: 1})
	if result.Bytes != 1 {
		t.Error("Segment Normalize is not working with fallback SamplingRate.")
	}
}

// SNMPInterface Segment test
// TODO: find a way to run this elsewhere, as this currently only works by
// having the local 161/udp port forwarded to some router.
// func TestSegment_SNMPInterface(t *testing.T) {
// 	result := testSegmentWithFlows(
// 		&SNMPInterface{
// 			Community: "public",
// 			Regex:     "^_[a-z]{3}_[0-9]{5}_[0-9]{5}_ [A-Z0-9]+ (.*?) *( \\(.*)?$",
// 			ConnLimit: 1,
// 		},
// 		[]*flow.FlowMessage{
// 			&flow.FlowMessage{Type: 42, SamplerAddress: []byte{127, 0, 0, 1}, InIf: 70},
// 			&flow.FlowMessage{SamplerAddress: []byte{127, 0, 0, 1}, InIf: 70},
// 		})
// 	if result.SrcIfDesc == "" {
// 		t.Error("Segment SNMPInterface is not adding a SrcIfDesc.")
// 	}
// }
